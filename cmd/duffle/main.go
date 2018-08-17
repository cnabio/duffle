package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/deis/duffle/pkg/duffle/home"
)

var (
	// duffleHome depicts the home directory where all duffle config is stored.
	duffleHome string
)

func unimplemented(msg string) {
	panic(fmt.Errorf("unimplemented: %s", msg))
}

func homePath() string {
	return os.ExpandEnv(duffleHome)
}

func defaultDuffleHome() string {
	if home := os.Getenv(home.HomeEnvVar); home != "" {
		return home
	}

	homeEnvPath := os.Getenv("HOME")
	if homeEnvPath == "" && runtime.GOOS == "windows" {
		homeEnvPath = os.Getenv("USERPROFILE")
	}

	return filepath.Join(homeEnvPath, ".duffle")
}

func must(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "duffle: fatal: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				must(err)
			}
		}
	}()
	must(newRootCmd(os.Stdout).Execute())
}
