package main

import (
	"companies/pkg/client"
	"flag"
	"log"
	"time"
)

var (
	dsn = flag.String("dsn", "postgresql://docker:password@localhost/docker?sslmode=disable", "database source")
	serverAddr = flag.String("serverAddr", "http://localhost:8080", "server address")
	updateInterval = flag.Duration("updateInterval", time.Minute, "update interval")

)

func main() {
	c, err := client.New(*dsn, *serverAddr, *updateInterval)
	if err != nil {
		log.Fatalf("failed to create a client: %v", err)
	}

	err = c.Run()
	if err != nil {
		log.Fatal(err)
	}
}
