const { events, Job, Group } = require("brigadier");

const projectOrg = "cnabio";
const projectName = "duffle";

const builderImg = "docker:stable-dind"
const goImg = "golang:1.13";
const gopath = "/go"
const localPath = gopath + `/src/github.com/${projectOrg}/${projectName}`;

const releaseTagRegex = /^refs\/tags\/([0-9]+(?:\.[0-9]+)*(?:\-.+)?)$/;

// **********************************************
// Event Handlers
// **********************************************

events.on("exec", (e, p) => {
  return Group.runEach([
    test(),
    validateExamples()
  ]);
})

events.on("push", (e, p) => {
  let matchStr = e.revision.ref.match(releaseTagRegex);
  if (matchStr) {
    // This is an official release with a semantically versioned tag
    let matchTokens = Array.from(matchStr);
    let version = matchTokens[1];
    return buildAndPublishImage(p, version).run()
      .then(() => {
        githubRelease(p, version).run();
      });
  }
  if (e.revision.ref == "refs/heads/master") {
    // This runs tests then builds and publishes "edge" images
    return Group.runEach([
      test(),
      buildAndPublishImage(p, "")
    ]);
  }
})

events.on("check_suite:requested", runSuite);
events.on("check_suite:rerequested", runSuite);
events.on("check_run:rerequested", checkRequested);
events.on("issue_comment:created", handleIssueComment);
events.on("issue_comment:edited", handleIssueComment);

// **********************************************
// Actions
// **********************************************

sharedStorage = {
  enabled: true,
  path: "/duffle-binaries"
}

function test() {
  // Create a new job to run Go tests
  var job = new Job("tests", goImg);
  job.storage = sharedStorage;
  job.mountPath = localPath;
  // Set a few environment variables.
  job.env = {
    "GO111MODULE": "on",
    "SKIP_DOCKER": "true"
  };
  // Run Go unit tests
  job.tasks = [
    `cd ${localPath}`,
    "make build-all-bins test",
    `cp bin/* ${sharedStorage.path}`
  ];
  return job;
}

function validateExamples() {
  var job = new Job("validate-examples", builderImg);
  job.privileged = true;
  job.storage = sharedStorage;
  job.tasks = [
    "apk add --update --no-cache curl git make npm",
    "dockerd-entrypoint.sh &",
    "sleep 20",
    "cd /src",
    `install ${sharedStorage.path}/duffle-linux-amd64 /usr/local/bin/duffle`,
    "duffle init",
    "make validate"
  ];
  return job;
}

function buildAndPublishImage(project, version) {
  let dockerRegistry = project.secrets.dockerhubRegistry || "docker.io";
  let dockerOrg = project.secrets.dockerhubOrg || "deislabs";
  var job = new Job("build-and-publish-image", builderImg);
  job.privileged = true;
  job.tasks = [
    "apk add --update --no-cache make git",
    "dockerd-entrypoint.sh &",
    "sleep 20",
    "cd /src",
    `docker login ${dockerRegistry} -u ${project.secrets.dockerhubUsername} -p ${project.secrets.dockerhubPassword}`,
    `DOCKER_REGISTRY=${dockerRegistry} DOCKER_ORG=${dockerOrg} VERSION=${version} make build-image push-image`,
    `docker logout ${dockerRegistry}`
  ];
  return job;
}

function githubRelease(p, tag) {
  if (!p.secrets.ghToken) {
    throw new Error("Project must have 'secrets.ghToken' set");
  }
  // Cross-compile binaries for a given release and upload them to GitHub.
  var job = new Job("release", goImg);
  job.mountPath = localPath;
  parts = p.repo.name.split("/", 2);
  // Set a few environment variables.
  job.env = {
    "GO111MODULE": "on",
    "SKIP_DOCKER": "true",
    "GITHUB_USER": parts[0],
    "GITHUB_REPO": parts[1],
    "GITHUB_TOKEN": p.secrets.ghToken,
  };
  job.shell = "/bin/bash";
  job.tasks = [
    "go get github.com/aktau/github-release",
    `cd ${localPath}`,
    `VERSION=${tag} make build-all-bins`,
    `last_tag=$(git describe --tags ${tag}^ --abbrev=0 --always)`,
    `github-release release \
      -t ${tag} \
      -n "${parts[1]} ${tag}" \
      -d "$(git log --no-merges --pretty=format:'- %s %H (%aN)' HEAD ^$last_tag)" 2>&1 | sed -e "s/\${GITHUB_TOKEN}/<REDACTED>/"`,
    `for bin in ./bin/*; do \
      github-release upload -f $bin -n $(basename $bin) -t ${tag} 2>&1 | sed -e "s/\${GITHUB_TOKEN}/<REDACTED>/"; \
    done`
  ];
  console.log(job.tasks);
  console.log(`releases at https://github.com/${p.repo.name}/releases/tag/${tag}`);
  return job;
}

// handleIssueComment handles an issue_comment event, parsing the comment text
// and determining whether or not to trigger an action
function handleIssueComment(e, p) {
  payload = JSON.parse(e.payload);

  // Extract the comment body and trim whitespace
  comment = payload.body.comment.body.trim();

  // Here we determine if a comment should provoke an action
  switch (comment) {
    case "/brig run":
      return runSuite(e, p);
    default:
      console.log(`No applicable action found for comment: ${comment}`);
  }
}

// checkRequested is the default function invoked on a check_run:* event
//
// It determines which check is being requested (from the payload body)
// and runs this particular check, or else throws an error if the check
// is not found
function checkRequested(e, p) {
  payload = JSON.parse(e.payload);

  // Extract the check name
  name = payload.body.check_run.name;

  // Determine which check to run
  switch (name) {
    case "tests":
      return runTests(e, p, test);
    case "validateExamples":
      return runTests(e, p, validateExamples);
    default:
      throw new Error(`No check found with name: ${name}`);
  }
}

// Here we can add additional Check Runs, which will run in parallel and
// report their results independently to GitHub
function runSuite(e, p) {
  return Promise.all([
    runTests(e, p, test).catch((err) => { return err }),
    runTests(e, p, validateExamples).catch((err) => { return err }),
  ])
  .then((values) => {
    values.forEach((value) => {
      if (value instanceof Error) throw value;
    });
  })
}

// runTests is a Check Run that is run as part of a Checks Suite
function runTests(e, p, jobFunc) {
  console.log("Check requested");

  var check = new Check(e, p, jobFunc(),
    `https://brigadecore.github.io/kashti/builds/${e.buildID}`);
  return check.run();
}
