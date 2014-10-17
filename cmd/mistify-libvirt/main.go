package main

import (
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/mistifyio/mistify-agent-libvirt"
	"github.com/mistifyio/mistify-agent/rpc"
	"log"
	"net/http"
	"os"
)

type (
	Libvirt struct {
	}
)

func main() {

	var port int
	var h bool

	flag.BoolVar(&h, []string{"h", "#help", "-help"}, false, "display the help")
	flag.IntVar(&port, []string{"p", "#port", "-port"}, 19999, "listen port")
	flag.Parse()

	if h {
		flag.PrintDefaults()
		os.Exit(0)
	}
	s, err := rpc.NewServer(port)
	if err != nil {
		log.Fatal(err)
	}
	s.RegisterService(&Libvirt{})
	log.Fatal(s.ListenAndServe())
}
