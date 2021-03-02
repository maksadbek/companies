package main

import (
	"companies/pkg/server"
	"flag"
	"log"
	"net/http"
	"os"
)

var (
	addr = flag.String("addr", ":8080", "address")
	file = flag.String("file", "testdata/companies.csv", "file with companies")
)

func main() {
	csvFile, err := os.OpenFile(*file, os.O_RDWR, 0644)
	if err != nil {
		log.Fatalf("failed to open csv file %s: %s", *file, err.Error())
	}

	defer csvFile.Close()

	s, err := server.New(csvFile)
	if err != nil {
		log.Fatalf("failed to create a server: %s", err.Error())
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.ListCompanies)
	mux.HandleFunc("/add", s.AddCompany)
	mux.HandleFunc("/delete", s.DeleteCompany)

	http.ListenAndServe(*addr, mux)
}

