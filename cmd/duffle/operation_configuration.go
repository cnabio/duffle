package main

import (
	"io"

	"github.com/deislabs/cnab-go/action"
	"github.com/deislabs/cnab-go/driver"
)

func setOut(w io.Writer) action.OperationConfigFunc {
	return func(op *driver.Operation) error {
		op.Out = w
		return nil
	}
}
