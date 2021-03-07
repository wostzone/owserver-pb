package internal

import (
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/gateway/pkg/config"
	"github.com/wostzone/gateway/pkg/messaging"
)

// PluginID is the default ID of the WoST Logger plugin
const PluginID = "owserver"

// PluginConfig with owserver plugin configuration
type PluginConfig struct {
	EdsAddress string `yaml:"edsAddress"`
	LoginName  string `yaml:"loginName"`
	Password   string `yaml:"password"`
}

// WostOWServer is a  gateway protocol adapter plugin for capturing 1-wire OWServer V2 Data
type WostOWServer struct {
	config    PluginConfig          // options for accessing EDS OWServer
	edsAPI    *EdsAPI               // EDS device access
	gwConfig  *config.GatewayConfig // gateway based configuration
	messenger messaging.IGatewayMessenger
}

// Poll the OWServer gateway for device updates
func (svc *WostOWServer) Poll() error {
	return errors.New("Not implemented")
}

// Start connects to the gateway internal message bus and starts polling
// the owserver.
func (svc *WostOWServer) Start(gwConfig *config.GatewayConfig, pluginConfig *PluginConfig) error {
	var err error
	svc.config = *pluginConfig
	svc.gwConfig = gwConfig
	svc.edsAPI = NewEdsAPI(pluginConfig.EdsAddress, pluginConfig.LoginName, pluginConfig.Password)
	svc.messenger, err = messaging.StartGatewayMessenger(PluginID, gwConfig)

	logrus.Infof("Service OWServer startup completed")
	return err
}

// Stop the service
func (svc *WostOWServer) Stop() {
	logrus.Info("Stopping service OWServer")
}
