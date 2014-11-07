package main

import (
	"os"

	flag "github.com/docker/docker/pkg/mflag"
	"github.com/mistifyio/mistify-agent-libvirt"
	"github.com/mistifyio/mistify-agent/log"
	"github.com/mistifyio/mistify-agent/rpc"
)

type (
	Libvirt struct {
	}
)

func main() {

	var port uint
	var help bool

	flag.BoolVar(&help, []string{"h", "#help", "-help"}, false, "display the help")
	flag.UintVar(&port, []string{"p", "#port", "-port"}, 19999, "listen port")
	flag.Parse()

	if help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	server, err := rpc.NewServer(port)
	if err != nil {
		log.Fatal(err)
	}

	lv, err := libvirt.NewLibvirt("qemu:///system", 4)
	if err != nil {
		log.Fatal(err)
	}
	server.RegisterService(lv)
	log.Fatal(server.ListenAndServe())
}
