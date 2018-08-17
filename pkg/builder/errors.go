package builder

import "errors"

var (
	// ErrDockerfileNotExist is returned when no Dockerfile exists during "kubed up."
	ErrDockerfileNotExist = errors.New("Dockerfile does not exist. Please create it using 'kubed create' before calling 'kubed up'")
)
