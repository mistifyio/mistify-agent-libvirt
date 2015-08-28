package libvirt

import (
	"fmt"
	"net/http"
	"syscall"

	"encoding/xml"

	log "github.com/Sirupsen/logrus"
	"github.com/alexzorin/libvirt-go"
	"github.com/mistifyio/mistify-agent/client"
	"github.com/mistifyio/mistify-agent/rpc"
)

// StateNames maps libvirt domain running states to common name strings
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

type (
	// Connection is a libvirt connection
	Connection struct {
		*libvirt.VirConnection
		lv *Libvirt
	}

	// Libvirt is the main struct for interacting with libvirt
	Libvirt struct {
		uri         string
		connections chan struct{}
		max         int
		zpool       string
	}

	// Domain is a libvirt domain with running state
	Domain struct {
		*libvirt.VirDomain
		State int
	}

	// DiskDriver http://libvirt.org/formatdomain.html#elementsDisks
	DiskDriver struct {
		Name string `xml:"name,attr" json:"name"`
		Type string `xml:"type,attr" json:"type"`
	}

	// DiskSource http://libvirt.org/formatdomain.html#elementsDisks
	DiskSource struct {
		File   string `xml:"file,omitempty,attr" json:"file,omitempty"`
		Device string `xml:"dev,omitempty,attr" json:"device,omitempty"`
	}

	// DiskTarget http://libvirt.org/formatdomain.html#elementsDisks
	DiskTarget struct {
		Device string `xml:"dev,attr" json:"device"`
		Bus    string `xml:"bus,attr" json:"bus"`
	}

	// Disk http://libvirt.org/formatdomain.html#elementsDisks
	Disk struct {
		Type   string     `xml:"type,attr"  json:"type"`
		Device string     `xml:"device,attr" json:"device"`
		Driver DiskDriver `xml:"driver"  json:"driver"`
		Source DiskSource `xml:"source" json:"source"`
		Target DiskTarget `xml:"target" json:"target"`
	}

	// InterfaceSource  http://libvirt.org/formatdomain.html#elementsNICS
	InterfaceSource struct {
		Network string `xml:"network,omitempty,attr" json:"network,omitempty"`
		Bridge  string `xml:"bridge,omitempty,attr" json:"bridge,omitempty"`
	}

	// InterfaceMac http://libvirt.org/formatdomain.html#elementsNICS
	InterfaceMac struct {
		// MAC address
		Address string `xml:"address,attr" json:"address"`
	}

	// InterfaceModel http://libvirt.org/formatdomain.html#elementsNICSModel
	InterfaceModel struct {
		Type string `xml:"type,omitempty,attr" json:"type,omitempty"`
	}

	// InterfaceTarget http://libvirt.org/formatdomain.html#elementsNICS
	InterfaceTarget struct {
		// Underlying network device
		Device string `xml:"dev,omitempty,attr" json:"device"`
	}

	// InterfaceAlias http://libvirt.org/formatdomain.html#elementsNICS
	InterfaceAlias struct {
		Name string `xml:"name,omitempty,attr" json:"name"`
	}

	// FilterRefParameter http://libvirt.org/formatnwfilter.html#nwfconcepts
	FilterRefParameter struct {
		Name  string `xml:"name,attr" json:"name"`
		Value string `xml:"value,attr" json:"value"`
	}

	// FilterRef http://libvirt.org/formatnwfilter.html#nwfconcepts
	FilterRef struct {
		Filter     string               `xml:"filter,attr"  json:"filter"`
		Parameters []FilterRefParameter `xml:"parameter" json:"parameters"`
	}

	// Interface http://libvirt.org/formatdomain.html#elementsNICS
	Interface struct {
		Type      string          `xml:"type,attr"  json:"type"`
		Source    InterfaceSource `xml:"source,omitempty" json:"source,omitempty"`
		Mac       InterfaceMac    `xml:"mac,omitempty" json:"mac,omitempty"`
		Model     InterfaceModel  `xml:"model,omitempty" json:"model,omitempty"`
		FilterRef FilterRef       `xml:"filterref,omitempty" json:"filterref,omitempty"`
		Target    InterfaceTarget `xml:"target,omitempty" json:"target,omitempty"`
		Alias     InterfaceAlias  `xml:"alias,omitempty" json:"alias,omitempty"`
	}

	// Device http://libvirt.org/formatdomain.html#elementsDevices
	Device struct {
		Disks      []Disk      `xml:"disk,omitempty" json:"disks,omitempty"`
		Interfaces []Interface `xml:"interface,omitempty" json:"interfaces,omitempty"`
		Graphics   Graphics    `xml:"graphics" json:"graphics"`
	}

	// OsType http://libvirt.org/formatdomain.html#elementsOS
	OsType struct {
		Type    string `xml:",chardata" json:"type,omitempty"`
		Arch    string `xml:"arch,attr,omitempty" json:"arch,omitempty"`
		Machine string `xml:"machine,attr,omitempty" json:"machine,omitempty"`
	}

	// OsBoot http://libvirt.org/formatdomain.html#elementsOS
	OsBoot struct {
		Dev string `xml:"dev,attr,omitempty" json:"dev,omitempty"`
	}

	// Os http://libvirt.org/formatdomain.html#elementsOS
	Os struct {
		Type OsType `xml:"type,omitempty" json:"type,omitempty"`
		Boot OsBoot `xml:"boot,omitempty" json:"boot,omitempty"`
	}

	// Graphics http://libvirt.org/formatdomain.html#elementsGraphics
	Graphics struct {
		Type string `xml:"type,attr,omitempty" json:"type,omitempty"`
		Port string `xml:"port,attr,omitempty" json:"port,omitempty"`
	}

	// MetadataDisk is metadata about a disk
	MetadataDisk struct {
		XMLName xml.Name `xml:"http://mistify.io/xml/device/1 disk"`
		Device  string   `xml:"device,attr"`
		Image   string   `xml:"image,attr"`
		Volume  string   `xml:"volume,attr"`
	}

	// MetadataDevice is metadata about disks
	MetadataDevice struct {
		XMLName xml.Name       `xml:"http://mistify.io/xml/device/1 device"`
		Disks   []MetadataDisk `xml:"http://mistify.io/xml/device/1 disk"`
	}

	// UserMetadataParameter is an item of metadata about a user
	UserMetadataParameter struct {
		Name  string `xml:"name,attr"`
		Value string `xml:"value,attr"`
	}

	// UserMetadata is metadata about a use
	UserMetadata struct {
		XMLName    xml.Name                `xml:"http://mistify.io/xml/user_metadata/1 user_metadata"`
		Parameters []UserMetadataParameter `xml:"http://mistify.io/xml/user_metadata/1 parameter"`
	}

	// Metadata holds all metadata
	Metadata struct {
		Device       MetadataDevice `xml:"http://mistify.io/xml/device/1 device"`
		UserMetadata UserMetadata   `xml:"http://mistify.io/xml/user_metadata/1 user_metadata"`
	}

	// VirDomain http://libvirt.org/formatdomain.html#elementsMetadata
	VirDomain struct {
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
)

// NewLibvirt creates a new Libvirt object and initializes the connection limit
func NewLibvirt(uri string, zpool string, max int) (*Libvirt, error) {
	lv := &Libvirt{
		uri:         uri,
		max:         max,
		connections: make(chan struct{}, max),
		zpool:       zpool,
	}

	for i := 0; i < max; i++ {
		lv.connections <- struct{}{}
	}

	return lv, nil
}

// Release closes a connection and returns one limiter to the pool
func (c *Connection) Release() error {
	c.lv.connections <- struct{}{}
	_, err := c.CloseConnection()
	if err != nil {
		return err
	}
	return nil
}

// RunHTTP runs the HTTP server
func (lv *Libvirt) RunHTTP(port uint) error {
	server, err := rpc.NewServer(port)
	if err != nil {
		return err
	}

	if err := server.RegisterService(lv); err != nil {
		return err
	}
	return server.ListenAndServe()
}

func (lv *Libvirt) getConnection() (*Connection, error) {
	<-lv.connections
	virConn, err := libvirt.NewVirConnection(lv.uri)
	if err != nil {
		lv.connections <- struct{}{}
		return nil, err
	}

	conn := &Connection{
		VirConnection: &virConn,
		lv:            lv,
	}
	return conn, nil
}

// LookupDomainByName retrieves a libvirt domain based on a name string
func (lv *Libvirt) LookupDomainByName(name string) (*libvirt.VirDomain, error) {
	conn, err := lv.getConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	domain, err := conn.LookupDomainByName(name)
	if err != nil {
		return nil, err
	}

	return &domain, err
}

// LookupNetworkByName retrieves a libvirt domain based on a name string
func (lv *Libvirt) LookupNetworkByName(name string) (*libvirt.VirNetwork, error) {
	conn, err := lv.getConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	network, err := conn.LookupNetworkByName(name)
	if err != nil {
		return nil, err
	}

	return &network, err
}

// NewDomain creates a new libvirt domain from a guest
func (lv *Libvirt) NewDomain(guest *client.Guest) (*libvirt.VirDomain, error) {
	conn, err := lv.getConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	xml, err := lv.DomainXML(guest)
	if err != nil {
		return nil, err
	}

	domain, err := conn.DomainDefineXML(xml)
	if err != nil {
		return nil, err
	}

	return &domain, err
}

// GetState looks up the running state of a libvirt domain
func GetState(domain *libvirt.VirDomain) (int, error) {
	state, err := domain.GetState()
	if err != nil {
		return libvirt.VIR_DOMAIN_NOSTATE, err
	}

	return state[0], nil
}

// DomainWrapper looks up a libvirt domain and state for a request guest, runs a
// function on it, and updates the guest for the response
func (lv *Libvirt) DomainWrapper(fn func(*libvirt.VirDomain, int) error) func(*http.Request, *rpc.GuestRequest, *rpc.GuestResponse) error {
	return func(r *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
		if request.Guest == nil || request.Guest.Id == "" {
			return syscall.EINVAL
		}

		domain, err := lv.LookupDomainByName(request.Guest.Id)
		if err != nil {
			return err
		}
		defer domain.Free()

		state, err := GetState(domain)
		if err != nil {
			return err
		}

		err = fn(domain, state)
		if err != nil {
			return err
		}

		*response = rpc.GuestResponse{
			Guest: request.Guest,
		}

		state, err = GetState(domain)
		if err != nil {
			return err
		}

		response.Guest.State = StateNames[state]

		return nil
	}
}

// Guest Actions

// Restart reboots a libvirt domain for a guest. Internally calls Reboot.
func (lv *Libvirt) Restart(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	return lv.Reboot(http, request, response)
}

// Poweroff destroys a libvirt domain for a guest. Effectively pulls the (virtual) power cord.
// https://libvirt.org/html/libvirt-libvirt-domain.html#virDomainDestroy
func (lv *Libvirt) Poweroff(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	log.WithFields(log.Fields{
		"guest": request.Guest.Id,
	}).Info("Libvirt.Poweroff")

	return lv.DomainWrapper(func(domain *libvirt.VirDomain, state int) error {
		return domain.Destroy()
	})(http, request, response)
}

// Delete completely removes a libvirt domain for a guest
func (lv *Libvirt) Delete(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	log.WithFields(log.Fields{
		"guest": request.Guest.Id,
	}).Info("Libvirt.Delete")

	domain, err := lv.LookupDomainByName(request.Guest.Id)
	if err != nil {
		return err
	}
	defer domain.Free()

	state, err := GetState(domain)
	if err != nil {
		return err
	}

	if state == libvirt.VIR_DOMAIN_RUNNING || state == libvirt.VIR_DOMAIN_PAUSED {
		err = domain.Destroy()
		if err != nil {
			return err
		}
	}

	err = domain.Undefine()
	if err != nil {
		return err
	}

	for _, nic := range request.Guest.Nics {
		network, err := lv.LookupNetworkByName(nic.Mac)
		if err != nil {
			return err
		}

		if err = network.Destroy(); err != nil {
			return err
		}
		if err = network.Undefine(); err != nil {
			return err
		}
	}

	*response = rpc.GuestResponse{
		Guest: request.Guest,
	}

	return nil
}

// Create creates a new libvirt domain for a guest
func (lv *Libvirt) Create(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	log.WithFields(log.Fields{
		"guest": request.Guest.Id,
	}).Info("Libvirt.Create")

	domain, err := lv.NewDomain(request.Guest)
	if err != nil {
		return err
	}
	defer domain.Free()

	err = domain.Create()
	if err != nil {
		return err
	}

	*response = rpc.GuestResponse{
		Guest: request.Guest,
	}

	state, err := GetState(domain)
	if err != nil {
		return err
	}

	response.Guest.State = StateNames[state]

	return nil
}

// Run creates or resumes a libvirt domain for a guest
func (lv *Libvirt) Run(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	log.WithFields(log.Fields{
		"guest": request.Guest.Id,
	}).Info("Libvirt.Run")

	return lv.DomainWrapper(func(domain *libvirt.VirDomain, state int) error {

		xmldesc, err := domain.GetXMLDesc(0)
		if err != nil {
			return err
		}

		v := VirDomain{}
		err = xml.Unmarshal([]byte(xmldesc), &v)
		if err != nil {
			return err
		}

		for i := range request.Guest.Nics {
			nic := &request.Guest.Nics[i]
			nic.Device = v.Devices.Interfaces[i].Target.Device
			nic.Name = v.Devices.Interfaces[i].Alias.Name
		}

		switch state {

		case libvirt.VIR_DOMAIN_RUNNING:
			// nothing to do

		case libvirt.VIR_DOMAIN_SHUTDOWN, libvirt.VIR_DOMAIN_SHUTOFF, libvirt.VIR_DOMAIN_BLOCKED, libvirt.VIR_DOMAIN_NOSTATE:
			return domain.Create()

		case libvirt.VIR_DOMAIN_PAUSED, libvirt.VIR_DOMAIN_PMSUSPENDED:
			return domain.Resume()
		}

		return nil
	})(http, request, response)
}

// Reboot reboots a libvirt domain for a guest
func (lv *Libvirt) Reboot(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	log.WithFields(log.Fields{
		"guest": request.Guest.Id,
	}).Info("Libvirt.Reboot")

	return lv.DomainWrapper(func(domain *libvirt.VirDomain, state int) error {
		return domain.Reboot(0)
	})(http, request, response)
}

// Shutdown requests a libvirt domain for a guest to cleanly shutdown.
// https://libvirt.org/html/libvirt-libvirt-domain.html#virDomainShutdown
func (lv *Libvirt) Shutdown(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	log.WithFields(log.Fields{
		"guest": request.Guest.Id,
	}).Info("Libvirt.Shutdown")

	return lv.DomainWrapper(func(domain *libvirt.VirDomain, state int) error {

		switch state {
		case libvirt.VIR_DOMAIN_SHUTDOWN, libvirt.VIR_DOMAIN_SHUTOFF:
			// nothing to do

		default:
			return domain.Shutdown()
		}

		return nil
	})(http, request, response)
}

// Status looks up the running status of a libvirt domain for a guest
func (lv *Libvirt) Status(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	log.WithFields(log.Fields{
		"guest": request.Guest.Id,
	}).Info("Libvirt.Status")

	return lv.DomainWrapper(func(domain *libvirt.VirDomain, state int) error {
		// DomainWrapper gets the state already, no need to do anything here
		return nil
	})(http, request, response)
}

// CpuMetrics looks up the cpu metrics for a libvirt domain for a guest
func (lv *Libvirt) CpuMetrics(r *http.Request, request *rpc.GuestMetricsRequest, response *rpc.GuestMetricsResponse) error {

	domain, err := lv.LookupDomainByName(request.Guest.Id)
	if err != nil {
		return err
	}
	defer domain.Free()

	metrics := make([]*client.GuestCPUMetrics, 0, request.Guest.Cpu)

	// see virsh-domain.c in libvirt as this does not make sense without
	// seeing it in context there
	maxID, err := domain.GetCPUStats(nil, 0, 0, 0, 0)
	if err != nil {
		return err
	}

	nparams, err := domain.GetCPUStats(nil, 0, 0, 1, 0)
	if err != nil {
		return err
	}

	for i := 0; i < maxID; i++ {
		params := libvirt.VirTypedParameters{}

		_, err := domain.GetCPUStats(&params, nparams, i, 1, 0)
		if err != nil {
			return err
		}
		c := client.GuestCPUMetrics{}
		for _, p := range params {

			switch p.Name {
			case "cpu_time":
				c.CPUTime = float64(p.Value.(uint64)) / 1000000000
			case "vcpu_time":
				c.VCPUTime = float64(p.Value.(uint64)) / 1000000000
			}
		}
		metrics = append(metrics, &c)
	}

	*response = rpc.GuestMetricsResponse{
		CPU:  metrics,
		Type: "cpu",
	}

	return nil
}

// DiskMetrics looks up the disk metrics for a libvirt domain for a guest
func (lv *Libvirt) DiskMetrics(r *http.Request, request *rpc.GuestMetricsRequest, response *rpc.GuestMetricsResponse) error {

	domain, err := lv.LookupDomainByName(request.Guest.Id)
	if err != nil {
		return err
	}
	defer domain.Free()

	metrics := make(map[string]*client.GuestDiskMetrics)

	for _, disk := range request.Guest.Disks {
		nparams, err := domain.BlockStatsFlags(disk.Device, nil, 0, 0)
		if err != nil {
			return err
		}

		params := libvirt.VirTypedParameters{}
		nparams, err = domain.BlockStatsFlags(disk.Device, &params, nparams, 0)
		if err != nil {
			return err
		}

		m := &client.GuestDiskMetrics{Disk: disk.Device}
		for _, p := range params {
			switch p.Name {
			case "rd_operations":
				m.ReadOps = p.Value.(int64)
			case "rd_bytes":
				m.ReadBytes = p.Value.(int64)
			case "rd_total_times":
				m.ReadTime = float64(p.Value.(int64)) / 1000000000
			case "wr_operations":
				m.WriteOps = p.Value.(int64)
			case "wr_bytes":
				m.WriteBytes = p.Value.(int64)
			case "wr_total_times":
				m.WriteTime = float64(p.Value.(int64)) / 1000000000
			case "flush_operations":
				m.FlushOps = p.Value.(int64)
			case "flush_total_times":
				m.FlushTime = float64(p.Value.(int64)) / 1000000000
			}
		}
		metrics[disk.Device] = m
	}

	*response = rpc.GuestMetricsResponse{
		Disk: metrics,
		Type: "disk",
	}
	return nil
}

// NicMetrics looks up the nic metrics for a libvirt domain for a guest
func (lv *Libvirt) NicMetrics(r *http.Request, request *rpc.GuestMetricsRequest, response *rpc.GuestMetricsResponse) error {
	domain, err := lv.LookupDomainByName(request.Guest.Id)
	if err != nil {
		return err
	}
	defer domain.Free()

	metrics := make(map[string]*client.GuestNicMetrics)

	for _, nic := range request.Guest.Nics {
		m, err := domain.InterfaceStats(nic.Device)
		if err != nil {
			return err
		}

		metrics[nic.Name] = &client.GuestNicMetrics{
			Name:      nic.Name,
			RxBytes:   m.RxBytes,
			RxPackets: m.RxPackets,
			RxErrs:    m.RxErrs,
			RxDrop:    m.RxDrop,
			TxBytes:   m.TxBytes,
			TxPackets: m.TxPackets,
			TxErrs:    m.TxErrs,
			TxDrop:    m.TxDrop,
		}
	}

	*response = rpc.GuestMetricsResponse{
		Nic:  metrics,
		Type: "nic",
	}
	return nil
}

// CreateGuest creates disks and defines a new domain and network for a guest
func (lv *Libvirt) CreateGuest(r *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	conn, err := lv.getConnection()
	if err != nil {
		return err
	}
	defer conn.Release()

	guest := request.Guest

	if guest.Type == "" {
		guest.Type = "kvm"
	}

	dev := 'a'
	for i := range guest.Disks {
		disk := &guest.Disks[i]
		// TODO: do we want to support non-virtio??
		disk.Bus = "virtio"
		disk.Device = fmt.Sprintf("vd%c", dev)
		dev++
	}

	for _, nic := range guest.Nics {
		n, err := lv.NetworkXML(nic)
		if err != nil {
			return err
		}

		network, err := conn.NetworkDefineXML(n)
		if err != nil {
			return err
		}
		defer network.Free()

		if err = network.SetAutostart(true); err != nil {
			return err
		}
		if err = network.Create(); err != nil {
			return err
		}
	}

	x, err := lv.DomainXML(guest)
	if err != nil {
		return err
	}

	domain, err := conn.DomainDefineXML(x)
	if err != nil {
		return err
	}
	defer domain.Free()

	*response = rpc.GuestResponse{
		Guest: guest,
	}

	return err
}
