package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/mistify-agent-libvirt"
	"github.com/mistifyio/mistify-agent/rpc"
	logx "github.com/mistifyio/mistify-logrus-ext"
	flag "github.com/spf13/pflag"
)

func main() {

	var port uint

	flag.UintVarP(&port, "port", "p", 20001, "listen port")
	flag.Parse()

	if err := logx.DefaultSetup("info"); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "logx.DefaultSetup",
		}).Fatal("failed to set up logrus")
	}

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
	if err := server.RegisterService(lv); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "server.RegisterService",
		}).Fatal(err)
	}
	if err = server.ListenAndServe(); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "rpc.Server.ListenAndServe",
		}).Fatal(err)
	}
}
