const { events, Job, Group } = require("brigadier")

const projectOrg = "deis"
const projectName = "duffle"

const goImg = "golang:1.11"
const gopath = "/go"
const localPath = gopath + `/src/github.com/${projectOrg}/${projectName}`;

const noop = {run: () => {return Promise.resolve()}}

function build(e, project) {
  // Create a new job to run Go tests
  var build = new Job(`${projectName}-build`, goImg);

  // Set a few environment variables.
  build.env = {
      "DEST_PATH": localPath,
      "GOPATH": gopath
  };

  // Run Go unit tests
  build.tasks = [
    // Need to move the source into GOPATH so vendor/ works as desired.
    `mkdir -p ${localPath}`,
    `cp -a /src/* ${localPath}`,
    `cp -a /src/.git ${localPath}`,
    `cd ${localPath}`,
    "make bootstrap",
    "make dep-validate",
    "make lint",
    "make test"
  ];

  return build
}

// Here we can add additional Check Runs, which will run in parallel and
// report their results independently to GitHub
function runSuite(e, p) {
  // For now, this is the one-stop shop running build, lint and test targets
  runTests(e, p).catch(e => {console.error(e.toString())});
}

// runTests is a Check Run that is ran as part of a Checks Suite
function runTests(e, p) {
  console.log("Check requested")

    // Create Notification object (which is just a Job to update GH using the Checks API)
    var note = new Notification(`tests`, e, p);
    note.conclusion = "";
    note.title = "Run Tests";
    note.summary = "Running the test targets for " + e.revision.commit;
    note.text = "This test will ensure build, linting and tests all pass."

    // Send notification, then run, then send pass/fail notification
    return notificationWrap(build(e, p), note)
}

// A GitHub Check Suite notification
class Notification {
  constructor(name, e, p) {
      this.proj = p;
      this.payload = e.payload;
      this.name = name;
      this.externalID = e.buildID;
      // TODO: add Kashti link when available
      // this.detailsURL = `https://<kashti domain>/kashti/builds/${ e.buildID }`;
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
      this.count++
      var j = new Job(`${ this.name }-${ this.count }`, "technosophos/brigade-github-check-run:latest");
      j.env = {
          CHECK_CONCLUSION: this.conclusion,
          CHECK_NAME: this.name,
          CHECK_TITLE: this.title,
          CHECK_PAYLOAD: this.payload,
          CHECK_SUMMARY: this.summary,
          CHECK_TEXT: this.text,
          // TODO: add when applicable
          // CHECK_DETAILS_URL: this.detailsURL,
          CHECK_EXTERNAL_ID: this.externalID
      }
      return j.run();
  }
}

// Helper to wrap a job execution between two notifications.
async function notificationWrap(job, note, conclusion) {
  if (conclusion == null) {
      conclusion = "success"
  }
  await note.run();
  try {
      let res = await job.run()
      const logs = await job.logs();

      note.conclusion = conclusion;
      note.summary = `Task "${ job.name }" passed`;
      note.text = note.text = "```" + res.toString() + "```\nTest Complete";
      return await note.run();
  } catch (e) {
      const logs = await job.logs();
      note.conclusion = "failure";
      note.summary = `Task "${ job.name }" failed for ${ e.buildID }`;
      note.text = "```" + logs + "```\nFailed with error: " + e.toString();
      try {
          return await note.run();
      } catch (e2) {
          console.error("failed to send notification: " + e2.toString());
          console.error("original error: " + e.toString());
          return e2;
      }
  }
}

function release(project, tag) {
  if (!project.secrets.ghToken) {
    throw new Error(`Project ${projectName} must have 'secrets.ghToken' set`)
  }

  // Cross-compile binaries for a given release and upload them to GitHub.
  var release = new Job(`${projectName}-release`, goImg)

  parts = project.repo.name.split("/", 2)

  release.env = {
    GITHUB_USER: parts[0],
    GITHUB_REPO: parts[1],
    GITHUB_TOKEN: project.secrets.ghToken,
    GOPATH: gopath
  }

  release.tasks = [
    "go get github.com/aktau/github-release",
    `cd /src`,
    `git checkout ${tag}`,
    // Need to move the source into GOPATH so vendor/ works as desired.
    `mkdir -p ${localPath}`,
    `cp -a /src/* ${localPath}`,
    `cp -a /src/.git ${localPath}`,
    `cd ${localPath}`,
    "make bootstrap",
    "make build-release",
    `last_tag=$(git describe --tags ${tag}^ --abbrev=0 --always)`,
    `github-release release \
      -t ${tag} \
      -n "${parts[1]} ${tag}" \
      -d "$(git log --no-merges --pretty=format:'- %s %H (%aN)' HEAD ^$last_tag)" \
      || echo "release ${tag} exists"`,
    "for bin in ./bin/*; do github-release upload -f ${bin} -n $(basename ${bin}) -t " + tag + "; done"
  ];

  console.log(`release at https://github.com/${project.repo.name}/releases/tag/${tag}`);

  return Group.runEach([
    release,
    slackNotify("Duffle Release", `${tag} release now on GitHub! <https://github.com/${project.repo.name}/releases/tag/${tag}>`, project)
  ])
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
    return noop
  }
}
 
events.on("exec", (e, p) => {
  return build(e, p).run()
})

events.on("check_suite:requested", runSuite)
events.on("check_suite:rerequested", runSuite)
events.on("check_run:rerequested", runSuite)

events.on("release", (e, p) => {
  /*
   * Expects JSON of the form {'tag': 'v1.2.3'}
   */
  payload = JSON.parse(e.payload)
  if (!payload.tag) {
    throw error("No tag specified")
  }

  release(p, payload.tag)
})
