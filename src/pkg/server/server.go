package server

import (
	"companies/pkg/types"
	"encoding/csv"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Server struct {
	csvFile *os.File
	csvFileMutex sync.Mutex

	companiesMutex sync.RWMutex
	companies []*types.Company
	removedCompaniesCount int
	companiesByName map[string]*types.Company
	companiesByINN map[string]*types.Company
}

func New(file *os.File) (*Server, error) {
	s := &Server{
		csvFile: file,
		companiesByINN: make(map[string]*types.Company),
		companiesByName: make(map[string]*types.Company),
	}

	err := s.syncIndices()

	time.AfterFunc(time.Minute * 20, func() {
		for {
			s.maybeCleanCompanies()
			time.Sleep(time.Minute * 20)
		}
	})

	return s, err
}

func (s *Server) ListCompanies(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)

	w.Header().Set("content-type", "application/json")

	s.companiesMutex.RLock()
	companies := make([]*types.Company, 0, len(s.companies))

	for _, c := range s.companies {
		if c.Removed {
			continue
		}

		companies = append(companies, c)
	}

	encoder.Encode(companies)
	s.companiesMutex.RUnlock()
}

func (s *Server) AddCompany(w http.ResponseWriter, r *http.Request) {
	var company types.Company

	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&company)
	if err != nil {
		return
	}

	s.companiesMutex.Lock()
	c, ok := s.companiesByName[company.Name]
	if ok {
		c.Name = company.Name
		c.INN = company.INN
		c.Phone = company.Phone
		c.Address = company.Address
		c.Individual = company.Individual
	} else {
		s.companiesByName[company.Name] = &company
		s.companiesByINN[company.INN] = &company
		s.companies = append(s.companies, &company)
	}
	s.companiesMutex.Unlock()

	err = s.flushChangesToFile()
	if err != nil {
		log.Printf("failed flush change to file: %v", err)
	}
}

func (s *Server) DeleteCompany(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	id := r.Form.Get("id")
	if id == "" {
		http.Error(w, "empty id", http.StatusBadRequest)
		return
	}

	s.companiesMutex.Lock()
	if c, ok := s.companiesByName[id]; ok {
		c.Removed = true
		delete(s.companiesByName, id)
		delete(s.companiesByINN, c.INN)
		s.removedCompaniesCount += 1
	} else if c, ok := s.companiesByINN[id]; ok {
		c.Removed = true
		delete(s.companiesByINN, id)
		delete(s.companiesByName, c.Name)
		s.removedCompaniesCount += 1
	}
	s.companiesMutex.Unlock()
}

func (s *Server) readCompaniesFromFile() ([]*types.Company, error) {
	s.csvFileMutex.Lock()
	defer s.csvFileMutex.Unlock()

	reader := csv.NewReader(s.csvFile)
	var companies []*types.Company

	rec, err := reader.Read()
	for  err == nil {
		companies = append(companies, &types.Company{
			Name: rec[0],
			INN: rec[1],
			Phone: rec[2],
			Address: rec[3],
			Individual: rec[4],
		})

		rec, err = reader.Read()
	}

	if err == io.EOF {
		return companies, nil
	}

	return companies, err
}

func (s *Server) syncIndices() error {
	companies, err := s.readCompaniesFromFile()
	if err != nil {
		return err
	}

	s.companiesMutex.Lock()
	s.companies = companies
	for _, company := range companies {
		s.companiesByName[company.Name] = company
		s.companiesByINN[company.INN] = company
	}
	s.companiesMutex.Unlock()

	log.Printf("synced indices: size: %v", len(companies))

	return nil
}

// maybeCleanCompanies заново строит массив s.companies из хеш таблицы s.companiesByName
// потому что когда удаляем записи, они не сразу удаляются из этого массива,
// а отмечаются как company.Removed=true. Из-за этого массив может распухнуть со временем.
func (s *Server) maybeCleanCompanies() {
	removesCount := 0
	actualCount := 0

	s.companiesMutex.Lock()
	removesCount = s.removedCompaniesCount
	actualCount = len(s.companies)
	s.companiesMutex.Unlock()

	if removesCount * 2 < actualCount {
		return
	}

	var companies []*types.Company

	log.Printf("cleaning array: array size: %v, map size: %v", len(s.companies), len(s.companiesByName))

	s.companiesMutex.Lock()
	for _, company := range s.companiesByName {
		companies = append(companies, company)
	}
	s.companies = companies
	s.companiesMutex.Unlock()
}

func (s *Server) flushChangesToFile() error {
	var records [][]string

	s.companiesMutex.RLock()
	for _, company := range s.companiesByName {
		records = append(records, []string{
			company.Name,
			company.INN,
			company.Phone,
			company.Address,
			company.Individual,
		})
	}
	s.companiesMutex.RUnlock()

	s.csvFileMutex.Lock()
	defer s.csvFileMutex.Unlock()

	s.csvFile.Seek(io.SeekStart, 0)
	writer := csv.NewWriter(s.csvFile)
	err := writer.WriteAll(records)
	if err != nil {
		log.Printf("could not write records to file: %v", err)
		return err
	}

	err = s.csvFile.Sync()
	if err != nil {
		log.Printf("could not fsync: %v", err)
		return err
	}

	return nil
}
