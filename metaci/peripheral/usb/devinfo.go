// Copyright 2017 John Scherff
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package usb

import (
	`encoding/json`
	`encoding/xml`
	`fmt`
	`os`

	`github.com/google/gousb`
	`github.com/jscherff/goutil`
)

const (
	MarshalPrefix		string	= ""
	MarshalIndent		string	= "\t"
)

type DeviceInfo struct {

	HostName	string		`json:"host_name"     csv:"host_name"`
	VendorID	string		`json:"vendor_id"     csv:"vendor_id"`
	ProductID	string		`json:"product_id"    csv:"product_id"`
	SerialNumber	string		`json:"serial_number" csv:"serial_number"`
	VendorName	string		`json:"vendor_name"   csv:"vendor_name"`
	ProductName	string		`json:"product_name"  csv:"product_name"`
	ProductVer	string		`json:"product_ver"   csv:"product_ver"`
	FirmwareVer	string		`json:"firmware_ver"  csv:"firmware_ver"`
	SoftwareID	string		`json:"software_id"   csv:"software_id"`

	PortNumber	int		`json:"port_number"   csv:"-" nvp:"-" cmp:"-"`
	BusNumber	int		`json:"bus_number"    csv:"-" nvp:"-" cmp:"-"`
	BusAddress	int		`json:"bus_address"   csv:"-" nvp:"-" cmp:"-"`
	BufferSize	int		`json:"buffer_size"   csv:"-" nvp:"-"`
	MaxPktSize	int		`json:"max_pkt_size"  csv:"-" nvp:"-"`
	USBSpec		string		`json:"usb_spec"      csv:"-" nvp:"-"`
	USBClass	string		`json:"usb_class"     csv:"-" nvp:"-"`
	USBSubClass	string		`json:"usb_subclass"  csv:"-" nvp:"-"`
	USBProtocol	string		`json:"usb_protocol"  csv:"-" nvp:"-"`
	DeviceSpeed	string		`json:"device_speed"  csv:"-" nvp:"-"`
	DeviceVer	string		`json:"device_ver"    csv:"-" nvp:"-"`
	ObjectType	string		`json:"object_type"   csv:"-" nvp:"-"`

	DeviceSN	string		`json:"device_sn"     csv:"-" nvp:"-"`
	FactorySN	string		`json:"factory_sn"    csv:"-" nvp:"-"`
	DescriptorSN	string		`json:"descriptor_sn" csv:"-" nvp:"-"`

	Custom01	string		`json:"custom_01,omitempty" xml:",omitempty" csv:"-" nvp:"-"`
	Custom02	string		`json:"custom_02,omitempty" xml:",omitempty" csv:"-" nvp:"-"`
	Custom03	string		`json:"custom_03,omitempty" xml:",omitempty" csv:"-" nvp:"-"`
	Custom04	string		`json:"custom_04,omitempty" xml:",omitempty" csv:"-" nvp:"-"`
	Custom05	string		`json:"custom_05,omitempty" xml:",omitempty" csv:"-" nvp:"-"`
	Custom06	string		`json:"custom_06,omitempty" xml:",omitempty" csv:"-" nvp:"-"`
	Custom07	string		`json:"custom_07,omitempty" xml:",omitempty" csv:"-" nvp:"-"`
	Custom08	string		`json:"custom_08,omitempty" xml:",omitempty" csv:"-" nvp:"-"`
	Custom09	string		`json:"custom_09,omitempty" xml:",omitempty" csv:"-" nvp:"-"`
	Custom10	string		`json:"custom_10,omitempty" xml:",omitempty" csv:"-" nvp:"-"`
}

func NewDeviceInfo(desc *gousb.DeviceDesc) (this *DeviceInfo, err error) {

	this = &DeviceInfo{
		VendorID:	desc.Vendor.String(),
		ProductID:	desc.Product.String(),
		PortNumber:	desc.Port,
		BusNumber:	desc.Bus,
		BusAddress:	desc.Address,
		MaxPktSize:	desc.MaxControlPacketSize,
		USBSpec:	desc.Spec.String(),
		USBClass:	desc.Class.String(),
		USBSubClass:	desc.SubClass.String(),
		USBProtocol:	desc.Protocol.String(),
		DeviceSpeed:	desc.Speed.String(),
		DeviceVer:	desc.Device.String(),
	}

	this.ObjectType = fmt.Sprintf(`%T`, this)

	if this.HostName, err = os.Hostname(); err != nil {
		return nil, err
	}

	return this, nil
}

// ID is a convenience method to retrieve the device serial number.
func (this *DeviceInfo) ID() (string) {
	return this.SerialNumber
}

// VID is a convenience method to retrieve the device vendor ID.
func (this *DeviceInfo) VID() (string) {
	return this.VendorID
}

// PID is a convenience method to retrieve the device product ID.
func (this *DeviceInfo) PID() (string) {
	return this.ProductID
}

// Host is a convenience method to retrieve the device hostname.
func (this *DeviceInfo) Host() (string) {
	return this.HostName
}

// Type is a convenience method to help identify object type to other apps.
func (this *DeviceInfo) Type() (string) {
	return this.ObjectType
}

// Save saves the object to a JSON file.
func (this *DeviceInfo) Save(fn string) (error) {
	return goutil.SaveObject(this, fn)
}

// RestoreFile restores the object from a JSON file.
func (this *DeviceInfo) RestoreFile(fn string) (error) {
	return goutil.RestoreObject(fn, this)
}

// RestoreJSON restores the object from a JSON file.
func (this *DeviceInfo) RestoreJSON(j []byte) (error) {
	return json.Unmarshal(j, &this)
}

// CompareFile compares fields and properties and returns array of changes.
func (this *DeviceInfo) CompareFile(fn string) (ss [][]string, err error) {

	other := &DeviceInfo{}

	if err = other.RestoreFile(fn); err != nil {
		return ss, err
	}

	return goutil.CompareObjects(other, this, `cmp`)
}

// CompareJSON compares fields and properties and returns an array of differences.
func (this *DeviceInfo) CompareJSON(b []byte) (ss [][]string, err error) {

	other := &DeviceInfo{}

	if err = other.RestoreJSON(b); err != nil {
		return ss, err
	}

	return goutil.CompareObjects(other, this, `cmp`)
}

// JSON reports all unfiltered fields in JSON format.
func (this *DeviceInfo) JSON() ([]byte, error) {
	return json.Marshal(this)
}

// XML reports all unfiltered fields in XML format.
func (this *DeviceInfo) XML() ([]byte, error) {
	return xml.Marshal(this)
}

// CSV reports all unfiltered fields in CSV format.
func (this *DeviceInfo) CSV() ([]byte, error) {
	return goutil.ObjectToCSV(this)
}

// NVP reports all unfiltered fields as name-value pairs.
func (this *DeviceInfo) NVP() ([]byte, error) {
	return goutil.ObjectToNVP(this)
}

// PrettyJSON reports all unfiltered fields in formatted JSON format.
func (this *DeviceInfo) PrettyJSON() ([]byte, error) {
	return json.MarshalIndent(this, MarshalPrefix, MarshalIndent)
}

// PrettyXML reports all unfiltered fields in formatted XML format.
func (this *DeviceInfo) PrettyXML() ([]byte, error) {
	return json.MarshalIndent(this, MarshalPrefix, MarshalIndent)
}
