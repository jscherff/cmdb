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
	`bufio`
	`encoding/json`
	`fmt`
	`io/ioutil`
	`net/http`
	`os`
	`path/filepath`
	`regexp`
)

const (
	usbMetaDirMode = 0755
	usbMetaFileMode = 0644
	usbMetaSourceUrl = `http://www.linux-usb.org/usb.ids`
)

var (
	// urlRgx helps determine if the init parameter is a file or URL.
	urlRgx = regexp.MustCompile(`^https?://`)

	// usbVendorRgx extracts the vendor IDs and names from the source data.
	vendorRgx = regexp.MustCompile(`^([0-9A-Fa-f]{4})\s+(.+)$`)

	// usbProductRgx extracts the product IDs and names from the source data.
	productRgx = regexp.MustCompile(`^\t([0-9A-Fa-f]{4})\s+(.+)$`)

	// usbClassRgx extracts the class IDs and descriptions from the source data.
	classRgx = regexp.MustCompile(`^C\s+([0-9A-Fa-f]{2})\s+(.+)$`)

	// usbSubclassRgx extracts the subclass IDs and descriptions from the source data.
	subclassRgx = regexp.MustCompile(`^\t([0-9A-Fa-f]{2})\s+(.+)$`)

	// usbProtocolRgx extracts the protocol IDs and descriptions from the source data.
	protocolRgx = regexp.MustCompile(`^\t\t([0-9A-Fa-f]{2})\s+(.+)$`)
)

// Usb contains all known information USB this.Vendors, products, this.Classes,
// subthis.Classes, and protocols.
type Usb struct {
	Vendors map[string]*Vendor
	Classes map[string]*Class
}

// Vendor contains the vendor name and mappings to all the vendor's products.
type Vendor struct {
	Name string
	Product	map[string]*Product
}

// String returns the name of the vendor.
func (this *Vendor) String() string {
	return this.Name
}

// Product contains the name of the product.
type Product struct {
	Name string
}

// String returns the name of the product.
func (this *Product) String() string {
	return this.Name
}

// Class contains the name of the class and mappings for each subclass.
type Class struct {
	Name string
	Subclass map[string]*Subclass
}

// String returns the name of the class.
func (this *Class) String() string {
	return this.Name
}

// Subclass contains the name of the subclass and any associated protocols.
type Subclass struct {
	Name string
	Protocol map[string]*Protocol
}

// String returns the name of the Subclass.
func (this *Subclass) String() string {
	return this.Name
}

// Protocol contains the name of the protocol.
type Protocol struct {
	Name string
}

// String returns the name of the protocol.
func (this *Protocol) String() string {
	return this.Name
}

// NewUsb creates a new instance of Usb with empty vendor/class maps.
func NewUsb(f string) (*Usb, error) {

	this := &Usb{
		make(map[string]*Vendor),
		make(map[string]*Class),
	}

	if _, err := os.Stat(f); os.IsExist(err) {

		if err := this.Load(f); err != nil {
			return nil, err
		}

	} else {

		if err := this.Load(usbMetaSourceUrl); err != nil {
			return nil, err
		}
		if err := os.MkdirAll(filepath.Dir(f), usbMetaDirMode); err != nil {
			return nil, err
		}
		if err = this.Save(f); err != nil {
			return nil, err
		}
	}

	return this, nil
}

// GetVendor returns the USB vendor associated with a vendor ID.
func (this *Usb) GetVendor(vid string) (*Vendor, error) {

	if v, ok := this.Vendors[vid]; !ok {
		return nil, fmt.Errorf(`vendor %q not found`, vid)
	} else {
		return v, nil
	}
}

// GetProduct returns the USB product associated with a product ID.
func (this *Vendor) GetProduct(pid string) (*Product, error) {

	if p, ok := this.Product[pid]; !ok {
		return nil, fmt.Errorf(`product %q not found`, pid)
	} else {
		return p, nil
	}
}

// GetClass returns the USB class associated with a class ID.
func (this *Usb) GetClass(cid string) (*Class, error) {

	if c, ok := this.Classes[cid]; !ok {
		return nil, fmt.Errorf(`class %q not found`, cid)
	} else {
		return c, nil
	}
}

// GetSubclass returns the USB subclass associated with a subclass ID.
func (this *Class) GetSubclass(sid string) (*Subclass, error) {

	if s, ok := this.Subclass[sid]; !ok {
		return nil, fmt.Errorf(`subclass %q not found`, sid)
	} else {
		return s, nil
	}
}

// GetProtocol returns the USB protocol associated with a protocol ID.
func (this *Subclass) GetProtocol(pid string) (*Protocol, error) {

	if p, ok := this.Protocol[pid]; !ok {
		return nil, fmt.Errorf(`protocol %q not found`, pid)
	} else {
		return p, nil
	}
}

// LoadUrl reads vendor, product, class, subclass, and protocol information
// line-by-line from an io.Reader source and populates data structures for
// use by the application.
func (this *Usb) LoadUrl(url string) error {

	var (
		vendor   *Vendor
		class    *Class
		subclass *Subclass
	)

	resp, err := http.Get(url)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {

		if matches := productRgx.FindStringSubmatch(scanner.Text()); matches != nil {

			if vendor == nil {
				return fmt.Errorf(`product with no vendor`)
			}
			if vendor.Product == nil {
				vendor.Product = make(map[string]*Product)
			}

			vendor.Product[matches[1]] = &Product{Name: matches[2]}
			continue
		}

		if matches := vendorRgx.FindStringSubmatch(scanner.Text()); matches != nil {

			vendor = &Vendor{Name: matches[2]}
			this.Vendors[matches[1]] = vendor
			continue
		}

		if matches := protocolRgx.FindStringSubmatch(scanner.Text()); matches != nil {

			if subclass == nil {
				return fmt.Errorf(`protocol with no subclass`)
			}
			if subclass.Protocol == nil {
				subclass.Protocol = make(map[string]*Protocol)
			}

			subclass.Protocol[matches[1]] = &Protocol{Name: matches[2]}
			continue
		}

		if matches := subclassRgx.FindStringSubmatch(scanner.Text()); matches != nil {

			if class == nil {
				return fmt.Errorf(`subclass with no class`)
			}
			if class.Subclass == nil {
				class.Subclass = make(map[string]*Subclass)
			}

			subclass = &Subclass{Name: matches[2]}
			class.Subclass[matches[1]] = subclass
			continue
		}

		if matches := classRgx.FindStringSubmatch(scanner.Text()); matches != nil {

			class = &Class{Name: matches[2]}
			this.Classes[matches[1]] = class
		}
	}

	return scanner.Err()
}

// Load loads previously-saved USB information from disk.
func (this *Usb) Load(f string) error {

	j, err := ioutil.ReadFile(f)

	if err != nil {
		return err
	}
	if err := json.Unmarshal(j, &this); err != nil {
		return err
	}

	return nil
}

// Save saves current USB information to disk.
func (this *Usb) Save(f string) error {

	j, err := json.Marshal(this)

	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(f, j, usbMetaFileMode); err != nil {
		return err
	}

	return nil
}
