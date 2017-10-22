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
	`fmt`
	`log`
	`github.com/google/gousb`
	`github.com/jscherff/cmdb/ci/peripheral/usb`
)

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	ctx := gousb.NewContext()
	defer ctx.Close()

	dev, _ := ctx.OpenDeviceWithVIDPID(gousb.ID(0x0801), gousb.ID(0x0001))

	if dev == nil {
		log.Fatal(`Device not found`)
	}

	defer dev.Close()

	mdev, err := usb.NewMagtek(dev)

	if err != nil {
		log.Fatal(err)
	}

	var idev interface{} = mdev

	if d, ok := idev.(*usb.Magtek); ok {

		if err := d.EraseDeviceSN(); err != nil {
			fmt.Println(err)
		} else {
			d.Reset()
		}

		if b, err := d.PrettyJSON(); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(string(b))
		}

		if err := d.SetDeviceSN(`deadbeef`); err != nil {
			fmt.Println(err)
		} else {
			d.Reset()
		}

		if b, err := d.PrettyJSON(); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(string(b))
		}

		d.Save(d.ID() + `.json`)
	}
}
