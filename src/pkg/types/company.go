package types

type Company struct {
	Name string `json:"name"`
	INN string `json:"inn"`
	Phone string `json:"phone"`
	Address string `json:"address"`
	Individual string `json:"individual"`

	// Meta
	Removed bool `json:"-"`
}



