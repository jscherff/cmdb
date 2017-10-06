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
	`os`
	`reflect`

	`github.com/google/gousb`
)

const (
	ReqDirectionOut		uint8	= 0x00
	ReqDirectionIn		uint8	= 0x80

	ReqTypeStandard		uint8	= 0x00
	ReqTypeClass		uint8	= 0x20
	ReqTypeVendor		uint8	= 0x40

	ReqRecipDevice		uint8	= 0x00
	ReqRecipInterface	uint8	= 0x01
	ReqRecipEndpoint	uint8	= 0x02
	ReqRecipOther		uint8	= 0x03

	ReqGetReport		uint8	= 0x01
	ReqSetReport		uint8	= 0x09
	ReqGetDescriptor	uint8	= 0x06

	DeviceDescriptor	uint16	= 0x0100
	ConfigDescriptor	uint16	= 0x0200
	HidDescriptor		uint16	= 0x2200
	FeatureReport		uint16	= 0x0300

	ControlInterface	uint16	= 0x0000

	DeviceDescSize		int	= 18
	ConfigDescSize		int	= 9

	MarshalPrefix		string	= ""
	MarshalIndent		string	= "\t"
)

// Device decorates a gousb.Device with additional methods and properties.
type Device struct {
	*gousb.Device
	Info *DeviceInfo
}

// NewDevice instantiates a Device wrapper for an existing gousb Device.
func NewDevice(gd *gousb.Device) (this *Device, err error) {

	this = &Device{Device: gd, Info: &DeviceInfo{}}

	if gd == nil {
		return this, err
	}

	this.Info.VendorID	= this.Desc.Vendor.String()
	this.Info.ProductID	= this.Desc.Product.String()
	this.Info.PortNumber	= this.Desc.Port
	this.Info.BusNumber	= this.Desc.Bus
	this.Info.BusAddress	= this.Desc.Address
	this.Info.MaxPktSize	= this.Desc.MaxControlPacketSize
	this.Info.USBSpec	= this.Desc.Spec.String()
	this.Info.USBClass	= this.Desc.Class.String()
	this.Info.USBSubclass	= this.Desc.SubClass.String()
	this.Info.USBProtocol	= this.Desc.Protocol.String()
	this.Info.DeviceSpeed	= this.Desc.Speed.String()
	this.Info.DeviceVer	= this.Desc.Device.String()
	this.Info.ObjectType	= reflect.TypeOf(this).String()

	// this.Info.SoftwareID
	// this.Info.FirmwareVer
	// this.Info.ProductVer
	// this.Info.BufferSize
	// this.Info.DeviceSN
	// this.Info.FactorySN
	// this.Info.DescriptorSN

	if this.Info.VendorName, err = this.Manufacturer(); err != nil {
		return this, err
	}
	if this.Info.ProductName, err = this.Product(); err != nil {
		return this, err
	}
	if this.Info.SerialNumber, err = this.SerialNumber(); err != nil {
		return this, err
	}

	this.Info.HostName, err = os.Hostname()

	return this, err
}

// ID is a convenience method to retrieve the device serial number.
func (this *Device) ID() (string) {
	return this.Info.SerialNumber
}

// VID is a convenience method to retrieve the device vendor ID.
func (this *Device) VID() (string) {
	return this.Info.VendorID
}

// PID is a convenience method to retrieve the device product ID.
func (this *Device) PID() (string) {
	return this.Info.ProductID
}

// Host is a convenience method to retrieve the device hostname.
func (this *Device) Host() (string) {
	return this.Info.HostName
}

// Type is a convenience method to help identify object type to other apps.
func (this *Device) Type() (string) {
	return this.Info.ObjectType
}

// GetInfo detaches and returns just the DeviceInfo object.
func (this *Device) GetInfo() (*DeviceInfo) {
	return this.Info
}

// ControlSetReport performs a SetReport control transfer.
func (this *Device) ControlSetReport(data []byte) (n int, err error) {

	return this.Control(
		ReqDirectionOut | ReqTypeClass | ReqRecipInterface,
		ReqSetReport,
		FeatureReport,
		ControlInterface,
		data,
	)
}

// ControlSetReport performs a SetReport control transfer.
func (this *Device) ControlGetReport(data []byte) (n int, err error) {

	return this.Control(
		ReqDirectionIn | ReqTypeClass | ReqRecipInterface,
		ReqGetReport,
		FeatureReport,
		ControlInterface,
		data,
	)
}
