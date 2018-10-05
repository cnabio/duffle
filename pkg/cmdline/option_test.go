package cmdline

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOption(t *testing.T) {
	is := assert.New(t)
	buf := bytes.NewBuffer(nil)
	opt := &options{}
	WithStderr(buf)(opt)
	WithStdout(buf)(opt)
	WithBuildID("buildid")(opt)
	opt.stderr.Write([]byte("te"))
	opt.stdout.Write([]byte("st"))
	is.Equal(buf.String(), "test")
	is.Equal(opt.buildID, "buildid")
}
