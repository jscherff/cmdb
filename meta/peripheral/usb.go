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

package peripheral

import (
	`bufio`
	`encoding/json`
	`fmt`
	`io/ioutil`
	`net/http`
	`os`
	`path/filepath`
	`regexp`
	`time`
)

const (
	usbMetaDirMode = 0755
	usbMetaFileMode = 0644
	usbMetaSourceUrl = `http://www.linux-usb.org/usb.ids`
	marshalPrefix = ""
	marshalIndent = "\t"
	maximumAge = 720
	minimumAge = 1
	tagVendor = `vendor_name`
	tagProduct = `product_name`
	tagClass = `usb_class`
	tagSubClass = `usb_subclass`
	tagProtocol = `usb_protocol`
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

	// usbSubClassRgx extracts the subClass IDs and descriptions from the source data.
	subClassRgx = regexp.MustCompile(`^\t([0-9A-Fa-f]{2})\s+(.+)$`)

	// usbProtocolRgx extracts the protocol IDs and descriptions from the source data.
	protocolRgx = regexp.MustCompile(`^\t\t([0-9A-Fa-f]{2})\s+(.+)$`)
)

// Usb contains all known information USB this.Vendors, products, this.Classes,
// subthis.Classes, and protocols.
type Usb struct {
	Source string
	Updated time.Time
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

// Class contains the name of the class and mappings for each subClass.
type Class struct {
	Name string
	SubClass map[string]*SubClass
}

// String returns the name of the class.
func (this *Class) String() string {
	return this.Name
}

// SubClass contains the name of the subClass and any associated protocols.
type SubClass struct {
	Name string
	Protocol map[string]*Protocol
}

// String returns the name of the SubClass.
func (this *SubClass) String() string {
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
func NewUsb(cf string) (this *Usb, err error) {

	this = &Usb{

		Source: usbMetaSourceUrl,		// Default
		Updated: time.Now(),			// Default

		Vendors: make(map[string]*Vendor),
		Classes: make(map[string]*Class),
	}

	if err = this.Load(cf); err != nil {
		err = this.LoadUrl(usbMetaSourceUrl)
	}

	if err != nil {
		return nil, err
	}

	if time.Since(this.Updated).Hours() > maximumAge {
		this.Refresh()
	}

	if time.Since(this.Updated).Minutes() < minimumAge {

		if err := os.MkdirAll(filepath.Dir(cf), usbMetaDirMode); err != nil {
			return this, err
		}

		if err := this.Save(cf); err != nil {
			return this, err
		}
	}

	return this, nil
}

// LastUpdate returns the date the metadata was last updated from source.
func (this *Usb) LastUpdate() (time.Time) {
	return this.Updated
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

// GetSubClass returns the USB subClass associated with a subClass ID.
func (this *Class) GetSubClass(sid string) (*SubClass, error) {

	if s, ok := this.SubClass[sid]; !ok {
		return nil, fmt.Errorf(`subClass %q not found`, sid)
	} else {
		return s, nil
	}
}

// GetProtocol returns the USB protocol associated with a protocol ID.
func (this *SubClass) GetProtocol(pid string) (*Protocol, error) {

	if p, ok := this.Protocol[pid]; !ok {
		return nil, fmt.Errorf(`protocol %q not found`, pid)
	} else {
		return p, nil
	}
}

// LoadUrl reads vendor, product, class, subClass, and protocol information
// line-by-line from an io.Reader source and populates data structures for
// use by the application.
func (this *Usb) LoadUrl(url string) error {

	vendors := make(map[string]*Vendor)
	classes := make(map[string]*Class)

	var (
		vendor   *Vendor
		class    *Class
		subClass *SubClass
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
			vendors[matches[1]] = vendor
			continue
		}

		if matches := protocolRgx.FindStringSubmatch(scanner.Text()); matches != nil {

			if subClass == nil {
				return fmt.Errorf(`protocol with no subClass`)
			}
			if subClass.Protocol == nil {
				subClass.Protocol = make(map[string]*Protocol)
			}

			subClass.Protocol[matches[1]] = &Protocol{Name: matches[2]}
			continue
		}

		if matches := subClassRgx.FindStringSubmatch(scanner.Text()); matches != nil {

			if class == nil {
				return fmt.Errorf(`subClass with no class`)
			}
			if class.SubClass == nil {
				class.SubClass = make(map[string]*SubClass)
			}

			subClass = &SubClass{Name: matches[2]}
			class.SubClass[matches[1]] = subClass
			continue
		}

		if matches := classRgx.FindStringSubmatch(scanner.Text()); matches != nil {

			class = &Class{Name: matches[2]}
			classes[matches[1]] = class
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	this.Source = url
	this.Updated = time.Now()
	this.Vendors = vendors
	this.Classes = classes

	return nil
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

	j, err := json.MarshalIndent(this, marshalPrefix, marshalIndent)

	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(f, j, usbMetaFileMode); err != nil {
		return err
	}

	return nil
}

// Refresh refreshes the USB information from the source URL.
func (this *Usb) Refresh() error {
	return this.LoadUrl(this.Source)
}
