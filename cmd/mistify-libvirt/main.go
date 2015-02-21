package main

import (
	"os"

	"github.com/mistifyio/mistify-agent-libvirt"
	"github.com/mistifyio/mistify-agent/log"
	"github.com/mistifyio/mistify-agent/rpc"
	flag "github.com/spf13/pflag"
)

type (
	Libvirt struct {
	}
)

func main() {

	var port uint
	var help bool

	flag.BoolVarP(&help, "help", "h", false, "display the help")
	flag.UintVarP(&port, "port", "p", 19999, "listen port")
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
