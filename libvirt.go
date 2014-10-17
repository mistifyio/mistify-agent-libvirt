package libvirt

import (
	"github.com/alexzorin/libvirt-go"
	"github.com/mistifyio/mistify-agent/rpc"
	"syscall"
)

var NotFound = errors.New("not found")

const (
	EAGAIN = syscall.EAGAIN
	EEXIST = syscall.EEXIST
	ENOSPC = syscall.ENOSPC
	EINVAL = syscall.EINVAL
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

func (lv *Libvirt) RunHTTP(port int) error {
	s, _ := rpc.NewServer(port)
	s.RegisterService(lv)
	return s.ListenAndServe()
}

func newDomain(vDom *libvirt.VirDomain) *Domain {
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

	domain.State = state
	return domain
}

func (d *Domain) Free() {
	if d.VirDomain != nil {
		d.VirDomain.Free()
		d.VirDomain = nil
	}
}

// LookupDomainByName will return a Domain with the given name
func (lv *Libvirt) LookupDomainByName(name string) (*Domain, error) {
	v := <-lv.connections
	defer func() {
		lv.connections <- v
	}()
	vDom, err := v.LookupDomainByName(name)

	if err != nil {
		if strings.Contains(err.Error(), "Domain not found:") {
			err = NotFound
		}
		return nil, err
	}

	domain := newDomain(vDom)

	return domain, nil
}

func (lv *Libvirt) DomainWrapper(fn func(*Domain) error) func(*http.Request, *rpc.Request, *rpc.Response) error {
	return func(r *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
		if request.Guest == nil || request.Guest.Id == "" {
			return EINVAL
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
			Guest: &request.Guest,
		}
		return nil
	}
}

func (lv *Libvirt) Reboot(r *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	return lv.DomainWrapper(func(domain *Domain) error {
		return domain.Reboot(0)
	})
}

func (lv *Libvirt) Run(r *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	return lv.DomainWrapper(func(domain *Domain) error {

		switch domian.State {

		case libvirt.VIR_DOMAIN_RUNNING:
			// nothing to do

		case libvirt.VIR_DOMAIN_SHUTDOWN, libvirt.VIR_DOMAIN_SHUTOFF, libvirt.VIR_DOMAIN_BLOCKED, libvirt.VIR_DOMAIN_NOSTATE:
			return domain.Create()

		case libvirt.VIR_DOMAIN_PAUSED, libvirt.VIR_DOMAIN_PMSUSPENDED:
			return domain.Resume()
		}

		return nil
	})

}

func (lv *Libvirt) Shutdown(r *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	return lv.DomainWrapper(func(domain *Domain) error {

		switch domain.State {
		case libvirt.VIR_DOMAIN_SHUTDOWN, libvirt.VIR_DOMAIN_SHUTOFF:
			// nothing to do

		default:
			return domain.Shutdown()
		}

		return nil
	})

}
