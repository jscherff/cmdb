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
	`bytes`
	`fmt`
	`io`
	`reflect`
	`time`
	`github.com/google/gousb`
)

// SETTING COMMAND
//
// The setting data command is a collection of function setting blocks:
//
// 	command:
//		<STX> <S> <FuncSETBLOCK1> ... <FuncSETBLOCKn> <ETX> <LRC>
//
//	Response:
//		<ACK> for success
//		<NAK> for wrong command (invalid function ID, length, value)
//
// Each function-setting block <FuncSETBLOCK> has following format:
//
//	<FuncID> <Len> <FuncData>
//
//
// GET SETTING COMMAND
//
// This command will send current setting to application.
//
//	command:
//		<STX> <R> <FuncID> <ETX> <LRC>
//
//	Response:
//		<ACK> <STX> <FuncID> <Len> <FuncData> <ETX> <LRC>
//
// Where:
//	<FuncID>	0x--		Setting Function ID
//	<Len>		0x--		Length of Setting Data
//	<FuncData>	----		Value of Setting
//	<STX>		0x02		Start of Text
//	<ETX>		0x03		End of Text
//	<ACK>		0x06		Acknowledge
//	<NAK>		----		Negative Acknowledge
//			0x15		RS232, USB HID <NAK>
//			0xFD		USB KB <NAK>
//	<UnknownID>	0x16		Unsupported ID
//	<AlreadyInPOS>	0x17		Already in OPOS mode
//	<R>		0x52		Review Setting
//	<S>		0x53		Send Setting
//	<LRC>		0x--		XOR of Previous Data

const (
	// Exported

	IDTechVID		uint16	= 0x0ACD
	IDTechKbPID		uint16	= 0x2030
	IDTechHidPID		uint16	= 0x2010

	IDTechValBeepNone	string	= `0`
	IDTechValBeepLowLong	string	= `1`
	IDTechValBeepHighLong	string	= `2`
	IDTechValBeepHighShort	string	= `3`
	IDTechValBeepLowShort	string	= `4`

	// Non-Exported

	idtechSymStartOfText	uint8	= 0x02
	idtechSymEndOfText	uint8	= 0x03
	idtechSymReviewSetting	uint8	= 0x52
	idtechSymSendSetting	uint8	= 0x53

	idtechCmdCopyright	uint8	= 0x38
	idtechCmdVersion	uint8	= 0x39
	idtechCmdReset		uint8	= 0x49

	idtechPropBeep		uint8	= 0x11
	idtechPropDeviceSN	uint8	= 0x4e
	idtechPropFirmwareVer	uint8	= 0x22

	idtechBufSizeSecureMag	int	= 8
)

type idtechRespCode uint8

func (r idtechRespCode) Ok() bool {
	return r == 0x06
}

func (r idtechRespCode) String() (v string) {

	switch r {

	case 0x06:
		v = `Acknowledge`
	case 0x15:
		v = `Negative Acknowledge`
	case 0x16:
		v = `Unknown ID`
	case 0x17:
		v = `Already in POS Mode`
	case 0xFD:
		v = `Negative Acknowledge`
	default:
		v = `Unknown Result Code`
	}

	return v
}

// IDTech decorates a gousb.Device with additional methods and properties.
type IDTech struct {
	*Device
}

// NewIDTech instantiates a IDTech wrapper for an existing gousb Device.
func NewIDTech(gd *gousb.Device) (this *IDTech, err error) {

	d, err := NewDevice(gd)

	if err != nil {
		return this, err
	}

	this = &IDTech{d}

	if gd == nil {
		return this, err
	}
	if this.Info.FirmwareVer, err = this.GetFirmwareVer(); err != nil {
		return this, err
	}
	if this.Info.ProductVer, err = this.GetProductVer(); err != nil {
		return this, err
	}
	if err = this.Refresh(); err != nil {
		return this, err
	}

	this.Info.SoftwareID = this.Info.FirmwareVer
	this.Info.ObjectType = reflect.TypeOf(this).String()

	return this, err
}

// Refresh updates API properties whose values may have changed.
func (this *IDTech) Refresh() (err error) {

	if this.Info.DeviceSN, err = this.GetDeviceSN(); err != nil {
		return err
	}
	if this.Info.DescriptorSN, err = this.SerialNumber(); err != nil {
		return err
	}

	this.Info.SerialNumber = this.Info.DeviceSN

	return err
}

// GetSoftwareID retrieves the software ID of the device from NVRAM.
func (this *IDTech) GetFirmwareVer() (string, error) {
	return this.getProperty(idtechPropFirmwareVer)
}

// EraseDeviceSN removes the device configurable serial number from NVRAM.
func (this *IDTech) EraseDeviceSN() (error) {
	return this.setProperty(idtechPropDeviceSN, ``)
}

// SetDeviceSN sets the device configurable serial number in NVRAM.
func (this *IDTech) SetDeviceSN(v string) (error) {
	return this.setProperty(idtechPropDeviceSN, v)
}

// GetDeviceSN retrieves the device configurable serial number from NVRAM.
func (this *IDTech) GetDeviceSN() (v string, err error) {
	return this.getProperty(idtechPropDeviceSN)
}

// SetBeep sets the beep frequency and duration on the device.
func (this *IDTech) SetBeep(v string) (err error) {
	return this.setProperty(idtechPropBeep, v)
}

// GetProductVer retrieves the product version of the device from NVRAM.
func (this *IDTech) GetProductVer() (v string, err error) {

	var cmd bytes.Buffer

	if err = cmd.WriteByte(idtechCmdVersion); err != nil {
		return v, err
	}
	if resp, err := this.sendCommand(cmd); err != nil {
		return v, err
	} else {
		v = string(bytes.TrimSpace(resp))
	}

	return v, err
}

// Reset overides inherited Reset method with a low-level vendor reset.
func (this *IDTech) Reset() (err error) {

	var cmd bytes.Buffer

	if err := cmd.WriteByte(idtechCmdReset); err != nil {
		return err
	}

	_, err = this.sendCommand(cmd)

	return err
}

// getProperty retrieves a property from device NVRAM using low-level commands.
func (this *IDTech) getProperty(p byte) (v string, err error) {

	var cmd bytes.Buffer

	if _, err := cmd.Write([]byte{idtechSymReviewSetting, p}); err != nil {
		return v, err
	}
	if resp, err := this.sendCommand(cmd); err != nil {
		return v, err
	} else {
		// Hack to accommodate inconsistent device API:
		// sometimes get setting command returns function
		// ID and value length, sometimes it returns just
		// the value.
		if len(resp) > 2 && resp[0] == p {
			resp = resp[2:]
		}
		v = string(bytes.TrimSpace(resp))
	}

	return v, err
}

// setProperty configures a property in device NVRAM using low-level commands.
func (this *IDTech) setProperty(p byte, v string) (err error) {

	var cmd bytes.Buffer

	if _, err := cmd.Write([]byte{idtechSymSendSetting, p}); err != nil {
		return err
	}
	if err := cmd.WriteByte(byte(len(v))); err != nil {
		return err
	}
	if _, err := cmd.WriteString(v); err != nil {
		return err
	}

	_, err = this.sendCommand(cmd)

	return err
}

// wrapCommand adds the prefix, suffix, LRC, and zero-padding to a command
// in preparation for transmission.
func (this *IDTech) wrapCommand(cin bytes.Buffer) (cout bytes.Buffer, err error) {

	if err = cout.WriteByte(idtechSymStartOfText); err != nil {
		return cout, err
	}
	if _, err = cout.Write(cin.Bytes()); err != nil {
		return cout, err
	}
	if err = cout.WriteByte(idtechSymEndOfText); err != nil {
		return cout, err
	}
	if err = cout.WriteByte(LRC(cout.Bytes())); err != nil {
		return cout, err
	}

	return cout, nil
}

// sendCommand wraps a command with necessary codes and sends to the device.
func (this *IDTech) sendCommand(cmd bytes.Buffer) (resp []byte, err error) {


	if cmd, err = this.wrapCommand(cmd); err != nil {
		return resp, err
	}

	buf := make([]byte, idtechBufSizeSecureMag)

	for {
		if _, err := cmd.Read(buf); err == io.EOF {
			break
		}
		if _, err = this.ControlSetReport(buf); err != nil {
			return resp, err
		}
	}

	time.Sleep(1 * time.Second)

	for {
		n, err := this.ControlGetReport(buf)

		if err != nil {
			return resp, err
		}
		if n == 0 {
			break
		}

		resp = append(resp, buf...)
	}

	resp = bytes.Trim(resp, "\x00")

	if len(resp) == 0 {
		return resp, fmt.Errorf(`no response`)
	}
	if rc := idtechRespCode(resp[0]); !rc.Ok() {
		err = fmt.Errorf(`device command response %d: %q`, rc, rc)
	}

	st := bytes.IndexByte(resp, idtechSymStartOfText) + 1
	et := bytes.IndexByte(resp, idtechSymEndOfText)

	if et > st {
		resp = resp[st:et]
	}

	return resp, err
}

// LRC performs a bitwise XOR on bytes in an array.
func LRC(bs []byte) (bx byte) {

	for _, b := range bs {
		bx ^= b
	}

	return bx
}
