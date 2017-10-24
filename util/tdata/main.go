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
	Jsn map[string][]byte
	Gen map[string]*usb.Generic
	Mag map[string]*usb.Magtek
	Idt map[string]*usb.IDTech
	Sig map[string]map[string][32]byte
	Chg [][]string
	Clg []string
}

var (
	td = &TestData{

		Jsn: make(map[string][]byte),

		Gen: make(map[string]*usb.Generic),
		Mag: make(map[string]*usb.Magtek),
		Idt: make(map[string]*usb.IDTech),

		Sig: map[string]map[string][32]byte{
			`CSV`:  make(map[string][32]byte),
			`NVP`:  make(map[string][32]byte),
			`XML`:  make(map[string][32]byte),
			`JSN`:  make(map[string][32]byte),
			`Leg`:  make(map[string][32]byte),
			`PXML`: make(map[string][32]byte),
			`PJSN`: make(map[string][32]byte),
		},

		Chg: [][]string{
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

	fho, err := os.Create(`tdata.json`)

	if err != nil {
		log.Fatal(err)
	}

	defer fho.Close()

	if err := json.NewEncoder(fho).Encode(td); err != nil {
		log.Fatal(err)
	}
}

func generateSigs() error {

	for k, d := range td.Gen {

		if b, err := d.CSV(); err != nil {
			return err
		} else {
			td.Sig[`CSV`][k] = sha256.Sum256(b)
		}
		if b, err := d.NVP(); err != nil {
			return err
		} else {
			td.Sig[`NVP`][k] = sha256.Sum256(b)
		}
		if b, err := d.XML(); err != nil {
			return err
		} else {
			td.Sig[`XML`][k] = sha256.Sum256(b)
		}
		if b, err := d.JSON(); err != nil {
			return err
		} else {
			td.Sig[`JSN`][k] = sha256.Sum256(b)
		}
		if b, err := d.PrettyXML(); err != nil {
			return err
		} else {
			td.Sig[`PXML`][k] = sha256.Sum256(b)
		}
		if b, err := d.PrettyJSON(); err != nil {
			return err
		} else {
			td.Sig[`PJSN`][k] = sha256.Sum256(b)
		}
	}

	for k, d := range td.Mag {

		if b, err := d.CSV(); err != nil {
			return err
		} else {
			td.Sig[`CSV`][k] = sha256.Sum256(b)
		}
		if b, err := d.NVP(); err != nil {
			return err
		} else {
			td.Sig[`NVP`][k] = sha256.Sum256(b)
		}
		if b, err := d.XML(); err != nil {
			return err
		} else {
			td.Sig[`XML`][k] = sha256.Sum256(b)
		}
		if b, err := d.JSON(); err != nil {
			return err
		} else {
			td.Sig[`JSN`][k] = sha256.Sum256(b)
		}
		if b, err := d.PrettyXML(); err != nil {
			return err
		} else {
			td.Sig[`PXML`][k] = sha256.Sum256(b)
		}
		if b, err := d.PrettyJSON(); err != nil {
			return err
		} else {
			td.Sig[`PJSN`][k] = sha256.Sum256(b)
		}
	}

	for k, d := range td.Idt {

		if b, err := d.CSV(); err != nil {
			return err
		} else {
			td.Sig[`CSV`][k] = sha256.Sum256(b)
		}
		if b, err := d.NVP(); err != nil {
			return err
		} else {
			td.Sig[`NVP`][k] = sha256.Sum256(b)
		}
		if b, err := d.XML(); err != nil {
			return err
		} else {
			td.Sig[`XML`][k] = sha256.Sum256(b)
		}
		if b, err := d.JSON(); err != nil {
			return err
		} else {
			td.Sig[`JSN`][k] = sha256.Sum256(b)
		}
		if b, err := d.PrettyXML(); err != nil {
			return err
		} else {
			td.Sig[`PXML`][k] = sha256.Sum256(b)
		}
		if b, err := d.PrettyJSON(); err != nil {
			return err
		} else {
			td.Sig[`PJSN`][k] = sha256.Sum256(b)
		}
	}

	return nil
}

func generateJson() error {

	for k, d := range td.Gen {

		if b, err := json.Marshal(d); err != nil {
			return err
		} else {
			td.Jsn[k] = b
		}
	}

	for k, d := range td.Mag {

		if b, err := json.Marshal(d); err != nil {
			return err
		} else {
			td.Jsn[k] = b
		}
	}

	for k, d := range td.Idt {

		if b, err := json.Marshal(d); err != nil {
			return err
		} else {
			td.Jsn[k] = b
		}
	}

	return nil
}
