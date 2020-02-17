package main

import (
	"io"

	"github.com/cnabio/cnab-go/action"
	"github.com/cnabio/cnab-go/driver"
)

func setOut(w io.Writer) action.OperationConfigFunc {
	return func(op *driver.Operation) error {
		op.Out = w
		return nil
	}
}
