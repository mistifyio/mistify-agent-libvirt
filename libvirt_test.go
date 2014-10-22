package libvirt_test

import (
	"github.com/mistifyio/mistify-agent/rpc"
	"testing"
	"github.com/mistifyio/mistify-agent-libvirt"
	"time"
	"github.com/mistifyio/mistify-agent/client"
)

var port uint = 9001

type TestClient struct {
	rpc *rpc.Client
}

func setup(t *testing.T) *TestClient {
	lv, err := libvirt.NewLibvirt("test:///default", 1)
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

	return cli
}

func create(t *testing.T, cli *TestClient) {
	guest := client.Guest{Id: "testlibvirt", Memory: 1024}
	request := rpc.GuestRequest{Guest: &guest}
	response := rpc.GuestResponse{}

	err := cli.rpc.Do("Libvirt.Create", &request, &response)
	if err != nil {
		t.Fatalf("Error in create: %s\n", err.Error())
	}
}

func destroy(t *testing.T, cli *TestClient) {
	guest := client.Guest{Id: "testlibvirt"}
	request := rpc.GuestRequest{Guest: &guest}
	response := rpc.GuestResponse{}

	err := cli.rpc.Do("Libvirt.Delete", &request, &response)
	if err != nil {
		t.Fatalf("Error in delete: %s\n", err.Error())
	}
}

func TestCreateDestroy(t *testing.T) {
	cli := setup(t)

	create(t, cli)
	destroy(t, cli)
}
