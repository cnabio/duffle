package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/deis/duffle/pkg/claim"
	"github.com/deis/duffle/pkg/duffle/home"
	"github.com/deis/duffle/pkg/utils/crud"
)

var (
	// duffleHome depicts the home directory where all duffle config is stored.
	duffleHome string
	rootCmd    *cobra.Command
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

func claimStorage() claim.Store {
	h := home.Home(homePath())
	return claim.NewClaimStore(crud.NewFileSystemStore(h.Claims(), "json"))
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
	rootCmd = newRootCmd(os.Stdout)
	must(rootCmd.Execute())
}
