package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/deis/duffle/pkg/duffle/home"
	"github.com/deis/duffle/pkg/loader"

	"github.com/gosuri/uitable"
	"github.com/renstrom/fuzzysearch/fuzzy"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newSearchCmd(w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "perform a fuzzy search on available bundles",
		Run: func(cmd *cobra.Command, args []string) {
			found := search(args)
			table := uitable.New()
			table.AddRow("NAME", "REPOSITORY", "VERSION")
			for _, bundle := range found {
				bundleFile, repo, err := getBundleFile(bundle)
				if err != nil {
					log.Debugf("bundle %s failed to load. skipping, %s", bundle, err)
					continue
				}

				l, err := loader.New(bundleFile)
				if err != nil {
					log.Debugf("failed to determine the correct loader for %s: %v", bundleFile, err)
					continue
				}
				bundle, err := l.Load()
				if err != nil {
					log.Debugf("failed to load %s: %v", bundleFile, err)
					log.Debugln(err)
					continue
				}
				table.AddRow(bundle.Name, repo, bundle.Version)
			}
			fmt.Fprintln(w, table)
		},
	}
	return cmd
}

func search(keywords []string) []string {
	var names = findNames()
	var found = make(map[string]bool)
	// if no keywords are given, display all available bundles
	if len(keywords) == 0 {
		for _, foundName := range names {
			found[foundName] = true
		}
	} else {
		for _, keyword := range keywords {
			for _, foundName := range fuzzy.Find(keyword, names) {
				found[foundName] = true
			}
		}
	}
	foundNames := []string{}
	for n := range found {
		foundNames = append(foundNames, n)
	}
	sort.Strings(foundNames)
	return foundNames
}

func findNames() []string {
	home := home.Home(homePath())
	repoPath := home.Repositories()
	var bundleNames []string
	filepath.Walk(repoPath, func(p string, f os.FileInfo, err error) error {
		if err != nil {
			log.Errorln(err)
			return nil
		}
		if isBundle(p, f) {
			bundleName := strings.TrimSuffix(f.Name(), ".json")
			repoName := strings.TrimPrefix(p, repoPath+string(os.PathSeparator))
			repoName = strings.TrimSuffix(repoName, string(os.PathSeparator)+filepath.Join("bundles", f.Name()))
			// for Windows clients, we need to replace the path separator with forward slashes
			repoName = strings.Replace(repoName, "\\", "/", -1)
			name := bundleName
			if repoName != home.DefaultRepository() {
				name = path.Join(repoName, bundleName)
			}
			bundleNames = append(bundleNames, name)
		}
		return nil
	})
	sort.Strings(bundleNames)
	return bundleNames
}

func isBundle(filePath string, f os.FileInfo) bool {
	return !f.IsDir() &&
		strings.HasSuffix(f.Name(), ".json") &&
		filepath.Base(filepath.Dir(filePath)) == "bundles"
}
