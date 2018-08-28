package repo

import "errors"

var (
	// ErrExists indicates that a repository already exists
	ErrExists = errors.New("repository already exists")
	// ErrDoesNotExist indicates that a repository does not exist
	ErrDoesNotExist = errors.New("repository does not exist")
	// ErrNotARepo indicates that a repository exists, but is malformed
	ErrNotARepo = errors.New("repository is not a valid duffle repository (missing a `bundles/` directory in the root)")
	// ErrHomeMissing indicates that the directory expected to contain repositories does not exist
	ErrHomeMissing = errors.New(`repository home "$(duffle home)/repositories" does not exist`)
	// ErrMissingSource indicates that information about the source of the repository was not found
	ErrMissingSource = errors.New("cannot get information about the source of this repository")
	// ErrRepoDirty indicates that the repository was modified
	ErrRepoDirty = errors.New("repository is in a dirty git tree state so we cannot update. Try removing and adding this repo back")
	// ErrVersionDoesNotExist indicates that the requested version does not exist
	ErrVersionDoesNotExist = errors.New("requested version does not exist")
)
