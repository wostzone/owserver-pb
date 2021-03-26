package internal

import (
	"github.com/sirupsen/logrus"

	"github.com/wostzone/hubapi/api"
	"github.com/wostzone/hubapi/pkg/hubclient"
	"github.com/wostzone/hubapi/pkg/hubconfig"
)

// PluginID is the default ID of the WoST Logger plugin
const PluginID = "owserver"

// PluginConfig with owserver plugin configuration
type PluginConfig struct {
	EdsAddress string `yaml:"owserverAddress"`
	LoginName  string `yaml:"loginName"`
	Password   string `yaml:"password"`
}

// ThingInfo contains the last published TD and property values
type ThingInfo struct {
	thingTD        api.ThingTD
	propertyValues map[string]string
}

// OWServerPB is a  hub protocol binding plugin for capturing 1-wire OWServer V2 Data
type OWServerPB struct {
	config      PluginConfig         // options for accessing EDS OWServer
	edsAPI      *EdsAPI              // EDS device access
	hubConfig   *hubconfig.HubConfig // hub based configuration
	hubClient   api.IHubClient
	gatewayInfo ThingInfo            // TD of owserver gateway
	nodeInfo    map[string]ThingInfo // map of node thing info objects by thing ID
}

// Start connects to the hub internal message bus and starts polling
// the owserver.
func (pb *OWServerPB) Start(hubConfig *hubconfig.HubConfig, pluginConfig *PluginConfig) error {
	var err error
	pb.config = *pluginConfig
	pb.hubConfig = hubConfig
	pb.nodeInfo = make(map[string]ThingInfo)
	pb.edsAPI = NewEdsAPI(pluginConfig.EdsAddress, pluginConfig.LoginName, pluginConfig.Password)
	pb.hubClient = hubclient.NewPluginClient(PluginID, hubConfig)
	pb.hubClient.Start(false)

	logrus.Infof("Service OWServer startup completed")
	return err
}

// Stop the service
func (pb *OWServerPB) Stop() {
	logrus.Info("Stopping service OWServer")
}
