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
	`time`

	`github.com/google/gousb`
)

const (
	// Exported

	MagtekVID		= 0x0801
	MagtekPID		= 0x0001 // Default

	MagtekKbPID		= 0x0001
	MagtekSureswipeHidPID	= 0x0002
	MagtekMagnesafeHidPID	= 0x0011

	// Non-Exported

	magtekCmdGetProp	= 0x00
	magtekCmdSetProp	= 0x01
	magtekCmdReset		= 0x02

	magtekPropSoftwareID	= 0x00
	magtekPropDeviceSN	= 0x01
	magtekPropFactorySN	= 0x03
	magtekPropProductVer	= 0x04

	magtekBufSizeSureswipe	= 24
	magtekBufSizeMagnesafe	= 60

	magtekDefaultSNLength	= 7
)

var magtekBufferSizes = []int{24, 60}

// magtekRespCode represents the response code from control transfer
// vendor commands.
type magtekRespCode uint8

// Ok indicates control transfer vendor command success.
func (this magtekRespCode) Ok() bool {
	return this == 0x00
}

// String implements the Stringer interface for magtekRespCode.
func (this magtekRespCode) String() (s string) {

	switch this {

	case 0x00:
		s = `Success`
	case 0x01:
		s = `Failure`
	case 0x02:
		s = `Bad Parameter`
	case 0x05:
		s = `Delayed`
	case 0x07:
		s = `Invalid Operation`
	default:
		s = `Unknown Result Code`
	}

	return s
}

// Int converts the integer value of the magtekRespCode.
func (this magtekRespCode) Int() (n int) {
	return int(this)
}

// Magtek decorates a gousb.Device with additional methods and properties.
type Magtek struct {
	*Device
}

// IsMagtek helps determine whether or not an Magtek card reader.
func IsMagtek(vid, pid gousb.ID) (bool) {

	if vid != MagtekVID {
		return false
	}

	switch pid {
	case MagtekKbPID:
	case MagtekSureswipeHidPID:
	case MagtekMagnesafeHidPID:
	default:
		return false
	}

	return true
}

// NewMagtek instantiates a Magtek wrapper for an existing gousb Device.
func NewMagtek(i interface{}) (this *Magtek, err error) {

	if d, err := NewDevice(i); err != nil {
		return nil, err
	} else {
		this = &Magtek{d}
	}

	this.ObjectType = fmt.Sprintf(`%T`, this)

	if _, ok := i.(*gousb.Device); !ok {
		return this, nil
	}

	if this.BufferSize, err = this.getBufferSize(); err != nil {
		return this, err
	}
	if this.SoftwareID, err = this.GetSoftwareID(); err != nil {
		return this, err
	}
	if this.ProductVer, err = this.GetProductVer(); err != nil {
		return this, err
	}
	if err = this.Refresh(); err != nil {
		return this, err
	}

	this.FirmwareVer = this.SoftwareID

	return this, nil
}

// Refresh updates API properties whose values may have changed.
func (this *Magtek) Refresh() (err error) {

	if this.DeviceSN, err = this.GetDeviceSN(); err != nil {
		return err
	}
	if this.FactorySN, err = this.GetFactorySN(); err != nil {
		return err
	}
	if this.DescriptorSN, err = this.SerialNumber(); err != nil {
		return err
	}

	this.SerialNum = this.DeviceSN

	return err
}

// GetSoftwareID retrieves the software ID from NVRAM.
func (this *Magtek) GetSoftwareID() (string, error) {
	return this.getProperty(magtekPropSoftwareID)
}

// GetProductVer retrieves the product version from NVRAM.
func (this *Magtek) GetProductVer() (string, error) {

	if s, err := this.getProperty(magtekPropProductVer); err != nil {
		return ``, err
	} else if len(s) <= 1 {
		return ``, nil
	} else {
		return s, nil
	}
}

// GetDeviceSN retrieves the configurable serial number from NVRAM.
func (this *Magtek) GetDeviceSN() (string, error) {
	return this.getProperty(magtekPropDeviceSN)
}

// SetDeviceSN sets the configurable serial number in NVRAM.
func (this *Magtek) SetDeviceSN(s string) (error) {
	return this.setProperty(magtekPropDeviceSN, s)
}

// SetDefaultSN copies default-length characters from the factory
// serial number to the configurable serial number in NVRAM.
func (this *Magtek) SetDefaultSN() (error) {
	return this.CopyFactorySN(magtekDefaultSNLength)
}

// EraseDeviceSN removes the configurable serial number from NVRAM.
func (this *Magtek) EraseDeviceSN() (error) {
	return this.setProperty(magtekPropDeviceSN, ``)
}

// GetFactorySN retrieves the factory serial number from NVRAM.
func (this *Magtek) GetFactorySN() (string, error) {

	if s, err := this.getProperty(magtekPropFactorySN); err != nil {
		return ``, err
	} else if len(s) <= 1 {
		return ``, nil
	} else {
		return s, nil
	}
}

// SetFactorySN sets the factory device serial number in NVRAM. This
// will fail with result code 07 if serial number is already set.
func (this *Magtek) SetFactorySN(s string) (error) {
	return this.setProperty(magtekPropFactorySN, s)
}

// CopyFactorySN copies 'length' characters from the factory serial
// number to the configurable serial number in NVRAM.
func (this *Magtek) CopyFactorySN(n int) (error) {

	if s, err := this.GetFactorySN(); err != nil {
		return err
	} else if s == `` {
		return fmt.Errorf(`no factory serial number`)
	} else {
		return this.SetDeviceSN(s[:n])
	}
}

// Reset overides inherited Reset method with a low-level vendor reset.
func (this *Magtek) Reset() (error) {

	data := make([]byte, this.BufferSize)
	data[0] = magtekCmdReset

	if _, err := this.controlSetReport(data); err != nil {
		return err
	}
	if _, err := this.controlGetReport(data); err != nil {
		return err
	}
	if rc := magtekRespCode(data[0]); !rc.Ok() {
		return fmt.Errorf(`command response %02x: %q`, rc.Int(), rc)
	}

	time.Sleep(5 * time.Second)

	return nil
}

// getBufferSize uses trial and error to find the control transfer data
// buffer size of the device. Failure to use the correct size for control
// transfers carrying vendor commands will result in a LIBUSB_ERROR_PIPE
// error.
func (this *Magtek) getBufferSize() (n int, err error) {

	for _, n = range magtekBufferSizes {

		data := make([]byte, n)
		copy(data, []byte{magtekCmdGetProp, 0x01, magtekPropSoftwareID})

		if _, err = this.controlSetReport(data); err != nil {
			continue
		}
		if _, err = this.controlGetReport(data); err != nil {
			continue
		}

		break
	}

	return n, err
}

// getProperty retrieves a property from device NVRAM using low-level commands.
func (this *Magtek) getProperty(p byte) (string, error) {

	if this.BufferSize < 3 {
		return fmt.Errorf(`buffer size %d < %d`, this.BufferSize, 3)
	}

	data := make([]byte, this.BufferSize)
	copy(data, []byte{magtekCmdGetProp, 0x01, p})

	if _, err := this.controlSetReport(data); err != nil {
		return ``, err
	}
	if _, err := this.controlGetReport(data); err != nil {
		return ``, err
	}
	if rc := magtekRespCode(data[0]); !rc.Ok() {
		return ``, fmt.Errorf(`command response %02x: %q`, rc.Int(), rc)
	}
	if data[1] > 0x00 {
		data = data[2:int(data[1])+2]
	}

	return string(data), nil
}

// setProperty configures a property in device NVRAM using low-level commands.
func (this *Magtek) setProperty(p byte, v string) (error) {

	vlen := len(v)

	if this.BufferSize < vlen + 3 {
		return fmt.Errorf(`buffer size %d < %d`, this.BufferSize, vlen)
	}

	data := make([]byte, this.BufferSize)
	copy(data[0:], []byte{magtekCmdSetProp, byte(vlen), p})
	copy(data[3:], v)

	if _, err := this.controlSetReport(data); err != nil {
		return err
	}
	if _, err := this.controlGetReport(data); err != nil {
		return err
	}
	if rc := magtekRespCode(data[0]); !rc.Ok() {
		return fmt.Errorf(`command response %02x: %q`, rc.Int(), rc)
	}

	this.Refresh()

	return nil
}
