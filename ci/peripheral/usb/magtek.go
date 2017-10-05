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
	`reflect`
	`time`

	`github.com/google/gousb`
)

const (
	MagtekVID		uint16	= 0x0801
	MagtekPID		uint16	= 0x0001

	SureswipeKbPID		uint16	= 0x0001
	MagnesafeKbPID		uint16	= 0x0001
	SureswipeHidPID		uint16	= 0x0002
	MagnesafeHidPID		uint16	= 0x0011

	CommandGetProp		uint8	= 0x00
	CommandSetProp		uint8	= 0x01
	CommandReset		uint8	= 0x02

	ResultSuccess		uint8	= 0x00
	ResultFailure		uint8	= 0x01
	ResultBadParam		uint8	= 0x02

	PropSoftwareID		uint8	= 0x00
	PropDeviceSN		uint8	= 0x01
	PropFactorySN		uint8	= 0x03
	PropProductVer		uint8	= 0x04

	DefaultSNLength		int	= 7
)

var (
	BufferSizes = []int{24, 60}
)

// Magtek decorates a gousb.Device with usb.Device and usb.Magtek
// Properties and API.
type Magtek struct {
	*Device
}

// NewMagtek instantiates a Magtek wrapper for an existing gousb Device.
func NewMagtek(gd *gousb.Device) (this *Magtek, err error) {

	d, err := NewDevice(gd)

	if err != nil {
		return this, err
	}

	this = &Magtek{d}

	if gd == nil {
		return this, err
	}
	if this.Info.BufferSize, err = this.GetBufferSize(); err != nil {
		return this, err
	}
	if this.Info.SoftwareID, err = this.GetSoftwareID(); err != nil {
		return this, err
	}
	if this.Info.ProductVer, err = this.GetProductVer(); err != nil {
		return this, err
	}
	if err = this.Refresh(); err != nil {
		return this, err
	}

	this.Info.ObjectType	= reflect.TypeOf(this).String()
	this.Info.FirmwareVer	= this.Info.SoftwareID

	return this, err
}

// Refresh updates API properties whose values may have changed.
func (this *Magtek) Refresh() (err error) {

	if this.Info.DeviceSN, err = this.GetDeviceSN(); err != nil {
		return err
	}
	if this.Info.FactorySN, err = this.GetFactorySN(); err != nil {
		return err
	}
	if this.Info.DescriptorSN, err = this.SerialNumber(); err != nil {
		return err
	}

	this.Info.SerialNumber	= this.Info.DeviceSN

	return err
}

// DeviceSN retrieves the device configurable serial number from NVRAM.
func (this *Magtek) GetDeviceSN() (string, error) {
	return this.getProperty(PropDeviceSN)
}

// FactorySN retrieves the device factory serial number from NVRAM.
func (this *Magtek) GetFactorySN() (string, error) {
	s, err := this.getProperty(PropFactorySN)
	if len(s) <= 1 {s = ``}
	return s, err
}

// SoftwareID retrieves the software ID of the device from NVRAM.
func (this *Magtek) GetSoftwareID() (string, error) {
	return this.getProperty(PropSoftwareID)
}

// ProductVer retrieves the product version of the device from NVRAM.
func (this *Magtek) GetProductVer() (string, error) {
	s, err := this.getProperty(PropProductVer)
	if len(s) <= 1 {s = ``}
	return s, err
}

// SetDeviceSN sets the device configurable serial number in NVRAM.
func (this *Magtek) SetDeviceSN(s string) (error) {
	return this.setProperty(PropDeviceSN, s)
}

// EraseDeviceSN removes the device configurable serial number from NVRAM.
func (this *Magtek) EraseDeviceSN() (error) {
	return this.setProperty(PropDeviceSN, ``)
}

// SetFactorySN sets the device factory device serial number in NVRAM.
// This will fail with result code 07 if serial number is already set.
func (this *Magtek) SetFactorySN(s string) (error) {
	return this.setProperty(PropFactorySN, s)
}

// CopyFactorySN copies 'length' characters from the device factory
// serial number to the configurable serial number in NVRAM.
func (this *Magtek) CopyFactorySN(n int) (err error) {

	s, err := this.GetFactorySN()

	if err != nil {
		return err
	}
	if s == `` {
		return fmt.Errorf(`no factory serial number`)
	}
	if n > len(s) {
		n = len(s)
	}

	err = this.SetDeviceSN(s[:n])

	return err
}

// Reset overides inherited Reset method with a low-level vendor reset.
func (this *Magtek) Reset() (err error) {

	data := make([]byte, this.Info.BufferSize)
	data[0] = CommandReset

	if _, err = this.ControlSetReport(data); err != nil {
		return err
	}
	if _, err = this.ControlGetReport(data); err != nil {
		return err
	}
	if data[0] > 0x00 {
		return fmt.Errorf(`command error: %X`, data[0])
	}

	time.Sleep(5 * time.Second)

	return err
}

// GetBufferSize uses trial and error to find the control transfer data
// buffer size of the device. Failure to use the correct size for control
// transfers carrying vendor commands will result in a LIBUSB_ERROR_PIPE
// error.
func (this *Magtek) GetBufferSize() (n int, err error) {

	for _, n = range BufferSizes {

		data := make([]byte, n)
		copy(data, []byte{CommandGetProp, 0x01, PropSoftwareID})

		if _, err = this.ControlSetReport(data); err != nil {
			continue
		}
		if _, err = this.ControlGetReport(data); err != nil {
			continue
		}

		break
	}

	return n, err
}

// getProperty retrieves a property from device NVRAM using low-level commands.
func (this *Magtek) getProperty(p byte) (s string, err error) {

	data := make([]byte, this.Info.BufferSize)
	copy(data, []byte{CommandGetProp, 0x01, p})

	if _, err = this.ControlSetReport(data); err != nil {
		return s, err
	}
	if _, err = this.ControlGetReport(data); err != nil {
		return s, err
	}
	if data[0] > 0x00 {
		return s, fmt.Errorf(`command error: %X`, data[0])
	}
	if data[1] > 0x00 {
		s = string(data[2:int(data[1])+2])
	}

	return s, err
}

// setProperty configures a property in device NVRAM using low-level commands.
func (this *Magtek) setProperty(p byte, s string) (err error) {

	data := make([]byte, this.Info.BufferSize)
	copy(data[0:], []byte{CommandSetProp, byte(len(s)+1), p})
	copy(data[3:], s)

	if _, err = this.ControlSetReport(data); err != nil {
		return err
	}
	if _, err = this.ControlGetReport(data); err != nil {
		return err
	}
	if data[0] > 0x00 {
		return fmt.Errorf(`command error: %X`, data[0])
	}

	this.Refresh()

	return err
}
