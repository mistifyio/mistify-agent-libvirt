# libvirt

[![libvirt](https://godoc.org/github.com/mistifyio/mistify-agent-libvirt?status.png)](https://godoc.org/github.com/mistifyio/mistify-agent-libvirt)

Package libvirt is a mistify subagent that manages guests with libvirt, exposed
via JSON-RPC over HTTP.

### HTTP API Endpoint

    /_mistify_RPC_
    	* GET - Run a specified method

### Request Structure

    {
        "method": "RPC_METHOD",
        "params": [
            DATA_STRUCT
        ],
        "id": 0
    }

Where RPC_METHOD is the desired method and DATA_STRUCTURE is one of the request
structs defined in http://godoc.org/github.com/mistifyio/mistify-agent/rpc .

### Response Structure

    {
        "result": {
            KEY: RESPONSE_STRUCT
        },
        "error": null,
        "id": 0
    }

Where KEY is a string (e.g. "snapshot") and DATA is one of the response structs
defined in http://godoc.org/github.com/mistifyio/mistify-agent/rpc .

### RPC Methods

    CreateGuest
    Create
    Delete

    Restart
    Reboot
    Poweroff
    Shutdown
    Run

    Status
    CpuMetrics
    DiskMetrics
    NicMetrics

See the godocs and function signatures for each method's purpose and expected
request/response structs.

## Usage

```go
var StateNames = map[int]string{
	libvirt.VIR_DOMAIN_NOSTATE:     "unknown",
	libvirt.VIR_DOMAIN_RUNNING:     "running",
	libvirt.VIR_DOMAIN_BLOCKED:     "blocked",
	libvirt.VIR_DOMAIN_PAUSED:      "paused",
	libvirt.VIR_DOMAIN_SHUTDOWN:    "shutdown",
	libvirt.VIR_DOMAIN_CRASHED:     "crashed",
	libvirt.VIR_DOMAIN_PMSUSPENDED: "suspended",
	libvirt.VIR_DOMAIN_SHUTOFF:     "shutoff",
}
```
StateNames maps libvirt domain running states to common name strings

#### func  GetState

```go
func GetState(domain *libvirt.VirDomain) (int, error)
```
GetState looks up the running state of a libvirt domain

#### type Connection

```go
type Connection struct {
	*libvirt.VirConnection
}
```

Connection is a libvirt connection

#### func (*Connection) Release

```go
func (c *Connection) Release() error
```
Release closes a connection and returns one limiter to the pool

#### type Device

```go
type Device struct {
	Disks      []Disk      `xml:"disk,omitempty" json:"disks,omitempty"`
	Interfaces []Interface `xml:"interface,omitempty" json:"interfaces,omitempty"`
	Graphics   Graphics    `xml:"graphics" json:"graphics"`
}
```

Device http://libvirt.org/formatdomain.html#elementsDevices

#### type Disk

```go
type Disk struct {
	Type   string     `xml:"type,attr"  json:"type"`
	Device string     `xml:"device,attr" json:"device"`
	Driver DiskDriver `xml:"driver"  json:"driver"`
	Source DiskSource `xml:"source" json:"source"`
	Target DiskTarget `xml:"target" json:"target"`
}
```

Disk http://libvirt.org/formatdomain.html#elementsDisks

#### type DiskDriver

```go
type DiskDriver struct {
	Name string `xml:"name,attr" json:"name"`
	Type string `xml:"type,attr" json:"type"`
}
```

DiskDriver http://libvirt.org/formatdomain.html#elementsDisks

#### type DiskSource

```go
type DiskSource struct {
	File   string `xml:"file,omitempty,attr" json:"file,omitempty"`
	Device string `xml:"dev,omitempty,attr" json:"device,omitempty"`
}
```

DiskSource http://libvirt.org/formatdomain.html#elementsDisks

#### type DiskTarget

```go
type DiskTarget struct {
	Device string `xml:"dev,attr" json:"device"`
	Bus    string `xml:"bus,attr" json:"bus"`
}
```

DiskTarget http://libvirt.org/formatdomain.html#elementsDisks

#### type Domain

```go
type Domain struct {
	*libvirt.VirDomain
	State int
}
```

Domain is a libvirt domain with running state

#### type FilterRef

```go
type FilterRef struct {
	Filter     string               `xml:"filter,attr"  json:"filter"`
	Parameters []FilterRefParameter `xml:"parameter" json:"parameters"`
}
```

FilterRef http://libvirt.org/formatnwfilter.html#nwfconcepts

#### type FilterRefParameter

```go
type FilterRefParameter struct {
	Name  string `xml:"name,attr" json:"name"`
	Value string `xml:"value,attr" json:"value"`
}
```

FilterRefParameter http://libvirt.org/formatnwfilter.html#nwfconcepts

#### type Graphics

```go
type Graphics struct {
	Type string `xml:"type,attr,omitempty" json:"type,omitempty"`
	Port string `xml:"port,attr,omitempty" json:"port,omitempty"`
}
```

Graphics http://libvirt.org/formatdomain.html#elementsGraphics

#### type Interface

```go
type Interface struct {
	Type      string          `xml:"type,attr"  json:"type"`
	Source    InterfaceSource `xml:"source,omitempty" json:"source,omitempty"`
	Mac       InterfaceMac    `xml:"mac,omitempty" json:"mac,omitempty"`
	Model     InterfaceModel  `xml:"model,omitempty" json:"model,omitempty"`
	FilterRef FilterRef       `xml:"filterref,omitempty" json:"filterref,omitempty"`
	Target    InterfaceTarget `xml:"target,omitempty" json:"target,omitempty"`
	Alias     InterfaceAlias  `xml:"alias,omitempty" json:"alias,omitempty"`
}
```

Interface http://libvirt.org/formatdomain.html#elementsNICS

#### type InterfaceAlias

```go
type InterfaceAlias struct {
	Name string `xml:"name,omitempty,attr" json:"name"`
}
```

InterfaceAlias http://libvirt.org/formatdomain.html#elementsNICS

#### type InterfaceMac

```go
type InterfaceMac struct {
	// MAC address
	Address string `xml:"address,attr" json:"address"`
}
```

InterfaceMac http://libvirt.org/formatdomain.html#elementsNICS

#### type InterfaceModel

```go
type InterfaceModel struct {
	Type string `xml:"type,omitempty,attr" json:"type,omitempty"`
}
```

InterfaceModel http://libvirt.org/formatdomain.html#elementsNICSModel

#### type InterfaceSource

```go
type InterfaceSource struct {
	Network string `xml:"network,omitempty,attr" json:"network,omitempty"`
	Bridge  string `xml:"bridge,omitempty,attr" json:"bridge,omitempty"`
}
```

InterfaceSource http://libvirt.org/formatdomain.html#elementsNICS

#### type InterfaceTarget

```go
type InterfaceTarget struct {
	// Underlying network device
	Device string `xml:"dev,omitempty,attr" json:"device"`
}
```

InterfaceTarget http://libvirt.org/formatdomain.html#elementsNICS

#### type Libvirt

```go
type Libvirt struct {
}
```

Libvirt is the main struct for interacting with libvirt

#### func  NewLibvirt

```go
func NewLibvirt(uri string, max int) (*Libvirt, error)
```
NewLibvirt creates a new Libvirt object and initializes the connection limit

#### func (*Libvirt) CpuMetrics

```go
func (lv *Libvirt) CpuMetrics(r *http.Request, request *rpc.GuestMetricsRequest, response *rpc.GuestMetricsResponse) error
```
CpuMetrics looks up the cpu metrics for a libvirt domain for a guest

#### func (*Libvirt) Create

```go
func (lv *Libvirt) Create(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error
```
Create creates a new libvirt domain for a guest

#### func (*Libvirt) CreateGuest

```go
func (lv *Libvirt) CreateGuest(r *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error
```
CreateGuest creates disks and defines a new domain for a guest

#### func (*Libvirt) Delete

```go
func (lv *Libvirt) Delete(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error
```
Delete completely removes a libvirt domain for a guest

#### func (*Libvirt) DiskMetrics

```go
func (lv *Libvirt) DiskMetrics(r *http.Request, request *rpc.GuestMetricsRequest, response *rpc.GuestMetricsResponse) error
```
DiskMetrics looks up the disk metrics for a libvirt domain for a guest

#### func (*Libvirt) DomainWrapper

```go
func (lv *Libvirt) DomainWrapper(fn func(*libvirt.VirDomain, int) error) func(*http.Request, *rpc.GuestRequest, *rpc.GuestResponse) error
```
DomainWrapper looks up a libvirt domain and state for a request guest, runs a
function on it, and updates the guest for the response

#### func (*Libvirt) DomainXML

```go
func (lv *Libvirt) DomainXML(guest *client.Guest) (string, error)
```
DomainXML populates a libvirt domain xml template with guest properties

#### func (*Libvirt) LookupDomainByName

```go
func (lv *Libvirt) LookupDomainByName(name string) (*libvirt.VirDomain, error)
```
LookupDomainByName retrieves a libvirt domain based on a name string

#### func (*Libvirt) NewDomain

```go
func (lv *Libvirt) NewDomain(guest *client.Guest) (*libvirt.VirDomain, error)
```
NewDomain creates a new libvirt domain from a guest

#### func (*Libvirt) NicMetrics

```go
func (lv *Libvirt) NicMetrics(r *http.Request, request *rpc.GuestMetricsRequest, response *rpc.GuestMetricsResponse) error
```
NicMetrics looks up the nic metrics for a libvirt domain for a guest

#### func (*Libvirt) Poweroff

```go
func (lv *Libvirt) Poweroff(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error
```
Poweroff destroys a libvirt domain for a guest. Effectively pulls the (virtual)
power cord.
https://libvirt.org/html/libvirt-libvirt-domain.html#virDomainDestroy

#### func (*Libvirt) Reboot

```go
func (lv *Libvirt) Reboot(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error
```
Reboot reboots a libvirt domain for a guest

#### func (*Libvirt) Restart

```go
func (lv *Libvirt) Restart(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error
```
Restart reboots a libvirt domain for a guest. Internally calls Reboot.

#### func (*Libvirt) Run

```go
func (lv *Libvirt) Run(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error
```
Run creates or resumes a libvirt domain for a guest

#### func (*Libvirt) RunHTTP

```go
func (lv *Libvirt) RunHTTP(port uint) error
```
RunHTTP runs the HTTP server

#### func (*Libvirt) Shutdown

```go
func (lv *Libvirt) Shutdown(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error
```
Shutdown requests a libvirt domain for a guest to cleanly shutdown.
https://libvirt.org/html/libvirt-libvirt-domain.html#virDomainShutdown

#### func (*Libvirt) Status

```go
func (lv *Libvirt) Status(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error
```
Status looks up the running status of a libvirt domain for a guest

#### type Metadata

```go
type Metadata struct {
	Device       MetadataDevice `xml:"http://mistify.io/xml/device/1 device"`
	UserMetadata UserMetadata   `xml:"http://mistify.io/xml/user_metadata/1 user_metadata"`
}
```

Metadata holds all metadata

#### type MetadataDevice

```go
type MetadataDevice struct {
	XMLName xml.Name       `xml:"http://mistify.io/xml/device/1 device"`
	Disks   []MetadataDisk `xml:"http://mistify.io/xml/device/1 disk"`
}
```

MetadataDevice is metadata about disks

#### type MetadataDisk

```go
type MetadataDisk struct {
	XMLName xml.Name `xml:"http://mistify.io/xml/device/1 disk"`
	Device  string   `xml:"device,attr"`
	Image   string   `xml:"image,attr"`
	Volume  string   `xml:"volume,attr"`
}
```

MetadataDisk is metadata about a disk

#### type Os

```go
type Os struct {
	Type OsType `xml:"type,omitempty" json:"type,omitempty"`
	Boot OsBoot `xml:"boot,omitempty" json:"boot,omitempty"`
}
```

Os http://libvirt.org/formatdomain.html#elementsOS

#### type OsBoot

```go
type OsBoot struct {
	Dev string `xml:"dev,attr,omitempty" json:"dev,omitempty"`
}
```

OsBoot http://libvirt.org/formatdomain.html#elementsOS

#### type OsType

```go
type OsType struct {
	Type    string `xml:",chardata" json:"type,omitempty"`
	Arch    string `xml:"arch,attr,omitempty" json:"arch,omitempty"`
	Machine string `xml:"machine,attr,omitempty" json:"machine,omitempty"`
}
```

OsType http://libvirt.org/formatdomain.html#elementsOS

#### type UserMetadata

```go
type UserMetadata struct {
	XMLName    xml.Name                `xml:"http://mistify.io/xml/user_metadata/1 user_metadata"`
	Parameters []UserMetadataParameter `xml:"http://mistify.io/xml/user_metadata/1 parameter"`
}
```

UserMetadata is metadata about a use

#### type UserMetadataParameter

```go
type UserMetadataParameter struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}
```

UserMetadataParameter is an item of metadata about a user

#### type VirDomain

```go
type VirDomain struct {
	*libvirt.VirDomain `xml:"-" json:"-"`
	XMLName            struct{} `xml:"domain" json:"-"`
	Type               string   `xml:"type,attr" json:"type"`
	UUID               string   `xml:"uuid" json:"uuid"`
	Name               string   `xml:"name" json:"name"`
	Memory             int      `xml:"memory" json:"memory"`
	VCPU               int      `xml:"vcpu" json:"vcpu"`
	Devices            Device   `xml:"devices,omitempty" json:"devices"`
	Os                 Os       `xml:"os,omitempty" json:"os,omitempty"`
	State              string   `xml:"-" json:"state"`
	Metadata           Metadata `xml:"metadata"`
}
```

VirDomain http://libvirt.org/formatdomain.html#elementsMetadata

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
