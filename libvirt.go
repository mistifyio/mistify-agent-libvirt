package libvirt

import (
	"github.com/alexzorin/libvirt-go"
	"github.com/mistifyio/mistify-agent/client"
	"github.com/mistifyio/mistify-agent/log"
	"github.com/mistifyio/mistify-agent/rpc"
	"net/http"
	"syscall"
)

var StateNames = map[int]string{
	libvirt.VIR_DOMAIN_NOSTATE:     "No State",
	libvirt.VIR_DOMAIN_RUNNING:     "Running",
	libvirt.VIR_DOMAIN_BLOCKED:     "Blocked",
	libvirt.VIR_DOMAIN_PAUSED:      "Paused",
	libvirt.VIR_DOMAIN_SHUTDOWN:    "Shutdown",
	libvirt.VIR_DOMAIN_CRASHED:     "Crashed",
	libvirt.VIR_DOMAIN_PMSUSPENDED: "Suspended",
	libvirt.VIR_DOMAIN_SHUTOFF:     "Shutoff",
}

type (
	Libvirt struct {
		uri         string
		connections chan *libvirt.VirConnection
		max         int
	}

	Domain struct {
		*libvirt.VirDomain
		State int
	}
)

func NewLibvirt(uri string, max int) (*Libvirt, error) {
	lv := &Libvirt{
		uri:         uri,
		max:         max,
		connections: make(chan *libvirt.VirConnection, max),
	}

	for i := 0; i < max; i++ {
		v, err := libvirt.NewVirConnection(uri)
		if err != nil {
			return nil, err
		}
		lv.connections <- &v
	}

	return lv, nil
}

func (lv *Libvirt) RunHTTP(port uint) error {
	server, err := rpc.NewServer(int(port))
	if err != nil {
		return err
	}

	server.RegisterService(lv)
	return server.ListenAndServe()
}

func (lv *Libvirt) getConnection() *libvirt.VirConnection {
	conn := <-lv.connections
	defer func() {
		lv.connections <- conn
	}()

	return conn
}

func (lv *Libvirt) LookupDomainByName(name string) (*libvirt.VirDomain, error) {
	conn := lv.getConnection()

	domain, err := conn.LookupDomainByName(name)
	if err != nil {
		return nil, err
	}

	return &domain, err
}

func (lv *Libvirt) NewDomain(guest *client.Guest) (*libvirt.VirDomain, error) {
	conn := lv.getConnection()

	xml := lv.DomainXML(guest)
	domain, err := conn.DomainDefineXML(xml)
	if err != nil {
		return nil, err
	}
	defer domain.Free()

	return &domain, err
}

func GetState(domain *libvirt.VirDomain) (int, error) {
	state, err := domain.GetState()
	if err != nil {
		return libvirt.VIR_DOMAIN_NOSTATE, err
	}

	return state[0], nil
}

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

func (lv *Libvirt) Restart(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	return lv.Reboot(http, request, response)
}

func (lv *Libvirt) Poweroff(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	return lv.Shutdown(http, request, response)
}

func (lv *Libvirt) Delete(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	log.Info("Libvirt.Delete %s\n", request.Guest.Id)

	domain, err := lv.LookupDomainByName(request.Guest.Id)
	defer domain.Free()
	if err != nil {
		return err
	}

	err = domain.Destroy()
	if err != nil {
		return err
	}

	*response = rpc.GuestResponse{
		Guest: request.Guest,
	}

	return nil
}

func (lv *Libvirt) Create(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	log.Info("Libvirt.Create %s\n", request.Guest.Id)

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

func (lv *Libvirt) Run(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	log.Info("Libvirt.Run %s\n", request.Guest.Id)

	return lv.DomainWrapper(func(domain *libvirt.VirDomain, state int) error {
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

func (lv *Libvirt) Reboot(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	log.Info("Libvirt.Reboot %s\n", request.Guest.Id)

	return lv.DomainWrapper(func(domain *libvirt.VirDomain, state int) error {
		return domain.Reboot(0)
	})(http, request, response)
}

func (lv *Libvirt) Shutdown(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	log.Info("Libvirt.Shutdown %s\n", request.Guest.Id)

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

func (lv *Libvirt) Status(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	log.Info("Libvirt.Status %s\n", request.Guest.Id)

	return lv.DomainWrapper(func(domain *libvirt.VirDomain, state int) error {
		// DomainWrapper gets the state already, no need to do anything here
		return nil
	})(http, request, response)
}

func (lv *Libvirt) CpuMetrics(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestMetricsResponse) error {
	var params libvirt.VirTypedParameters
	metrics := make([]*client.GuestCpuMetrics, request.Guest.Cpu)

	domain, err := lv.LookupDomainByName(request.Guest.Id)
	if err != nil {
		return err
	}
	defer domain.Free()

	nparams, err := domain.GetCPUStats(nil, 0, 0, uint32(request.Guest.Cpu), 0)
	if err != nil {
		return err
	}

	_, err = domain.GetCPUStats(&params, nparams, 0, uint32(request.Guest.Cpu), 0)
	if err != nil {
		return err
	}

	*response = rpc.GuestMetricsResponse{
		CPU: metrics,
	}
	return nil
}

func (lv *Libvirt) DiskMetrics(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestMetricsResponse) error {
	metrics := make(map[string]*client.GuestDiskMetrics, len(request.Guest.Disks))
	var params libvirt.VirTypedParameters

	domain, err := lv.LookupDomainByName(request.Guest.Id)
	if err != nil {
		return err
	}
	defer domain.Free()

	_, err = domain.BlockStatsFlags("", &params, 0, 0)
	if err != nil {
		return err
	}

	*response = rpc.GuestMetricsResponse{
		Disk: metrics,
	}
	return nil
}

func (lv *Libvirt) NicMetrics(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestMetricsResponse) error {
	metrics := make(map[string]*client.GuestNicMetrics, len(request.Guest.Nics))

	*response = rpc.GuestMetricsResponse{
		Nic: metrics,
	}
	return nil
}
