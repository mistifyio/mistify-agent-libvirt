package libvirt

import (
	"github.com/alexzorin/libvirt-go"
	"github.com/mistifyio/mistify-agent/rpc"
	"github.com/mistifyio/mistify-agent/client"
	"syscall"
	"net/http"
	"fmt"
)

var StateNames = map[int]string{
	libvirt.VIR_DOMAIN_NOSTATE: "No State",
	libvirt.VIR_DOMAIN_RUNNING: "Running",
	libvirt.VIR_DOMAIN_BLOCKED: "Blocked",
	libvirt.VIR_DOMAIN_PAUSED: "Paused",
	libvirt.VIR_DOMAIN_SHUTDOWN: "Shutdown",
	libvirt.VIR_DOMAIN_CRASHED: "Crashed",
	libvirt.VIR_DOMAIN_PMSUSPENDED: "Suspended",
	libvirt.VIR_DOMAIN_SHUTOFF: "Shutoff",
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

	domain, err := conn.DomainDefineXML(fmt.Sprintf(`<domain type="test"><name>%s</name><memory unit="MiB">%d</memory><os><type>hvm</type></os></domain>`, guest.Id, guest.Memory))
	if err != nil {
		return nil, err
	}

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
		defer domain.Free()
		if err != nil {
			return err
		}

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
		Guest: nil,
	}

	return nil
}

func (lv *Libvirt) Create(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	domain, err := lv.NewDomain(request.Guest)
	defer domain.Free()
	if err != nil {
		return err
	}

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
	return lv.DomainWrapper(func(domain *libvirt.VirDomain, state int) error {
		return domain.Reboot(0)
	})(http, request, response)
}

func (lv *Libvirt) Shutdown(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
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
	return lv.DomainWrapper(func(domain *libvirt.VirDomain, state int) error {
		_, err := domain.GetState()
		return err
	})(http, request, response)
}

/*
func (lv *Libvirt) CpuMetrics(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
}

func (lv *Libvirt) DiskMetrics(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
}

func (lv *Libvirt) NicMetrics(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
}
*/
