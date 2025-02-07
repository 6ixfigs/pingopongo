package main

import (
	"log"
	"net/http"
	"os"

	"github.com/6ixfigs/pingypongy/internal/rest"
)

func main() {
	logfile, err := os.OpenFile("/var/log/pongo/pongo.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("error opening log file: %v", err)
	}
	defer logfile.Close()

	log.SetOutput(logfile)
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Ltime)

	s, err := rest.NewServer()
	if err != nil {
		log.Fatal("server failed to start: ", err)
	}

	s.MountRoutes()

	log.Printf("server running on port %s", s.Cfg.ServerPort)
	if err := http.ListenAndServe(":"+s.Cfg.ServerPort, s.Rtr); err != nil {
		log.Fatal("server failed to start: ", err)
	}
}
