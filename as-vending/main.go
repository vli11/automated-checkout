// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"os"

	"as-vending/functions"
	"as-vending/routes"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/transforms"
	utilities "github.com/intel-iot-devkit/automated-checkout-utilities"
)

const (
	serviceKey = "as-vending"
)

func main() {
	service, ok := pkg.NewAppService(serviceKey)
	if !ok {
		os.Exit(1)
	}

	lc := service.LoggingClient()

	// get the application settings from configuration.toml
	appSettings := service.ApplicationSettings()
	if appSettings == nil {
		lc.Error("No application settings found")
		os.Exit(1)
	}

	var vendingState functions.VendingState
	vendingState.Configuration = new(functions.ServiceConfiguration)
	vendingState.CommandClient = service.CommandClient()
	if vendingState.CommandClient == nil {
		lc.Error("Error command service missing from client's configuration")
		os.Exit(1)
	}

	// retrieve & parse the required application settings into a proper
	// configuration struct
	if err := utilities.MarshalSettings(appSettings, vendingState.Configuration, true); err != nil {
		lc.Errorf("Application settings could not be processed: %v", err.Error())
		os.Exit(1)
	}

	lc.Infof("Running the application functions for %s devices...", vendingState.Configuration.DeviceNames)

	// create stop channels for each of the wait threads
	stopChannel := make(chan int)
	doorOpenStopChannel := make(chan int)
	doorCloseStopChannel := make(chan int)
	inferenceStopChannel := make(chan int)

	// Set default values for vending state
	vendingState.CVWorkflowStarted = false
	vendingState.MaintenanceMode = false
	vendingState.CurrentUserData = functions.OutputData{}
	vendingState.DoorClosed = true
	// global stop channel for threads
	vendingState.ThreadStopChannel = stopChannel
	// open event thread
	vendingState.DoorOpenedDuringCVWorkflow = false
	vendingState.DoorOpenWaitThreadStopChannel = doorOpenStopChannel
	// close event thread
	vendingState.DoorClosedDuringCVWorkflow = false
	vendingState.DoorCloseWaitThreadStopChannel = doorCloseStopChannel
	// inference thread
	vendingState.InferenceDataReceived = false
	vendingState.InferenceWaitThreadStopChannel = inferenceStopChannel

	var err error

	controller := routes.NewController(lc, service, &vendingState)
	err = controller.AddAllRoutes()
	if err != nil {
		lc.Errorf("failed to add all Routes: %s", err.Error())
		os.Exit(1)
	}

	// create the function pipeline to run when an event is read on the device channels
	err = service.SetFunctionsPipeline(
		transforms.NewFilterFor(vendingState.Configuration.DeviceNames).FilterByDeviceName,
		vendingState.DeviceHelper,
	)
	if err != nil {
		lc.Errorf("SDK initialization failed: %s", err.Error())
		os.Exit(1)
	}

	// tell the SDK to "start" and begin listening for events to trigger the pipeline.
	err = service.MakeItRun()
	if err != nil {
		lc.Errorf("MakeItRun returned error: %s", err.Error())
		os.Exit(1)
	}

	// do any required cleanup here

	os.Exit(0)
}
