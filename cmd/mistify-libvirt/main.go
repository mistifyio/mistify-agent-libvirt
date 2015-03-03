package main

import (
	"github.com/mistifyio/mistify-agent-libvirt"
	"github.com/mistifyio/mistify-agent/rpc"
	flag "github.com/spf13/pflag"
)

type (
	Libvirt struct {
	}
)

func main() {

	var port uint

	flag.UintVarP(&port, "port", "p", 19999, "listen port")
	flag.Parse()

	log.SetFormatter(&log.JSONFormatter{})

	server, err := rpc.NewServer(port)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "rpc.NewServer",
		}).Fatal(err)
	}

	lv, err := libvirt.NewLibvirt("qemu:///system", 4)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "libvirt.NewLibvirt",
		}).Fatal(err)
	}
	server.RegisterService(lv)
	if err = server.ListenAndServe(); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "rpc.Server.ListenAndServe",
		}).Fatal(err)
	}
}
