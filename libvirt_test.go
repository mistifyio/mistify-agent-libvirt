package libvirt_test

import (
	"github.com/mistifyio/mistify-agent-libvirt"
	"github.com/mistifyio/mistify-agent/client"
	"github.com/mistifyio/mistify-agent/rpc"
	"testing"
	"time"
)

var port uint = 9001

type TestClient struct {
	rpc      *rpc.Client
	guest    *client.Guest
	request  *rpc.GuestRequest
	response *rpc.GuestResponse
}

func setup(t *testing.T) *TestClient {
	lv, err := libvirt.NewLibvirt("qemu:///system", 1)
	if err != nil {
		t.Fatalf("NewLibvirt failed: %s\n", err.Error())
	}

	go lv.RunHTTP(port)
	time.Sleep(1 * time.Second)

	cli := new(TestClient)

	cli.rpc, err = rpc.NewClient(port)
	if err != nil {
		t.Fatalf("Can't create RPC client: %s\n", err.Error())
	}

	cli.guest = new(client.Guest)
	cli.guest.Id = "testlibvirt"
	cli.guest.Memory = 1024
	cli.guest.Cpu = 1

	disk := client.Disk{
		Bus:    "sata",
		Device: "/dev/hda",
		Size:   1024,
	}
	cli.guest.Disks = append(cli.guest.Disks, disk)

	nic := client.Nic{
		Name: "eth0",
		Mac:  "01:23:45:67:89:0a",
	}
	cli.guest.Nics = append(cli.guest.Nics, nic)

	cli.request = new(rpc.GuestRequest)
	cli.request.Guest = cli.guest

	cli.response = new(rpc.GuestResponse)

	return cli
}

func do(action string, t *testing.T, cli *TestClient) {
	err := cli.rpc.Do(action, cli.request, cli.response)
	if err != nil {
		t.Fatalf("Error running %s: %s\n", action, err.Error())
	}
}

func TestLibvirt(t *testing.T) {
	cli := setup(t)

	do("Libvirt.Create", t, cli)
	if cli.response.Guest.State != "Running" {
		t.Fatalf("After create, guest state is %s\n", cli.response.Guest.State)
	}

	do("Libvirt.Shutdown", t, cli)
	if cli.response.Guest.State != "Shutoff" {
		t.Fatalf("After shutdown, guest state is %s\n", cli.response.Guest.State)
	}

	do("Libvirt.Run", t, cli)
	if cli.response.Guest.State != "Running" {
		t.Fatalf("After run, guest state is %s\n", cli.response.Guest.State)
	}

	do("Libvirt.Restart", t, cli)
	if cli.response.Guest.State != "Running" {
		t.Fatalf("After restart, guest state is %s\n", cli.response.Guest.State)
	}

	do("Libvirt.Delete", t, cli)
}
