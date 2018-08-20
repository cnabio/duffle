package driver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type CommandDriver struct {
	Name string
}

func (d *CommandDriver) Run(op *Operation) error {
	return d.exec(op)
}

func (d *CommandDriver) Handles(dt string) bool {
	out, err := exec.Command(d.cliName(), "--handles").CombinedOutput()
	if err != nil {
		fmt.Printf("%s --handles: %s", d.cliName(), err)
		return false
	}
	types := strings.Split(string(out), ",")
	for _, tt := range types {
		if dt == strings.TrimSpace(tt) {
			return true
		}
	}
	return false
}

func (d *CommandDriver) cliName() string {
	return "duffle-" + strings.ToLower(d.Name)
}

func (d *CommandDriver) exec(op *Operation) error {
	data, err := json.Marshal(op)
	if err != nil {
		return err
	}
	args := []string{}
	cmd := exec.Command(d.cliName(), args...)
	cmd.Dir, err = os.Getwd()
	if err != nil {
		return err
	}
	// NB: Since we don't set cmd.Env, we inherit the parent process's environment.
	// This means that the driver has access to everything in the environment, which I think
	// is a desirable feature. Option B: Have CommandDriver implement Configurable and get
	// strict.
	cmd.Stdin = bytes.NewBuffer(data)
	out, err := cmd.CombinedOutput()
	fmt.Println(string(out))
	return err
}
