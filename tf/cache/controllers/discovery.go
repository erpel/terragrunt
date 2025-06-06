package controllers

import (
	"net/http"

	"maps"

	"github.com/gruntwork-io/terragrunt/tf/cache/router"
	"github.com/labstack/echo/v4"
)

const (
	discoveryPath = "/.well-known"
)

type Endpointer interface {
	// Endpoints returns controller endpoints.
	Endpoints() map[string]any
}

type DiscoveryController struct {
	*router.Router

	Endpointers []Endpointer
}

// Register implements router.Controller.Register
func (controller *DiscoveryController) Register(router *router.Router) {
	controller.Router = router.Group(discoveryPath)

	// Discovery Process
	// https://developer.hashicorp.com/terraform/internals/remote-service-discovery#discovery-process
	controller.GET("/terraform.json", controller.terraformAction)
}

func (controller *DiscoveryController) terraformAction(ctx echo.Context) error {
	endpoints := make(map[string]any)

	for _, endpointer := range controller.Endpointers {
		maps.Copy(endpoints, endpointer.Endpoints())
	}

	return ctx.JSON(http.StatusOK, endpoints)
}
