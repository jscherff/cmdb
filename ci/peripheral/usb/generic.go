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

import `fmt`

// Generic decorates a gousb.Device with additional methods and properties.
type Generic struct{
	*Device
}

// NewGeneric instantiates a Generic wrapper for an existing gousb Device.
func NewGeneric(i interface{}) (this *Generic, err error) {

	if d, err := NewDevice(i); err != nil {
		return nil, err
	} else {
		this = &Generic{d}
	}

	this.ObjectType = fmt.Sprintf(`%T`, this)

	return this, nil
}
