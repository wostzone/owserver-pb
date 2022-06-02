package main

import (
	"os"

	"github.com/wostzone/wost-go/pkg/config"
	"github.com/wostzone/wost-go/pkg/logging"
	"github.com/wostzone/wost-go/pkg/proc"

	"github.com/sirupsen/logrus"

	"github.com/wostzone/owserver/internal"
)

// Main entry to WoST protocol adapter for owserver-v2
// This setup the configuration from file and commandline parameters and launches the service
func main() {
	serviceConfig := internal.OWServerPBConfig{}
	hubConfig, err := config.LoadAllConfig(os.Args, "", internal.PluginID, &serviceConfig)
	logging.SetLogging(hubConfig.Loglevel, hubConfig.LogFile)
	if err != nil {
		logrus.Errorf("%s: Failed to configure: %s", internal.PluginID, err)
		os.Exit(1)
	}

	svc := internal.NewOWServerPB(serviceConfig,
		hubConfig.Address, hubConfig.MqttPortCert, hubConfig.CaCert, hubConfig.PluginCert)

	err = svc.Start()
	if err != nil {
		logrus.Errorf("%s: Failed to start: %s", internal.PluginID, err)
		os.Exit(1)
	}
	proc.WaitForSignal()
	svc.Stop()
	os.Exit(0)
}
