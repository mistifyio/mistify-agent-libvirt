package libvirt_test

import (
	"github.com/mistifyio/mistify-agent/rpc"
	"testing"
	"github.com/mistifyio/mistify-agent-libvirt"
	"time"
	"github.com/mistifyio/mistify-agent/client"
)

var port uint = 9001

func StartServer(t *testing.T) {
	lv, err := libvirt.NewLibvirt("test:///default", 1)
	if err != nil {
		t.Fatalf("NewLibvirt failed: %s\n", err.Error())
	}

	go lv.RunHTTP(port)
	time.Sleep(1 * time.Second)
}

func TestCreate(t *testing.T) {
	StartServer(t)

	rpcClient, err := rpc.NewClient(port)
	if err != nil {
		t.Fatalf("Can't create RPC client: %s\n", err.Error())
	}

	guest := client.Guest{Id: "testlibvirt", Memory: 1024}
	request := rpc.GuestRequest{Guest: &guest}
	response := rpc.GuestResponse{}

	err = rpcClient.Do("Libvirt.Create", &request, &response)
	if err != nil {
		t.Fatalf("Error running: %s\n", err.Error())
	}
}