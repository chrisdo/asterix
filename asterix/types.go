package asterix

import (
	"time"
)

//A Block is the main structure of ASTERIX data
//A block always contains a category field as well as its length
//followed by one or more Asterix Records
type Block struct {
	ReceivedAt time.Time
	Cat        uint8
	Length     uint16
	Data       []byte
	Records    []Record
}

//Record contains a set of DataItems
type Record struct {
	Length uint16
	Items  map[string]DataItem
}

//DataItem holds its []byte content and its Decoded Fields
type DataItem struct {
	Data   []byte           //populated when reading buffer
	Fields map[string]Field //map FieldName to Value
}

//A Field contains a decoded value
type Field struct {
	Value interface{}
}
