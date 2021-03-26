package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hubapi/pkg/hubconfig"
	"github.com/wostzone/hubapi/pkg/plugin"
	"github.com/wostzone/owserver/internal"
)

var pluginConfig = &internal.PluginConfig{}

// Main entry to WoST protocol adapter for owserver-v2
// This setup the configuration from file and commandline parameters and launches the service
func main() {
	hubConfig, err := hubconfig.LoadCommandlineConfig("", internal.PluginID, pluginConfig)

	svc := internal.OWServerPB{}
	err = svc.Start(hubConfig, pluginConfig)
	if err != nil {
		logrus.Errorf("Logger: Failed to start")
		os.Exit(1)
	}
	plugin.WaitForSignal()
	svc.Stop()
	os.Exit(0)
}
