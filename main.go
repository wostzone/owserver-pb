package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/gateway/pkg/config"
	"github.com/wostzone/gateway/pkg/gateway"
	"github.com/wostzone/owserver/internal"
)

var pluginConfig = &internal.PluginConfig{}

// Main entry to WoST protocol adapter for owserver-v2
func main() {
	gatewayConfig, err := config.SetupConfig("", internal.PluginID, pluginConfig)

	svc := internal.NewOWServer()
	err = svc.Start(gatewayConfig, pluginConfig)
	if err != nil {
		logrus.Errorf("Logger: Failed to start")
		os.Exit(1)
	}
	gateway.WaitForSignal()
	svc.Stop()
	os.Exit(0)
}
