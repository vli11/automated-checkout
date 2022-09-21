// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package routes

import (
	"as-vending/functions"
	"fmt"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

type Controller struct {
	lc           logger.LoggingClient
	service      interfaces.ApplicationService
	vendingState *functions.VendingState
}

func NewController(lc logger.LoggingClient, service interfaces.ApplicationService, vendingState *functions.VendingState) Controller {
	return Controller{
		lc:           lc,
		service:      service,
		vendingState: vendingState,
	}
}

func (c *Controller) AddAllRoutes() error {
	err := c.service.AddRoute("/boardStatus", c.vendingState.BoardStatus, "POST")
	if errWithMsg := errorAddRouteHandler(err); errWithMsg != nil {
		return errWithMsg
	}

	err = c.service.AddRoute("/resetDoorLock", c.vendingState.ResetDoorLock, "POST")
	if errWithMsg := errorAddRouteHandler(err); errWithMsg != nil {
		return errWithMsg
	}

	err = c.service.AddRoute("/maintenanceMode", c.vendingState.GetMaintenanceMode, "GET", "OPTIONS")
	if errWithMsg := errorAddRouteHandler(err); errWithMsg != nil {
		return errWithMsg
	}
	return nil
}

func errorAddRouteHandler(err error) error {
	if err != nil {
		return fmt.Errorf("error adding route: %s", err.Error())
	}
	return nil
}
