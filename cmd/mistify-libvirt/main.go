package main

import (
	flag "github.com/docker/docker/pkg/mflag"
//	"github.com/mistifyio/mistify-agent-libvirt"
	"github.com/mistifyio/mistify-agent/rpc"
	"github.com/mistifyio/mistify-agent/log"
	"os"
	"fmt"
)

type (
	Libvirt struct {
	}
)

const (
	DEFAULT_PORT = 19999
	DEFAULT_LOG_LEVEL = "warning"
)

var (
	port = DEFAULT_PORT
	logLevel = DEFAULT_LOG_LEVEL
	help bool
)

func init() {
	flag.BoolVar(&help, []string{"h", "#help", "-help"}, false, "display the help")
	flag.IntVar(&port, []string{"p", "#port", "-port"}, DEFAULT_PORT, "listen port")
	flag.StringVar(&logLevel, []string{"l", "-log-level"}, DEFAULT_LOG_LEVEL, "log level: debug/info/warning/error/critical/fatal")

	flag.Parse()
}

func main() {
	if help {
		flag.PrintDefaults()
		os.Exit(0)
	}
	
	fmt.Printf("Using port %d\n", port)

	// set log level, or default
	logOk := log.SetLogLevel(logLevel)
	if logOk != nil {
		fmt.Printf("%s, defaulting to '%s'\n", logOk, DEFAULT_LOG_LEVEL)
		log.SetLogLevel(DEFAULT_LOG_LEVEL)
	}

	s, err := rpc.NewServer(port)
	if err != nil {
		log.Fatal(err)
	}
	s.RegisterService(&Libvirt{})
	log.Fatal(s.ListenAndServe())
}
