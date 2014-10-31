package main

import (
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/mistifyio/mistify-agent-libvirt"
	"github.com/mistifyio/mistify-agent/rpc"
	"log"
	"os"
)

type (
	Libvirt struct {
	}
)

func main() {

	var port int
	var help bool

	flag.BoolVar(&help, []string{"h", "#help", "-help"}, false, "display the help")
	flag.IntVar(&port, []string{"p", "#port", "-port"}, 19999, "listen port")
	flag.Parse()

	if help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	server, err := rpc.NewServer(port)
	if err != nil {
		log.Fatal(err)
	}

	server.RegisterService(&libvirt.Libvirt{})
	log.Fatal(server.ListenAndServe())
}
