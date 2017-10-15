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
	`fmt`

	`github.com/google/gousb`
	`github.com/jscherff/cmdb/metaci/peripheral/usb`
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
)

// Device decorates a gousb.Device with additional methods and properties.
type Device struct {
	*gousb.Device
	Info *usb.DeviceInfo
}

// NewDevice instantiates a Device wrapper for an existing gousb Device.
func NewDevice(gd *gousb.Device) (this *Device, err error) {

	this = &Device{Device: gd}

	if this.Info, err = usb.NewDeviceInfo(this.Desc); err != nil {
		return nil, err
	}
	if this.Info.VendorName, err = this.Manufacturer(); err != nil {
		return nil, err
	}
	if this.Info.ProductName, err = this.Product(); err != nil {
		return nil, err
	}
	if this.Info.SerialNumber, err = this.SerialNumber(); err != nil {
		return nil, err
	}

	this.Info.ObjectType = fmt.Sprintf(`%T`, this)

	return this, nil
}

// NilDevice instantiates a Device wrapper for an empty gousb Device.
func NilDevice() (this *Device, err error) {

	this = &Device{Device: &gousb.Device{}}

	if this.Info, err = usb.NilDeviceInfo(); err != nil {
		return nil, err
	}

	this.Info.ObjectType = fmt.Sprintf(`%T`, this)

	return this, nil
}

// NewDeviceFromDesc instantiates a new Device from a gousb DeviceDesc.
func NewDeviceFromDesc(gdd *gousb.DeviceDesc) (this *Device, err error) {

	this = &Device{Device: &gousb.Device{Desc: gdd}}

	if this.Info, err = usb.NewDeviceInfo(this.Desc); err != nil {
		return nil, err
	}

	this.Info.ObjectType = fmt.Sprintf(`%T`, this)

	return this, nil
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
func (this *Device) GetInfo() (*usb.DeviceInfo) {
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
