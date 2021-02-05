const { events, Job, Group } = require("brigadier");
const { Check } = require("@brigadecore/brigade-utils");

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
  return Group.runAll([
    test(),
    testViaDocker()
  ])
  .then(() => {
    validateExamples().run();
  })
  .catch((err) => { return err });
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
  if (e.revision.ref == "refs/heads/main") {
    // This runs tests then builds and publishes "edge" images
    return Group.runEach([
      test(),
      testViaDocker(),
      buildAndPublishImage(p, "")
    ]);
  }
})

events.on("check_suite:requested", runSuite);
events.on("check_suite:rerequested", runSuite);
events.on("check_run:rerequested", checkRequested);
events.on("issue_comment:created", (e, p) => Check.handleIssueComment(e, p, runSuite));
events.on("issue_comment:edited", (e, p) => Check.handleIssueComment(e, p, runSuite));

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

// Standard Docker-In-Docker setup tasks that jobs below will use
dindSetupTasks = [
  "apk add --update --no-cache make git > /dev/null",
  "dockerd-entrypoint.sh > /dev/null 2>&1 &",
  // wait for docker daemon to set up
  "sleep 20"
];

function testViaDocker() {
  // Create a new job to run Go tests via Docker
  // to ensure Docker-based builds are functional
  var job = new Job("tests-via-docker", "docker:stable-dind");
  // Set privileged true to enable Docker-in-Docker
  job.privileged = true;
  // Set a few environment variables.
  job.env = {
    "DOCKER_INTERACTIVE": "false",
  };
  job.tasks = dindSetupTasks.concat([
    "cd /src",
    "make build-all-bins test"
  ]);
  return job;
}

function validateExamples() {
  var job = new Job("validate-examples", builderImg);
  job.privileged = true;
  job.storage = sharedStorage;
  job.tasks = dindSetupTasks.concat([
    "apk add --update --no-cache curl npm > /dev/null",
    "cd /src",
    `install ${sharedStorage.path}/duffle-linux-amd64 /usr/local/bin/duffle`,
    "duffle init",
    "make validate"
  ]);
  return job;
}

function buildAndPublishImage(project, version) {
  let dockerRegistry = project.secrets.dockerhubRegistry || "docker.io";
  let dockerOrg = project.secrets.dockerhubOrg || "deislabs";
  var job = new Job("build-and-publish-image", builderImg);
  job.privileged = true;
  job.tasks = dindSetupTasks.concat([
    "cd /src",
    `docker login ${dockerRegistry} -u ${project.secrets.dockerhubUsername} -p ${project.secrets.dockerhubPassword}`,
    `DOCKER_REGISTRY=${dockerRegistry} DOCKER_ORG=${dockerOrg} VERSION=${version} make build-image push-image`,
    `docker logout ${dockerRegistry}`
  ]);
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
      return runCheck(e, p, test);
    case "validate-examples":
      return runCheck(e, p, validateExamples);
    case "tests-via-docker":
      return runCheck(e, p, testViaDocker);
    default:
      throw new Error(`No check found with name: ${name}`);
  }
}

// Here we can add additional Check Runs, which will run in parallel and
// report their results independently to GitHub
function runSuite(e, p) {
  return Promise.all([
    // the test and validateExamples checks must run sequentially
    runCheck(e, p, test)
      .then(() => {
        runCheck(e, p, validateExamples)
      }).catch((err) => { return err }),
    runCheck(e, p, testViaDocker).catch((err) => { return err }),
  ])
  .then((values) => {
    values.forEach((value) => {
      if (value instanceof Error) throw value;
    });
  });
}

// runCheck is a Check Run that is run as part of a Checks Suite
function runCheck(e, p, jobFunc) {
  var check = new Check(e, p, jobFunc(),
    `https://brigadecore.github.io/kashti/builds/${e.buildID}`);
  return check.run();
}
