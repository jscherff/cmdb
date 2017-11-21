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

type Resetter interface {
	Reset() (error)
}

type Identifier interface {
	Resetter
	ID() (string)
	SN() (string)
	VID() (string)
	PID() (string)
	Host() (string)
	Type() (string)
	Conn() (string)
}

type Analyzer interface {
	Identifier
	GetState() (string, error)
}

type Reporter interface {
	Identifier
	CSV() ([]byte, error)
	NVP() ([]byte, error)
	XML() ([]byte, error)
	JSON() ([]byte, error)
	PrettyXML() ([]byte, error)
	PrettyJSON() ([]byte, error)
}

type Updater interface {
	Identifier
	GetVendorName() (string)
	GetProductName() (string)
	GetUSBClass() (string)
	GetUSBSubClass() (string)
	GetUSBProtocol() (string)
	SetVendorName(string)
	SetProductName(string)
	SetUSBClass(string)
	SetUSBSubClass(string)
	SetUSBProtocol(string)
}

type Auditer interface {
	Reporter
	Zero()
	Clone() (interface{})
	Save(string) (error)
	RestoreFile(string) (error)
	RestoreJSON([]byte) (error)
	CompareFile(string) ([][]string, error)
	CompareJSON([]byte) ([][]string, error)
	AuditFile(string) (error)
	AuditJSON([]byte) (error)
	SetChanges([][]string)
	GetChanges() ([][]string)
}

type Serializer interface {
	Auditer
	GetDeviceSN() (string, error)
	SetDeviceSN(string) (error)
	SetDefaultSN() (error)
	EraseDeviceSN() (error)
	Refresh() (error)
}

type FactorySerializer interface {
	Serializer
	GetFactorySN() (string, error)
	SetFactorySN(string) (error)
	CopyFactorySN(int) (error)
}
