package libvirt

import (
	"github.com/alexzorin/libvirt-go"
	"github.com/mistifyio/mistify-agent/rpc"
	"github.com/mistifyio/mistify-agent/client"
	"syscall"
	"net/http"
	"runtime"
	"fmt"
)

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

func newDomain(vDom *libvirt.VirDomain) (*Domain, error) {
	domain := &Domain{
		VirDomain: vDom,
	}
	runtime.SetFinalizer(domain, func(domain *Domain) {
		domain.Free()
	})

	state, err := vDom.GetState()
	if err != nil {
		return nil, err
	}

	domain.State = state[0]
	return domain, nil
}

func (d *Domain) Free() {
	if d.VirDomain != nil {
		d.VirDomain.Free()
		d.VirDomain = nil
	}
}

func (lv *Libvirt) getConnection() *libvirt.VirConnection {
	conn := <-lv.connections
	defer func() {
		lv.connections <- conn
	}()

	return conn
}

func (lv *Libvirt) LookupDomainByName(name string) (*Domain, error) {
	conn := lv.getConnection()

	vDom, err := conn.LookupDomainByName(name)
	if err != nil {
		return nil, err
	}

	return newDomain(&vDom)
}

func (lv *Libvirt) NewDomain(guest *client.Guest) (*Domain, error) {
	conn := lv.getConnection()

	vDom, err := conn.DomainDefineXML(fmt.Sprintf(`<domain type="test"><name>%s</name><memory unit="MiB">%d</memory><os><type>hvm</type></os></domain>`, guest.Id, guest.Memory))
	if err != nil {
		return nil, err
	}

	return newDomain(&vDom)
}

func (lv *Libvirt) DomainWrapper(fn func(*Domain) error) func(*http.Request, *rpc.GuestRequest, *rpc.GuestResponse) error {
	return func(r *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
		if request.Guest == nil || request.Guest.Id == "" {
			return syscall.EINVAL
		}

		domain, err := lv.LookupDomainByName(request.Guest.Id)
		if err != nil {
			return err
		}

		err = fn(domain)
		if err != nil {
			return err
		}

		*response = rpc.GuestResponse{
			Guest: request.Guest,
		}
		return nil
	}
}

func (lv *Libvirt) Create(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	domain, err := lv.NewDomain(request.Guest)
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

	return nil
}

func (lv *Libvirt) Reboot(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	return lv.DomainWrapper(func(domain *Domain) error {
		return domain.Reboot(0)
	})(http, request, response)
}

func (lv *Libvirt) Run(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	return lv.DomainWrapper(func(domain *Domain) error {

		switch domain.State {

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

func (lv *Libvirt) Shutdown(http *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	return lv.DomainWrapper(func(domain *Domain) error {

		switch domain.State {
		case libvirt.VIR_DOMAIN_SHUTDOWN, libvirt.VIR_DOMAIN_SHUTOFF:
			// nothing to do

		default:
			return domain.Shutdown()
		}

		return nil
	})(http, request, response)

}
