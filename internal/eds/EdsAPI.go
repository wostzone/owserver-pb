// Package eds with EDS OWServer API methods
package eds

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/wostlib-go/wostapi"
)

// family to device type. See also: http://owfs.sourceforge.net/simple_family.html
// Todo: get from config file so it is easy to update
var deviceTypeMap = map[string]wostapi.DeviceType{
	"10": wostapi.DeviceTypeThermometer,
	"28": wostapi.DeviceTypeThermometer,
	"7E": wostapi.DeviceTypeMultisensor,
}

// AttrVocab maps OWServer attribute names to IoT vocabulary
var AttrVocab = map[string]string{
	"MACAddress": wostapi.PropNameMAC,
	"DateTime":   wostapi.PropNameDateTime,
	"DeviceName": wostapi.PropNameName,
	"HostName":   wostapi.PropNameHostname,
}

// sensorTypeMap maps OWServer sensor names to IoT vocabulary
var SensorTypeVocab = map[string]struct {
	name     string
	dataType string
}{
	// "BarometricPressureHg": wostapi.PropNameAtmosphericPressure, // unit Hg
	"BarometricPressureMb": {name: wostapi.PropNameAtmosphericPressure,
		dataType: wostapi.WoTDataTypeNumber}, // unit Mb
	"DewPoint":    {name: wostapi.PropNameDewpoint, dataType: wostapi.WoTDataTypeNumber},
	"HeatIndex":   {name: wostapi.PropNameHeatIndex, dataType: wostapi.WoTDataTypeNumber},
	"Humidity":    {name: wostapi.PropNameHumidity, dataType: wostapi.WoTDataTypeNumber},
	"Humidex":     {name: wostapi.PropNameHumidex, dataType: wostapi.WoTDataTypeNumber},
	"Light":       {name: wostapi.PropNameLuminance, dataType: wostapi.WoTDataTypeNumber},
	"RelayState":  {name: wostapi.PropNameRelay, dataType: wostapi.WoTDataTypeBool},
	"Temperature": {name: wostapi.PropNameTemperature, dataType: wostapi.WoTDataTypeNumber},
}

// unitNameMap maps OWServer unit names to IoT vocabulary
var UnitNameVocab = map[string]string{
	"PercentRelativeHumidity": wostapi.UnitNamePercent,
	"Millibars":               wostapi.UnitNameMillibar,
	"Centigrade":              wostapi.UnitNameCelcius,
	"Fahrenheit":              wostapi.UnitNameFahrenheit,
	"InchesOfMercury":         wostapi.UnitNameMercury,
	"Lux":                     wostapi.UnitNameLux,
	"#":                       wostapi.UnitNameCount,
	"Volt":                    wostapi.UnitNameVolt,
}

// EdsAPI EDS device API properties and methods
type EdsAPI struct {
	address   string // EDS (IP) address or filename (file://./path/to/name.xml)
	loginName string // Basic Auth login name
	password  string // Basic Auth password
}

// XMLNode XML parsing node. Pure magic...
//--- https://stackoverflow.com/questions/30256729/how-to-traverse-through-xml-data-in-golang
type XMLNode struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:"-"`
	Content []byte     `xml:",innerxml"`
	Nodes   []XMLNode  `xml:",any"`
	// Possible attributes for subnodes, depending on the property name
	Description string `xml:"Description,attr"`
	Writable    string `xml:"Writable,attr"`
	Units       string `xml:"Units,attr"`
}

// OneWireNode with info on each node
type OneWireAttr struct {
	Name         string
	Unit         string
	Writable     bool
	Value        string
	PropertyType wostapi.ThingPropType
}
type OneWireNode struct {
	DeviceType wostapi.DeviceType
	// ThingID     string
	NodeID      string // hardware ID
	Name        string
	Description string
	Attr        map[string]OneWireAttr // attribute by name
}

// Apply the vocabulary to the name
// This returns the translated name from the vocabulary or the original name if not in the vocabulary
func applyVocabulary(name string, vocab map[string]string) (vocabName string, hasName bool) {
	vocabName, hasName = vocab[name]
	if !hasName {
		vocabName = name
	}
	return vocabName, hasName
}

// Parse the owserver xml data and returns a list of nodes and their parameters
//  xmlNode is the node to parse, its attribute and possibly subnodes
//  latency to add to the root node
//  isRootNode is set for the first node, eg the gateway itself
func (edsAPI *EdsAPI) ParseOneWireNodes(xmlNode *XMLNode, latency time.Duration, isRootNode bool) []*OneWireNode {
	owNodeList := []*OneWireNode{}

	owNode := OneWireNode{
		// ID:          xmlNode.Attrs["ROMId"],
		Name:        xmlNode.XMLName.Local,
		Description: xmlNode.Description,
		Attr:        make(map[string]OneWireAttr, 0),
		DeviceType:  wostapi.DeviceTypeGateway,
	}
	owNodeList = append(owNodeList, &owNode)
	// todo: find a better place for this
	if isRootNode {
		owAttr := OneWireAttr{
			Name:  wostapi.PropNameLatency,
			Value: fmt.Sprintf("%.3f", latency.Seconds()),
			Unit:  "sec",
		}
		owNode.Attr[owAttr.Name] = owAttr
	}
	// parse attributes
	for _, node := range xmlNode.Nodes {
		// if the xmlnode has no subnodes then it is a parameter describing the current node
		if len(node.Nodes) == 0 {
			// standardize the naming of properties and property types
			propType := wostapi.PropertyTypeAttr
			writable := (strings.ToLower(node.Writable) == "true")
			attrName := node.XMLName.Local
			sensorInfo, isSensor := SensorTypeVocab[attrName]
			if isSensor {
				attrName = sensorInfo.name
				propType = wostapi.PropertyTypeSensor
				if writable {
					propType = wostapi.PropertyTypeActuator
				}
			} else {
				propType = wostapi.PropertyTypeAttr
				if writable {
					propType = wostapi.PropertyTypeConfig
				}
				attrName, _ = applyVocabulary(attrName, AttrVocab)
			}

			unit, _ := applyVocabulary(node.Units, UnitNameVocab)
			owAttr := OneWireAttr{
				Name:         attrName,
				Value:        string(node.Content),
				Unit:         unit,
				PropertyType: propType,
				Writable:     writable,
			}
			owNode.Attr[owAttr.Name] = owAttr
			// Family is used to determine device type, default is gateway
			if node.XMLName.Local == "Family" {
				deviceType := deviceTypeMap[owAttr.Value]
				if deviceType == "" {
					deviceType = wostapi.DeviceTypeUnknown
				}
				owNode.DeviceType = deviceType
			} else if node.XMLName.Local == "ROMId" {
				// all subnodes use the ROMId as its ID
				owNode.NodeID = owAttr.Value
			} else if isRootNode && node.XMLName.Local == "DeviceName" {
				// The gateway itself uses the deviceName as its ID and name
				owNode.NodeID = owAttr.Value
				owNode.Name = owAttr.Value
				owNode.Description = "EDS OWServer Gateway"
			}

		} else {
			// The node contains subnodes which contain one or more sensors.
			subNodes := edsAPI.ParseOneWireNodes(&node, 0, false)
			owNodeList = append(owNodeList, subNodes...)
		}
	}
	// owNode.ThingID = td.CreatePublisherThingID(pb.hubConfig.Zone, PluginID, owNode.NodeID, owNode.DeviceType)

	return owNodeList
}

// ReadEds reads EDS hub and return the result as an XML node
// If edsAPI.address starts with file:// then read from file, otherwise from address
func (edsAPI *EdsAPI) ReadEds() (rootNode *XMLNode, err error) {
	if strings.HasPrefix(edsAPI.address, "file://") {
		filename := edsAPI.address[7:]
		buffer, err := ioutil.ReadFile(filename)
		if err != nil {
			logrus.Errorf("Unable to read EDS file from %s: %v", filename, err)
			return nil, err
		}
		err = xml.Unmarshal(buffer, &rootNode)
		return rootNode, err
	}
	// not a file, continue with http request
	edsURL := "http://" + edsAPI.address + "/details.xml"
	req, err := http.NewRequest("GET", edsURL, nil)
	req.SetBasicAuth(edsAPI.loginName, edsAPI.password)
	client := &http.Client{}
	resp, err := client.Do(req)

	// resp, err := http.Get(edsURL)
	if err != nil {
		logrus.Errorf("Unable to read EDS hub from %s: %v", edsURL, err)
		return nil, err
	}
	// Decode the EDS response into XML
	dec := xml.NewDecoder(resp.Body)
	err = dec.Decode(&rootNode)
	_ = resp.Body.Close()

	return rootNode, err
}

// UnmarshalXML parse xml
func (n *XMLNode) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	n.Attrs = start.Attr
	type node XMLNode

	return d.DecodeElement((*node)(n), &start)
}

// NewEdsAPI creates a new NewEdsAPI instance
func NewEdsAPI(address string, loginName string, password string) *EdsAPI {
	edsAPI := &EdsAPI{
		address:   address,
		loginName: loginName,
		password:  password,
	}
	return edsAPI
}
