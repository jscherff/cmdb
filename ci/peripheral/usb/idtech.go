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

	IDTechVID		= 0x0acd
	IDTechPID		= 0x2030 // Default

	IDTechKbPID		= 0x2030
	IDTechHidPID		= 0x2010

	IDTechValBeepNone	= `0`
	IDTechValBeepLowLong	= `1`
	IDTechValBeepHighLong	= `2`
	IDTechValBeepHighShort	= `3`
	IDTechValBeepLowShort	= `4`

	// Non-Exported

	idtechSymStartOfText	= 0x02
	idtechSymEndOfText	= 0x03
	idtechSymReviewSetting	= 0x52
	idtechSymSendSetting	= 0x53

	idtechCmdCopyright	= 0x38
	idtechCmdVersion	= 0x39
	idtechCmdReset		= 0x49

	idtechPropBeep		= 0x11
	idtechPropDeviceSN	= 0x4e
	idtechPropFirmwareVer	= 0x22

	idtechBufSizeSecureMag	= 8
)

// idtechRespCode represents the response code from control transfer
// vendor commands.
type idtechRespCode uint8

// Ok indicates control transfer vendor command success.
func (this idtechRespCode) Ok() bool {
	return this == 0x06
}

// String implements the Stringer interface for idtechRespCode.
func (this idtechRespCode) String() (s string) {

	switch this {

	case 0x06:
		s = `Acknowledge`
	case 0x15:
		s = `Negative Acknowledge`
	case 0x16:
		s = `Unknown ID`
	case 0x17:
		s = `Already in POS Mode`
	case 0xFD:
		s = `Negative Acknowledge`
	default:
		s = `Unknown Result Code`
	}

	return s
}

// Int converts the integer value of the idtechRespCode.
func (this idtechRespCode) Int() (n int) {
	return int(this)
}

// IDTech decorates a gousb.Device with additional methods and properties.
type IDTech struct {
	*Device
}

// IsIDTech helps determine whether or not an IDTech card reader.
func IsIDTech(vid, pid gousb.ID) (bool) {

	if vid != IDTechVID {
		return false
	}

	switch pid {
	case IDTechKbPID:
	case IDTechHidPID:
	default:
		return false
	}

	return true
}

// NewIDTech instantiates a IDTech wrapper for an existing gousb Device.
func NewIDTech(i interface{}) (this *IDTech, err error) {

	if d, err := NewDevice(i); err != nil {
		return nil, err
	} else {
		this = &IDTech{d}
	}

	this.ObjectType = fmt.Sprintf(`%T`, this)

	if _, ok := i.(*gousb.Device); !ok {
		return this, nil
	}

	if this.FirmwareVer, err = this.GetFirmwareVer(); err != nil {
		return this, err
	}
	if this.ProductVer, err = this.GetProductVer(); err != nil {
		return this, err
	}
	if err = this.Refresh(); err != nil {
		return this, err
	}

	this.SoftwareID = this.FirmwareVer

	return this, nil
}

// Refresh updates API properties whose values may have changed.
func (this *IDTech) Refresh() (err error) {

	if this.DeviceSN, err = this.GetDeviceSN(); err != nil {
		return err
	}
	if this.DescriptorSN, err = this.SerialNumber(); err != nil {
		return err
	}

	this.SerialNum = this.DeviceSN

	return err
}

// GetSoftwareID retrieves the software ID of the device from NVRAM.
func (this *IDTech) GetFirmwareVer() (string, error) {
	return this.getProperty(idtechPropFirmwareVer)
}

// GetDeviceSN retrieves the device configurable serial number from NVRAM.
func (this *IDTech) GetDeviceSN() (v string, err error) {
	return this.getProperty(idtechPropDeviceSN)
}

// SetDeviceSN sets the device configurable serial number in NVRAM.
func (this *IDTech) SetDeviceSN(v string) (error) {
	return this.setProperty(idtechPropDeviceSN, v)
}

// SetDefaultSN is a NOOP function to comply with the Serializer interface.
func (this *IDTech) SetDefaultSN(v string) (error) {
	return nil
}

// EraseDeviceSN removes the device configurable serial number from NVRAM.
func (this *IDTech) EraseDeviceSN() (error) {
	return this.setProperty(idtechPropDeviceSN, ``)
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
		if _, err = this.controlSetReport(buf); err != nil {
			return resp, err
		}
	}

	time.Sleep(1 * time.Second)

	for {
		n, err := this.controlGetReport(buf)

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
		return resp, fmt.Errorf(`command response nil`)
	}
	if rc := idtechRespCode(resp[0]); !rc.Ok() {
		err = fmt.Errorf(`command response %02x: %q`, rc.Int(), rc)
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
