package controller

import (
	"github.com/stevenchiu30801/free5gc-operator/pkg/controller/free5gcslice"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, free5gcslice.Add)
}
