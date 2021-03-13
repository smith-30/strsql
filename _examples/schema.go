package schema

import "encoding/json"

type ATable struct {
	ID       string `gorm:"primary_key"`
	Name     string
	Is       bool
	NumInt   int
	NumFloat float64
	Json     json.RawMessage
	Byte     []byte
}

type BTable struct {
	ID       string `gorm:"primary_key"`
	Name     string
	Is       bool
	NumInt   int
	NumFloat float64
	Json     json.RawMessage
	Byte     []byte
}
