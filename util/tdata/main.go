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

package main

import (
	`crypto/sha256`
	`encoding/json`
	`log`
	`os`

	`github.com/jscherff/cmdb/ci/peripheral/usb`
)

type TestData struct {
	DevJSON map[string][]byte
	InfJSON map[string][]byte
	DevGen map[string]*usb.Generic
	DevMag map[string]*usb.Magtek
	DevIdt map[string]*usb.IDTech
	InfSig map[string]map[string][32]byte
	InfChg [][]string
	Clg []string
}

var (
	td = &TestData{

		DevJSON: make(map[string][]byte),
		InfJSON: make(map[string][]byte),

		DevGen: make(map[string]*usb.Generic),
		DevMag: make(map[string]*usb.Magtek),
		DevIdt: make(map[string]*usb.IDTech),

		InfSig: map[string]map[string][32]byte{
			`CSV`:  make(map[string][32]byte),
			`NVP`:  make(map[string][32]byte),
			`XML`:  make(map[string][32]byte),
			`JSN`:  make(map[string][32]byte),
			`Leg`:  make(map[string][32]byte),
			`PXML`: make(map[string][32]byte),
			`PJSN`: make(map[string][32]byte),
		},

		InfChg: [][]string{
			[]string{`SoftwareID`, `21042840G01`, `21042840G02`},
			[]string{`USBSpec`, `1.10`, `2.00`},
		},

		Clg: []string{
			`"SoftwareID" was "21042840G01", now "21042840G02"`,
			`"USBSpec" was "1.10", now "2.00"`,
		},
	}
)

func main() {

	fhi, err := os.Open(`objects.json`)

	if err != nil {
		log.Fatal(err)
	}

	defer fhi.Close()

	if err := json.NewDecoder(fhi).Decode(&td); err != nil {
		log.Fatal(err)
	}

	if err := generateSigs(); err != nil {
		log.Fatal(err)
	}

	if err := generateJson(); err != nil {
		log.Fatal(err)
	}

	fho, err := os.Create(`testdata.json`)

	if err != nil {
		log.Fatal(err)
	}

	defer fho.Close()

	if err := json.NewEncoder(fho).Encode(td); err != nil {
		log.Fatal(err)
	}
}

func generateSigs() error {

	for k, d := range td.DevGen {

		i := d.GetInfo()

		if b, err := i.CSV(); err != nil {
			return err
		} else {
			td.InfSig[`CSV`][k] = sha256.Sum256(b)
		}
		if b, err := i.NVP(); err != nil {
			return err
		} else {
			td.InfSig[`NVP`][k] = sha256.Sum256(b)
		}
		if b, err := i.XML(); err != nil {
			return err
		} else {
			td.InfSig[`XML`][k] = sha256.Sum256(b)
		}
		if b, err := i.JSON(); err != nil {
			return err
		} else {
			td.InfSig[`JSN`][k] = sha256.Sum256(b)
		}
		if b, err := i.PrettyXML(); err != nil {
			return err
		} else {
			td.InfSig[`PXML`][k] = sha256.Sum256(b)
		}
		if b, err := i.PrettyJSON(); err != nil {
			return err
		} else {
			td.InfSig[`PJSN`][k] = sha256.Sum256(b)
		}
	}

	for k, d := range td.DevMag {

		i := d.GetInfo()

		if b, err := i.CSV(); err != nil {
			return err
		} else {
			td.InfSig[`CSV`][k] = sha256.Sum256(b)
		}
		if b, err := i.NVP(); err != nil {
			return err
		} else {
			td.InfSig[`NVP`][k] = sha256.Sum256(b)
		}
		if b, err := i.XML(); err != nil {
			return err
		} else {
			td.InfSig[`XML`][k] = sha256.Sum256(b)
		}
		if b, err := i.JSON(); err != nil {
			return err
		} else {
			td.InfSig[`JSN`][k] = sha256.Sum256(b)
		}
		if b, err := i.PrettyXML(); err != nil {
			return err
		} else {
			td.InfSig[`PXML`][k] = sha256.Sum256(b)
		}
		if b, err := i.PrettyJSON(); err != nil {
			return err
		} else {
			td.InfSig[`PJSN`][k] = sha256.Sum256(b)
		}
	}

	for k, d := range td.DevIdt {

		i := d.GetInfo()

		if b, err := i.CSV(); err != nil {
			return err
		} else {
			td.InfSig[`CSV`][k] = sha256.Sum256(b)
		}
		if b, err := i.NVP(); err != nil {
			return err
		} else {
			td.InfSig[`NVP`][k] = sha256.Sum256(b)
		}
		if b, err := i.XML(); err != nil {
			return err
		} else {
			td.InfSig[`XML`][k] = sha256.Sum256(b)
		}
		if b, err := i.JSON(); err != nil {
			return err
		} else {
			td.InfSig[`JSN`][k] = sha256.Sum256(b)
		}
		if b, err := i.PrettyXML(); err != nil {
			return err
		} else {
			td.InfSig[`PXML`][k] = sha256.Sum256(b)
		}
		if b, err := i.PrettyJSON(); err != nil {
			return err
		} else {
			td.InfSig[`PJSN`][k] = sha256.Sum256(b)
		}
	}

	return nil
}

func generateJson() error {

	for k, d := range td.DevGen {

		if b, err := json.Marshal(d); err != nil {
			return err
		} else {
			td.DevJSON[k] = b
		}

		i := d.GetInfo()

		if b, err := json.Marshal(i); err != nil {
			return err
		} else {
			td.InfJSON[k] = b
		}
	}

	for k, d := range td.DevMag {

		if b, err := json.Marshal(d); err != nil {
			return err
		} else {
			td.DevJSON[k] = b
		}

		i := d.GetInfo()

		if b, err := json.Marshal(i); err != nil {
			return err
		} else {
			td.InfJSON[k] = b
		}
	}

	for k, d := range td.DevIdt {

		if b, err := json.Marshal(d); err != nil {
			return err
		} else {
			td.DevJSON[k] = b
		}

		i := d.GetInfo()

		if b, err := json.Marshal(i); err != nil {
			return err
		} else {
			td.InfJSON[k] = b
		}
	}

	return nil
}
