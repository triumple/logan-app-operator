package controller

import (
	"github.com/logancloud/logan-app-operator/pkg/controller/phpboot"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, phpboot.Add)
}
