package internal

import (
	"fmt"
	"time"

	"github.com/wostzone/hubapi/api"
	"github.com/wostzone/hubapi/pkg/td"
)

const PropNameLatency = "latency"

// CreateOWGatewayTD creates a Thing Description document of the OWServer Gateway
func (pb *OWServerPB) CreateGatewayTD(gwParam map[string]string, latency time.Duration) (id string, thingTD api.ThingTD, newValues map[string]interface{}) {

	owsID := gwParam["DeviceName"]
	owsTD := td.CreateTD(owsID)
	newValues = make(map[string]interface{}, 0)

	for prop, val := range gwParam {
		propName := prop
		description := ""
		_ = val
		prop := td.CreateTDProperty(propName, description, td.PropertyTypeAttr)
		td.SetTDPropertyDataTypeString(prop, 0, 0)
		// Track changes to the value for publication
		oldValue := pb.gatewayInfo.propertyValues[propName]
		if oldValue != val {
			newValues[propName] = val
		}

		td.AddTDProperty(owsTD, propName, prop)
	}
	prop := td.CreateTDProperty(PropNameLatency, "OWServer Gateway connection latency", td.PropertyTypeState)
	td.SetTDPropertyDataTypeInteger(prop, 0, 0)
	td.AddTDProperty(owsTD, PropNameLatency, prop)
	newValues[PropNameLatency] = fmt.Sprint(latency)

	return owsID, owsTD, newValues
}

// AddOWNode adds a 1-wire node to the gateway TD
func (pb *OWServerPB) CreateNodeTD(nodeParam *XMLNode) (id string, thingTD api.ThingTD) {
	nodeID := nodeParam.Description
	nodeTD := td.CreateTD(nodeID)
	return nodeID, nodeTD
}
