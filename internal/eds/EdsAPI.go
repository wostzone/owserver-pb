// Package eds with EDS OWServer API methods
package eds

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/client/pkg/vocab"
)

// family to device type. See also: http://owfs.sourceforge.net/simple_family.html
// Todo: get from config file so it is easy to update
var deviceTypeMap = map[string]vocab.DeviceType{
	"10": vocab.DeviceTypeThermometer,
	"28": vocab.DeviceTypeThermometer,
	"7E": vocab.DeviceTypeMultisensor,
}

// AttrVocab maps OWServer attribute names to IoT vocabulary
var AttrVocab = map[string]string{
	"MACAddress": vocab.PropNameMAC,
	"DateTime":   vocab.PropNameDateTime,
	"DeviceName": vocab.PropNameName,
	"HostName":   vocab.PropNameHostname,
}

// sensorTypeMap maps OWServer sensor names to IoT vocabulary
var SensorTypeVocab = map[string]struct {
	name     string
	dataType string
}{
	// "BarometricPressureHg": vocab.PropNameAtmosphericPressure, // unit Hg
	"BarometricPressureMb": {name: vocab.PropNameAtmosphericPressure,
		dataType: vocab.WoTDataTypeNumber}, // unit Mb
	"DewPoint":    {name: vocab.PropNameDewpoint, dataType: vocab.WoTDataTypeNumber},
	"HeatIndex":   {name: vocab.PropNameHeatIndex, dataType: vocab.WoTDataTypeNumber},
	"Humidity":    {name: vocab.PropNameHumidity, dataType: vocab.WoTDataTypeNumber},
	"Humidex":     {name: vocab.PropNameHumidex, dataType: vocab.WoTDataTypeNumber},
	"Light":       {name: vocab.PropNameLuminance, dataType: vocab.WoTDataTypeNumber},
	"RelayState":  {name: vocab.PropNameRelay, dataType: vocab.WoTDataTypeBool},
	"Temperature": {name: vocab.PropNameTemperature, dataType: vocab.WoTDataTypeNumber},
}

// unitNameMap maps OWServer unit names to IoT vocabulary
var UnitNameVocab = map[string]string{
	"PercentRelativeHumidity": vocab.UnitNamePercent,
	"Millibars":               vocab.UnitNameMillibar,
	"Centigrade":              vocab.UnitNameCelcius,
	"Fahrenheit":              vocab.UnitNameFahrenheit,
	"InchesOfMercury":         vocab.UnitNameMercury,
	"Lux":                     vocab.UnitNameLux,
	"#":                       vocab.UnitNameCount,
	"Volt":                    vocab.UnitNameVolt,
}

// EdsAPI EDS device API properties and methods
type EdsAPI struct {
	address         string     // EDS (IP) address or filename (file://./path/to/name.xml)
	loginName       string     // Basic Auth login name
	password        string     // Basic Auth password
	discoTimeoutSec int        // EDS OWServer discovery timeout
	readMutex       sync.Mutex // prevent concurrent discovery
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
	PropertyType vocab.ThingPropType
}
type OneWireNode struct {
	DeviceType vocab.DeviceType
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
		Attr:        make(map[string]OneWireAttr),
		DeviceType:  vocab.DeviceTypeGateway,
	}
	owNodeList = append(owNodeList, &owNode)
	// todo: find a better place for this
	if isRootNode {
		owAttr := OneWireAttr{
			Name:  vocab.PropNameLatency,
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
			propType := vocab.PropertyTypeAttr
			writable := (strings.ToLower(node.Writable) == "true")
			attrName := node.XMLName.Local
			sensorInfo, isSensor := SensorTypeVocab[attrName]
			if isSensor {
				attrName = sensorInfo.name
				propType = vocab.PropertyTypeSensor
				if writable {
					propType = vocab.PropertyTypeActuator
				}
			} else {
				propType = vocab.PropertyTypeAttr
				if writable {
					propType = vocab.PropertyTypeConfig
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
					deviceType = vocab.DeviceTypeUnknown
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
// The timeout for HTTP access is 1 second
func (edsAPI *EdsAPI) ReadEds() (rootNode *XMLNode, err error) {
	// don't discover or read concurrently
	edsAPI.readMutex.Lock()
	defer edsAPI.readMutex.Unlock()
	if edsAPI.address == "" {
		edsAPI.address, err = edsAPI.Discover()
		if err != nil {
			return nil, err
		}
	} else if strings.HasPrefix(edsAPI.address, "file://") {
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
	req, _ := http.NewRequest("GET", edsURL, nil)

	req.SetBasicAuth(edsAPI.loginName, edsAPI.password)
	client := &http.Client{Timeout: time.Second}
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

// Discover any EDS OWServer ENet-2 on the local network for 3 seconds
// This uses a UDP Broadcast on port 30303 as stated in the manual
// If found, this sets the service address for further use
// Returns the address or an error if not found
func (edsAPI *EdsAPI) Discover() (addr string, err error) {
	logrus.Infof("Starting discovery")
	// listen
	pc, err := net.ListenPacket("udp4", ":30303")
	if err != nil {
		return "", err
	}
	defer pc.Close()

	addr2, err := net.ResolveUDPAddr("udp4", "255.255.255.255:30303")
	if err != nil {
		return "", err
	}

	_, err = pc.WriteTo([]byte("D"), addr2)
	if err != nil {
		return "", err
	}

	buf := make([]byte, 1024)
	// receive 2 messages, first the broadcast, followed by the response, if there is one
	// wait 3 seconds before giving up
	for {
		pc.SetReadDeadline(time.Now().Add(time.Second * time.Duration(edsAPI.discoTimeoutSec)))
		n, remoteAddr, err := pc.ReadFrom(buf)
		if err != nil {
			logrus.Infof("Discovery ended without results")
			return "", err
		} else if n > 1 {
			switch rxAddr := remoteAddr.(type) {
			case *net.UDPAddr:
				addr = rxAddr.IP.String()
				logrus.Infof("EdsAPI.Discover. Found at %s: %s", addr, buf[:n])
				return addr, nil
			}
		}
	}

}

// NewEdsAPI creates a new NewEdsAPI instance
//  address is optional to override the discovery
//  loginName if needed, "" if not needed
//  password if needed, "" if not needed
func NewEdsAPI(address string, loginName string, password string) *EdsAPI {
	edsAPI := &EdsAPI{
		address:         address,
		loginName:       loginName,
		password:        password,
		discoTimeoutSec: 3, // discovery timeout
	}
	return edsAPI
}
