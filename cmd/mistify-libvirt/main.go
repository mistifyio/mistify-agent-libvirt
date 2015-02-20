package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/mistifyio/mistify-agent-libvirt"
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
