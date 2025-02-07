package main

import (
	"log"
	"net/http"

	"github.com/6ixfigs/pingypongy/internal/rest"
)

func main() {
	s, err := rest.NewServer()
	if err != nil {
		log.Fatal("Server failed to start: ", err)
	}

	s.MountRoutes()

	log.Printf("Server running on port %s", s.Cfg.ServerPort)
	if err := http.ListenAndServe(":"+s.Cfg.ServerPort, s.Rtr); err != nil {
		log.Fatal("Server failed to start: ", err)
	}
}
