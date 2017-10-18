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
	`time`
)

/*
	D01 -- 22 09 00 03 00 00 00 00 02 52 4e 03 1d 00 00 00
	U02 -- 21 09 00 03 00 00 08 00

	D03 -- 22 01 00 03 00 00 00 00 
	U04 -- A1 01 00 03 00 00 08 00 06 02 4e 0b 0a 35 35 31
	                                              ^^ ^^ ^^
	D05 -- 22 01 00 03 00 00 00 00 
	U06 -- A1 01 00 03 00 00 08 00 55 30 34 33 37 32 38 03
	                               ^^ ^^ ^^ ^^ ^^ ^^ ^^
	D07 -- 22 01 00 03 00 00 00 00 
	U08 -- A1 01 00 03 00 00 08 00 20 00 00 00 00 00 00 00

	D09 -- 22 01 00 03 00 00 00 00
	U10 -- A1 01 00 03 00 00 08 00

	D11 -- 22 01 00 03 00 00 00 00
	U12 -- A1 01 00 03 00 00 08 00

	D13 -- 22 01 00 03 00 00 00 00
	U14 -- A1 01 00 03 00 00 08 00

	D15 -- 22 01 00 03 00 00 00 00
	U16 -- A1 01 00 03 00 00 08 00

	D17 -- 22 01 00 03 00 00 00 00
	U18 -- A1 01 00 03 00 00 08 00
*/

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	ctx := gousb.NewContext()
	defer ctx.Close()

	dev, _ := ctx.OpenDeviceWithVIDPID(gousb.ID(0x0acd), gousb.ID(0x2030))

	if dev == nil {
		log.Fatal(`Device not found`)
	}

	fmt.Println(dev)
	defer dev.Close()

	buf := make([][]byte, 9)

	for i := 0; i < 9; i++ {
		buf[i] = make([]byte, 24)
	}

	fmt.Println("Pausing 5 seconds to start packet capture")
	time.Sleep(5 * time.Second)

	copy(buf[0], []byte{0x02, 0x52, 0x4e, 0x03})
	buf[0][4] = xor(buf[0])

	for i := 0; i < 9; i++ {

		var n int
		var err error

		fmt.Printf("Packet %d start", i)

		if i == 0 {
			n, err = dev.Control(0x21, 0x09, 0x0300, 0x0000, buf[i])
			time.Sleep(1 * time.Second)
		} else {
			n, err = dev.Control(0xA1, 0x01, 0x0300, 0x0000, buf[i])
		}

		if err != nil {
			log.Fatalf("Packet %d error %v\n", i, err)
		} else {
			fmt.Printf("Packet %d bytes %d\n", i, n)

			fmt.Printf("[%02x %02x %02x %02x %02x %02x %02x %02x]\n",
				buf[i][0], buf[i][1], buf[i][2], buf[i][3],
				buf[i][4], buf[i][5], buf[i][6], buf[i][7])

			fmt.Printf("'%s'\n", string(buf[i]))
		}
	}
}

func xor(bs []byte) (bx byte) {
	for _, b := range bs {
		bx ^= b
	}
	return bx
}
