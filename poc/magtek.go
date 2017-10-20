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
	`github.com/jscherff/cmdb`
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

	if s, err := mdev.GetFactorySN(); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(`SN:`, s)
	}

	if err := mdev.CopyFactorySN(7); err != nil {
		fmt.Println(err)
	} else {
		mdev.Reset()
	}

	if err = mdev.SetFactorySN(`ABCDEFGHIJKLMNO`); err != nil {
		fmt.Println(err)
	} else {
		mdev.Reset()
	}

	fmt.Printf("VID = %T, PID = %T\n", mdev.Desc.Vendor, mdev.Desc.Product)

	b, err := mdev.PrettyJSON()
	fmt.Println(string(b))

	mdev.Save(mdev.ID() + `.json`)

	bad := []string{`a`,`b`,`c`}

	if _, err := usb.NewMagtek(bad); err != nil {
		fmt.Println(err)
	}

	var i interface{} = mdev

	switch t := i.(type) {
	case cmdb.Serializer:
		fmt.Printf("YES %T\n", t)
	}

	showID(mdev)
}

func showID(s cmdb.Serializer) {
	fmt.Println(s.ID())
}
