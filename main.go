package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hubapi/pkg/hubclient"
	"github.com/wostzone/hubapi/pkg/hubconfig"
	"github.com/wostzone/owserver/internal"
)

// Main entry to WoST protocol adapter for owserver-v2
// This setup the configuration from file and commandline parameters and launches the service
func main() {
	svc := internal.NewOWServerPB()
	hubConfig, err := hubconfig.LoadCommandlineConfig("", internal.PluginID, &svc.Config)

	err = svc.Start(hubConfig)
	if err != nil {
		logrus.Errorf("Logger: Failed to start")
		os.Exit(1)
	}
	hubclient.WaitForSignal()
	svc.Stop()
	os.Exit(0)
}
