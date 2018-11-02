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
    "make lint",
    "make test"
  ];

  start = ghNotify("pending", `Build started as ${ e.buildID }`, e, project)

  // Run tests in parallel. Then if it's a release, push binaries.
  // Then send GitHub a notification on the status.
  Group.runAll([start, build])
    .then(() => {
        return ghNotify("success", `Build ${ e.buildID } passed`, e, project).run()
    })
    .then(() => {
      var runRelease = false
      if (e.type == "push" && e.revision.ref.startsWith("refs/tags/")) {
        // Run the release in the background.
        runRelease = true
        let parts = e.revision.ref.split("/", 3)
        let tag = parts[2]
        return release(project, tag)
      }
      return Promise.resolve(runRelease)
    })
    .catch(err => {
      return ghNotify("failure", `failed build ${ e.buildID }`, e, project).run()
    });
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

function ghNotify(state, msg, e, project) {
  if (e.revision.commit) {
    const gh = new Job(`${projectName}-notify-${state}`, "technosophos/github-notify:latest")

    gh.env = {
      GH_REPO: project.repo.name,
      GH_STATE: state,
      GH_DESCRIPTION: msg,
      GH_CONTEXT: projectName,
      GH_TOKEN: project.secrets.ghToken,
      GH_COMMIT: e.revision.commit
    }

    return gh
  } else {
    console.log(`Warning: GitHub notification not sent for state '${state}' as no commit was found.`)
    return noop
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
    return noop
  }
}
 
events.on("exec", build)
events.on("push", build)
events.on("pull_request", build)

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