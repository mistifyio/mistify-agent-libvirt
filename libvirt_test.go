package libvirt_test

import (
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/mistify-agent-libvirt"
	"github.com/mistifyio/mistify-agent/client"
	"github.com/mistifyio/mistify-agent/rpc"
	logx "github.com/mistifyio/mistify-logrus-ext"
)

type TestClient struct {
	rpc      *rpc.Client
	guest    *client.Guest
	request  *rpc.GuestRequest
	response *rpc.GuestResponse
	metrics  *rpc.GuestMetricsResponse
}

func setup(t *testing.T, url string, port uint) *TestClient {
	lv, err := libvirt.NewLibvirt(url, "mistify", 1)
	if err != nil {
		t.Fatalf("NewLibvirt failed: %s\n", err.Error())
	}

	go logx.LogReturnedErr(func() error { return lv.RunHTTP(port) }, nil, "failed to run server")
	time.Sleep(1 * time.Second)

	cli := new(TestClient)

	cli.rpc, err = rpc.NewClient(port, "")
	if err != nil {
		t.Fatalf("Can't create RPC client: %s\n", err.Error())
	}

	cli.guest = new(client.Guest)
	cli.guest.ID = "testlibvirt"
	cli.guest.Memory = 1024
	cli.guest.CPU = 1

	disk := client.Disk{
		Bus:    "sata",
		Device: "hda",
		Size:   1024,
		Source: "/dev/zvol/guests/images/testlibvirt",
	}
	cli.guest.Disks = append(cli.guest.Disks, disk)

	nic := client.Nic{
		Name:    "eth0",
		Mac:     "00:0c:29:2f:00:00",
		Network: "default",
		Device:  "vnet0",
		VLANs:   []int{1},
	}
	cli.guest.Nics = append(cli.guest.Nics, nic)

	cli.request = new(rpc.GuestRequest)
	cli.request.Guest = cli.guest

	cli.response = new(rpc.GuestResponse)

	return cli
}

func do(action string, t *testing.T, cli *TestClient, expectedState string) {
	err := cli.rpc.Do(action, cli.request, cli.response)
	if err != nil {
		t.Fatalf("Error running %s: %s\n", action, err.Error())
	}

	if expectedState != "" {
		for i := 0; i < 10; i++ {
			time.Sleep(1 * time.Second)

			err := cli.rpc.Do("Libvirt.Status", cli.request, cli.response)
			if err != nil {
				t.Fatalf("Error running Libvirt.Status after %s: %s\n", action, err.Error())
			}

			if cli.response.Guest.State == expectedState {
				t.Logf("Ran %s, state is now %s\n", action, cli.response.Guest.State)
				return
			}
		}

		t.Fatalf("After %s, expected state %s, got state %s\n", action, expectedState, cli.response.Guest.State)
	}

	t.Logf("Ran %s, state is now %s\n", action, cli.response.Guest.State)
}

func metric(action string, t *testing.T, cli *TestClient) {
	err := cli.rpc.Do(action, cli.request, cli.metrics)
	if err != nil {
		t.Fatalf("Error running %s: %s\n", action, err.Error())
	}

	t.Logf("Ran %s\n", action)
}

func TestDummy(t *testing.T) {
	cli := setup(t, "test:///default", 9001)
	cli.guest.Type = "test"

	do("Libvirt.Create", t, cli, "running")
	do("Libvirt.Shutdown", t, cli, "shutoff")
	do("Libvirt.Run", t, cli, "running")
	do("Libvirt.Reboot", t, cli, "running")
	do("Libvirt.Delete", t, cli, "")
}

func TestQemu(t *testing.T) {
	cli := setup(t, "qemu:///system", 9002)
	cli.guest.Type = "qemu"

	do("Libvirt.Create", t, cli, "running")
	do("Libvirt.Delete", t, cli, "")
}

func TestMetrics(t *testing.T) {
	cli := setup(t, "qemu:///system", 9003)
	cli.guest.Type = "qemu"

	do("Libvirt.Create", t, cli, "running")

	metric("Libvirt.CPUMetrics", t, cli)
	metric("Libvirt.DiskMetrics", t, cli)
	metric("Libvirt.NicMetrics", t, cli)

	do("Libvirt.Delete", t, cli, "")
}

func init() {
	log.SetLevel(log.FatalLevel)
}
