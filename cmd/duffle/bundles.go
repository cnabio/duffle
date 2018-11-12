package main

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"

	"github.com/deis/duffle/pkg/duffle/home"
	"github.com/deis/duffle/pkg/repo"
)

// NamedRepositoryList is a list of bundle references.
// Implements a sorter on Name.
type NamedRepositoryList []*NamedRepository

// Len returns the length.
func (bl NamedRepositoryList) Len() int { return len(bl) }

// Swap swaps the position of two items in the versions slice.
func (bl NamedRepositoryList) Swap(i, j int) { bl[i], bl[j] = bl[j], bl[i] }

// Less returns true if the version of entry a is less than the version of entry b.
func (bl NamedRepositoryList) Less(a, b int) bool {
	return strings.Compare(bl[a].Name(), bl[b].Name()) < 1
}

// NamedRepository is a reference to a repository with a name, tag and digest.
type NamedRepository struct {
	name   string
	tag    string
	digest string
}

// Name returns the full name.
func (n *NamedRepository) String() string {
	return n.name + ":" + n.tag
}

// Name returns the name.
func (n *NamedRepository) Name() string {
	return n.name
}

// Tag returns the tag.
func (n *NamedRepository) Tag() string {
	return n.tag
}

// Digest returns the digest.
func (n *NamedRepository) Digest() string {
	return n.digest
}

func newBundlesCmd(w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bundles",
		Short: "lists bundles pulled or built and stored locally",
		RunE: func(cmd *cobra.Command, args []string) error {
			home := home.Home(homePath())
			references, err := searchLocal(home)
			if err != nil {
				return err
			}
			sort.Sort(references)
			table := uitable.New()
			table.AddRow("NAME", "VERSION", "DIGEST")
			for _, ref := range references {
				table.AddRow(ref.Name(), ref.Tag(), ref.Digest())
			}
			fmt.Fprintln(w, table)
			return nil
		},
	}

	return cmd
}

func searchLocal(home home.Home) (NamedRepositoryList, error) {
	references := NamedRepositoryList{}

	index, err := repo.LoadIndex(home.Repositories())
	if err != nil {
		return nil, fmt.Errorf("cannot open %s: %v", home.Repositories(), err)
	}

	for repo, tagList := range index {
		for tag, digest := range tagList {
			references = append(references, &NamedRepository{
				repo,
				tag,
				digest,
			})
		}
	}

	return references, nil
}
