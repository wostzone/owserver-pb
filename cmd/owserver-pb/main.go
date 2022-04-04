package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/client/pkg/config"
	"github.com/wostzone/hub/lib/client/pkg/proc"
	"github.com/wostzone/owserver-pb/internal"
)

// Main entry to WoST protocol adapter for owserver-v2
// This setup the configuration from file and commandline parameters and launches the service
func main() {
	pluginConfig := internal.OWServerPBConfig{}
	hubConfig, err := config.LoadAllConfig(os.Args, "", internal.PluginID, &pluginConfig)
	if err != nil {
		logrus.Errorf("%s: Failed to configure: %s", internal.PluginID, err)
		os.Exit(1)
	}

	mqttHostPort := fmt.Sprintf("%s:%d", hubConfig.Address, hubConfig.MqttPortCert)
	svc := internal.NewOWServerPB(pluginConfig.ClientID,
		mqttHostPort, hubConfig.CaCert, hubConfig.PluginCert)

	err = svc.Start()
	if err != nil {
		logrus.Errorf("%s: Failed to start: %s", internal.PluginID, err)
		os.Exit(1)
	}
	proc.WaitForSignal()
	svc.Stop()
	os.Exit(0)
}
