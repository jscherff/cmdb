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
	*gousb.Device `json:"-" xml:"-" csv:"-" nvp:"-" cmp:"-"`
	*usb.DeviceInfo
}

// NewDevice instantiates a Device wrapper for a gousb Device.
func NewDevice(i interface{}) (this *Device, err error) {

	switch t := i.(type) {

	case *gousb.Device:

		this = &Device{Device: t}

		if this.DeviceInfo, err = usb.NewDeviceInfo(this.Desc); err != nil {
			return nil, err
		}
		if this.SerialNum, err = this.SerialNumber(); err != nil {
			return nil, err
		}
		if this.VendorName, err = this.Manufacturer(); err != nil {
			return nil, err
		}
		if this.ProductName, err = this.Product(); err != nil {
			return nil, err
		}

	case *gousb.DeviceDesc:

		this = &Device{Device: &gousb.Device{Desc: t}}

		if this.DeviceInfo, err = usb.NewDeviceInfo(this.Desc); err != nil {
			return nil, err
		}

	case nil:

		this = &Device{Device: &gousb.Device{Desc: nil}}

		if this.DeviceInfo, err = usb.NewDeviceInfo(this.Desc); err != nil {
			return nil, err
		}

	default:

		return nil, fmt.Errorf(`unsupported base type %T`, t)
	}

	this.ObjectType = fmt.Sprintf(`%T`, this)

	return this, nil
}

// Zero replaces DeviceInfo with a new, empty DeviceInfo.
func (this *Device) Zero() {
	this.DeviceInfo = &usb.DeviceInfo{}
}

// Clone returns an empty instance of the device as an interface type.
func (this *Device) Clone() (interface{}) {
	return interface{}(new(Device))
}

// GetInfo detaches and returns just the DeviceInfo object.
func (this *Device) GetInfo() (*usb.DeviceInfo) {
	return this.DeviceInfo
}

// controlSetReport performs a SetReport control transfer.
func (this *Device) controlSetReport(data []byte) (n int, err error) {

	return this.Control(
		ReqDirectionOut | ReqTypeClass | ReqRecipInterface,
		ReqSetReport,
		FeatureReport,
		ControlInterface,
		data,
	)
}

// controlSetReport performs a SetReport control transfer.
func (this *Device) controlGetReport(data []byte) (n int, err error) {

	return this.Control(
		ReqDirectionIn | ReqTypeClass | ReqRecipInterface,
		ReqGetReport,
		FeatureReport,
		ControlInterface,
		data,
	)
}
