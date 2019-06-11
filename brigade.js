const { events, Job, Group } = require("brigadier");

const projectOrg = "deislabs";
const projectName = "duffle";

const goImg = "quay.io/deis/lightweight-docker-go:v0.7.0";
const gopath = "/go"
const localPath = gopath + `/src/github.com/${projectOrg}/${projectName}`;

const releaseTagRegex = /^refs\/tags\/(v[0-9]+(?:\.[0-9]+)*(?:\-.+)?)$/;

const noopJob = { run: () => { return Promise.resolve() } }

// **********************************************
// Event Handlers
// **********************************************

events.on("exec", (e, p) => {
  return test().run();
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
      })
      .then(() => {
        slackNotify(
          "Duffle Release",
          `${version} release now on GitHub! <https://github.com/${p.repo.name}/releases/tag/${version}>`,
          p
        ).run();
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

function test() {
  // Create a new job to run Go tests
  var job = new Job("tests", goImg);
  job.mountPath = localPath;
  // Set a few environment variables.
  job.env = {
      "SKIP_DOCKER": "true"
  };
  // Run Go unit tests
  job.tasks = [
    `cd ${localPath}`,
    "make verify-vendored-code lint test"
  ];
  return job;
}

function buildAndPublishImage(project, version) {
  let dockerRegistry = project.secrets.dockerhubRegistry || "docker.io";
  let dockerOrg = project.secrets.dockerhubOrg || "deislabs";
  var job = new Job("build-and-publish-image", "docker:stable-dind");
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
    "SKIP_DOCKER": "true",
    "GITHUB_USER": parts[0],
    "GITHUB_REPO": parts[1],
    "GITHUB_TOKEN": p.secrets.ghToken,
  };
  job.tasks = [
    "go get github.com/aktau/github-release",
    `cd ${localPath}`,
    `VERSION=${tag} make build-all-bins`,
    `last_tag=$(git describe --tags ${tag}^ --abbrev=0 --always)`,
    `github-release release \
      -t ${tag} \
      -n "${parts[1]} ${tag}" \
      -d "$(git log --no-merges --pretty=format:'- %s %H (%aN)' HEAD ^$last_tag)" \
      || echo "release ${tag} exists"`,
    `for bin in ./bin/*; do github-release upload -f $bin -n $(basename $bin) -t ${tag}; done`
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
  switch(comment) {
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
  switch(name) {
    case "tests":
      return runTests(e, p);
    default:
      throw new Error(`No check found with name: ${name}`);
  }
}

// Here we can add additional Check Runs, which will run in parallel and
// report their results independently to GitHub
function runSuite(e, p) {
  // For the master branch, we build and publish images in response to the push
  // event. We test as a precondition for doing that, so we DON'T test here
  // for the master branch.
  if (e.revision.ref != "master") {
    // For now, this is the one-stop shop running build, lint and test targets
    return runTests(e, p);
  }
}

// runTests is a Check Run that is run as part of a Checks Suite
function runTests(e, p) {
  console.log("Check requested");

  // Create Notification object (which is just a Job to update GH using the Checks API)
  var note = new Notification(`tests`, e, p);
  note.conclusion = "";
  note.title = "Run Tests";
  note.summary = "Running the test targets for " + e.revision.commit;
  note.text = "This test will ensure build, linting and tests all pass.";

  // Send notification, then run, then send pass/fail notification
  return notificationWrap(test(), note);
}

// **********************************************
// Classes/Helpers
// **********************************************

// A GitHub Check Suite notification
class Notification {
  constructor(name, e, p) {
    this.proj = p;
    this.payload = e.payload;
    this.name = name;
    this.externalID = e.buildID;
    this.detailsURL = `https://brigadecore.github.io/kashti/builds/${ e.buildID }`;
    this.title = "running check";
    this.text = "";
    this.summary = "";

    // count allows us to send the notification multiple times, with a distinct pod name
    // each time.
    this.count = 0;

    // One of: "success", "failure", "neutral", "cancelled", or "timed_out".
    this.conclusion = "neutral";
  }

  // Send a new notification, and return a Promise<result>.
  run() {
    this.count++;
    // Here we are using the mutable 'edge' version of this utility
    // as an exercise of vetting the current master version of the code in this repo.
    // It is recommended that immutable tags be used in other cases,
    // e.g., a proper semver tag of 'vX.X.X' or the short git sha of a particular commit.
    var job = new Job(`${ this.name }-notification-${ this.count }`, "brigadecore/brigade-github-check-run:edge");
    job.imageForcePull = true;
    job.env = {
      "CHECK_CONCLUSION": this.conclusion,
      "CHECK_NAME": this.name,
      "CHECK_TITLE": this.title,
      "CHECK_PAYLOAD": this.payload,
      "CHECK_SUMMARY": this.summary,
      "CHECK_TEXT": this.text,
      "CHECK_DETAILS_URL": this.detailsURL,
      "CHECK_EXTERNAL_ID": this.externalID
    };
    return job.run();
  }
}

// Helper to wrap a job execution between two notifications.
async function notificationWrap(job, note) {
  await note.run();
  try {
    let res = await job.run();
    const logs = await job.logs();
    note.conclusion = "success";
    note.summary = `Task "${ job.name }" passed`;
    note.text = "```" + res.toString() + "```\nTest Complete";
    return await note.run();
  } catch (e) {
    const logs = await job.logs();
    note.conclusion = "failure";
    note.summary = `Task "${ job.name }" failed for ${ e.buildID }`;
    note.text = "```" + logs + "```\nFailed with error: " + e.toString();
    try {
      await note.run();
    } catch (e2) {
      console.error("failed to send notification: " + e2.toString());
      console.error("original error: " + e.toString());
    }
    throw e;
  }
}

function slackNotify(title, msg, project) {
  if (project.secrets.SLACK_WEBHOOK) {
    var slack = new Job(`${projectName}-slack-notify`, "technosophos/slack-notify:latest")

    slack.env = {
      SLACK_WEBHOOK: project.secrets.SLACK_WEBHOOK,
      SLACK_USERNAME: "duffle-ci",
      SLACK_TITLE: title,
      SLACK_MESSAGE: msg,
      SLACK_COLOR: "#00ff00"
    }
    slack.tasks = ["/slack-notify"]

    return slack
  } else {
    console.log(`Slack Notification for '${title}' not sent; no SLACK_WEBHOOK secret found.`)
    return noopJob
  }
}
