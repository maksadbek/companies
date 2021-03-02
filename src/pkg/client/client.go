package client

import (
	"companies/pkg/types"
	"database/sql"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/lib/pq"
)

type Client struct {
	httpClient *http.Client
	updateInterval time.Duration
	db *sql.DB
	serverAddr string

	companies map[string]*types.Company
}

func New(dsn, serverAddr string, dur time.Duration) (*Client, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Printf("failed to initialize database connection: %v", err)
	}

	client := Client{
		httpClient: &http.Client{
			Timeout: time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   time.Second,
					KeepAlive: 2 * dur,
				}).DialContext,
				IdleConnTimeout:       2 * dur,
				TLSHandshakeTimeout:   time.Second,
				ExpectContinueTimeout: time.Second,
				MaxIdleConns:          10,
				MaxIdleConnsPerHost:   10,
			},
		},
		updateInterval: dur,
		db: db,
		serverAddr: serverAddr,
		companies: make(map[string]*types.Company),
	}

	err = client.initializeCompanies()
	if err != nil {
		log.Printf("failed to initialize companies: %v", err)
		return nil, err
	}

	log.Printf("initialized companies from database, size: %v", len(client.companies))

	return &client, nil
}

func (c *Client) initializeCompanies() error {
	rows, err := c.db.Query("select name, inn, phone, address, individual from companies")
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var company types.Company
		err = rows.Scan(&company.Name, &company.INN, &company.Phone, &company.Address, &company.Individual)
		if err != nil {
			log.Printf("failed to scan company from database: %v", err)
			continue
		}

		c.companies[company.Name] = &company
	}

	return rows.Err()
}

func (c *Client) Run() error {
	for {
		err := c.updateDatabase()
		if err != nil {
			log.Printf("failed to update database: %v", err)
		} else {
			log.Print("successfully updated database")
		}

		time.Sleep(c.updateInterval)
	}
}

func (c *Client) updateDatabase() error {
	companies, err := c.getCompanyList()
	if err != nil {
		return err
	}

	add, del := c.diff(companies)
	var delNames []string

	log.Printf("updating database: actual: %v, add: %v, del: %v", len(c.companies), len(add), len(del))

	for _, d := range del {
		delNames = append(delNames, d.Name)
	}

	tx, err := c.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			log.Printf("panic occured, rollback: %v", err)
			tx.Rollback()
		}
	} ()

	addQuery := `insert into companies(name, inn, phone, address, individual) values ($1, $2, $3, $4, $5)`
	for _, company := range add {
		_, err = tx.Exec(addQuery,
			company.Name,
			company.INN,
			company.Phone,
			company.Address,
			company.Individual,
		)

		if err != nil {
			tx.Rollback()
		}
	}

	delQuery := `delete from companies where name = any($1)`
	_, err = tx.Exec(delQuery, pq.Array(delNames))
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (c *Client) getCompanyList() ([]*types.Company, error) {
	resp, err := c.httpClient.Get(c.serverAddr + "/")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var companies []*types.Company

	err = json.NewDecoder(resp.Body).Decode(&companies)
	return companies, err
}

func (c *Client) diff(companies []*types.Company) ([]*types.Company, []*types.Company) {
	var add, del []*types.Company

	var companyMap = make(map[string]*types.Company)

	for _, company := range companies {
		companyMap[company.Name] = company

		if _, ok := c.companies[company.Name]; !ok {
			add = append(add, company)
			c.companies[company.Name] = company
		}
	}

	for _, company := range c.companies {
		if _, ok := companyMap[company.Name]; !ok {
			del = append(del, company)
			delete(c.companies, company.Name)
		}
	}

	return add, del
}
