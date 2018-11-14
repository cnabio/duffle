package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gosuri/uitable"
	"github.com/renstrom/fuzzysearch/fuzzy"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/deis/duffle/pkg/bundle"
	"github.com/deis/duffle/pkg/duffle/home"
	"github.com/deis/duffle/pkg/repo/remote"
	"github.com/deis/duffle/pkg/repo/remote/auth"
)

type RepoBundle struct {
	bundle.Bundle
	Repository string
}

// BundleList is a list of bundle references.
// Implements a sorter on Name.
type BundleList []*RepoBundle

// Len returns the length.
func (bl BundleList) Len() int { return len(bl) }

// Swap swaps the position of two items in the versions slice.
func (bl BundleList) Swap(i, j int) { bl[i], bl[j] = bl[j], bl[i] }

// Less returns true if the version of entry a is less than the version of entry b.
func (bl BundleList) Less(a, b int) bool {
	return strings.Compare(bl[a].Name, bl[b].Name) < 1
}

func newSearchCmd(w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "perform a fuzzy search on available bundles",
		RunE: func(cmd *cobra.Command, args []string) error {
			found, err := search(args)
			if err != nil {
				return err
			}
			sort.Sort(found)
			table := uitable.New()
			table.AddRow("NAME", "VERSION")
			for _, bundle := range found {
				table.AddRow(path.Join(bundle.Repository, bundle.Name), bundle.Version)
			}
			fmt.Fprintln(w, table)
			return nil
		},
	}

	return cmd
}

func search(keywords []string) (BundleList, error) {
	foundBundles := BundleList{}
	h := home.Home(homePath())

	loginCreds, err := auth.Load(filepath.Join(h.String(), "auth.json"))
	if err != nil {
		return nil, err
	}

	if len(loginCreds) == 0 {
		return nil, fmt.Errorf("You have not logged into any repository yet. Try `duffle login %s`", home.DefaultRepository())
	}

	for entry, creds := range loginCreds {
		url := &url.URL{
			Scheme: "https",
			Host:   entry,
			Path:   remote.IndexPath,
		}

		log.Debugf("Searching %s...", url.String())
		// if no keywords are given, display all available bundles
		if len(keywords) == 0 {
			repoBundles, err := searchRepo(url, creds)
			if err != nil {
				return nil, err
			}
			foundBundles = append(foundBundles, repoBundles...)
		}
		for _, keyword := range keywords {
			req, err := http.NewRequest("GET", url.String(), nil)
			if err != nil {
				return nil, err
			}
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", creds.Token))

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusUnauthorized {
				return nil, fmt.Errorf("token for %s expired. Please run `duffle login %s` again to fetch a new auth token", url.Hostname(), url.Hostname())
			} else if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("request to %s responded with a non-200 status code: %d", url.String(), resp.StatusCode)
			}

			index, err := remote.LoadIndexReader(resp.Body)
			if err != nil {
				return nil, err
			}

			var found = make(map[string]bool)
			names := make([]string, 0, len(index.Entries))
			for name := range index.Entries {
				names = append(names, name)
			}
			for _, foundName := range fuzzy.Find(keyword, names) {
				found[foundName] = true
			}
			// also check if the latest version of each bundle has a matching keyword
			for _, name := range names {
				for _, bundleKeyword := range index.Entries[name][0].Keywords {
					if bundleKeyword == keyword {
						found[name] = true
					}
				}
			}
			for n := range found {
				for name := range index.Entries {
					if n == name {
						foundBundles = append(foundBundles, &RepoBundle{*index.Entries[name][0], entry})
					}
				}
			}
		}
	}

	return foundBundles, nil
}

func searchRepo(url *url.URL, creds auth.RepositoryCredentials) (BundleList, error) {
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", creds.Token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("token for %s expired. Please run `duffle login %s` again to fetch a new auth token", url.Hostname(), url.Hostname())
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request to %s responded with a non-200 status code: %d", url.String(), resp.StatusCode)
	}

	index, err := remote.LoadIndexReader(resp.Body)
	if err != nil {
		return nil, err
	}

	bundles := make(BundleList, 0, len(index.Entries))
	for _, entry := range index.Entries {
		bundles = append(bundles, &RepoBundle{*entry[0], url.Hostname()})
	}
	return bundles, nil
}
