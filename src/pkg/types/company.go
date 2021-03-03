package types

import "hash/crc32"

type Company struct {
	Name       string `json:"name"`
	INN        string `json:"inn"`
	Phone      string `json:"phone"`
	Address    string `json:"address"`
	Individual string `json:"individual"`

	// Meta
	Removed bool `json:"-"`
}

func (c *Company) Hash() uint32 {
	payload := c.Name + c.INN + c.Phone + c.Address + c.Individual

	return crc32.ChecksumIEEE([]byte(payload))
}
