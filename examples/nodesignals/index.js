process.stdin.resume();
// Don't run uninstall, but exit.
process.on('SIGINT', exitQuietly);
// Run uninstall and stop
process.on('SIGTERM', destroy);
// Don't do anything admin-like, but don't stop running.
process.on('SIGTSTP', pause);
// SIGUSR1 is used by node
// Run an install operation
process.on('SIGUSR2', install);
// Run an upgrade.
process.on('SIGHUP', upgrade);
// SIGPIPE is not used... so if we need another...

// This is the entrypoint.
var timer = watchLogs();

function exitQuietly(){
    clearInterval(timer);
    console.log("Exiting by interrupt")
    process.exit(0);
}

function destroy(){
    clearInterval(timer);
    console.log("Running uninstall");
    process.exit(0);
}

function pause(){
    clearInterval(timer);
    timer = setInterval(() => {
        console.log("paused");
    }, 3000);
    console.log("Pausing");
}

function install(){
    console.log("install");
}
function upgrade(){
    console.log("upgrade");
}

function watchLogs(){
    return setInterval(() => {
        console.log("I'm still standing.")
    }, 3000);
}