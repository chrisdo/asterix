package asterix

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"strings"

	"github.com/Knetic/govaluate"
)

/**  The translation table for icao Strings. */
var translationTable = [...]string{"", "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P",
	"Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "", "", "", "", "",
	/* SPACE! */ " ", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "",
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "", "", "", "", "", ""}

//DecodeIcao6StrField decodes the field as an icao string with 6 bits per character
func (item *DataItem) DecodeIcao6StrField(uapField *UapField) {
	var rawField uint64

	data := make([]byte, 8-len(item.Data))
	data = append(data, item.Data...)
	for _, b := range data {
		rawField = (rawField << 8) | uint64(b)
	}
	bitmask := generateBitMask(uapField.Msb - uapField.Lsb + 1)
	rawField = rawField >> (uapField.Lsb - 1)
	rawField &= bitmask

	icao := make([]string, 0, 8)
	for i := uapField.Msb; i >= 6; i -= 6 {

		c := ((rawField >> (i - 6)) & 0x3f)
		icao = append(icao, translationTable[c])
	}

	item.Fields[uapField.Name] = Field{strings.Join(icao, "")}
}

//DecodeASCIIField decodes the given field as Ascii string
func (item *DataItem) DecodeASCIIField(uapField *UapField) {

	item.Fields[uapField.Name] = Field{string(item.Data)}
}

//DecodeUintField decodes the field as an unsigned integer with potential scale
func (item *DataItem) DecodeUintField(uapField *UapField) {

	//ifthe field is repetitive, we need to skip the first byte (repetitions)
	var rawField uint64

	data := make([]byte, 8-len(item.Data))
	data = append(data, item.Data...)
	for _, b := range data {
		rawField = (rawField << 8) | uint64(b)
	}
	bitmask := generateBitMask(uapField.Msb - uapField.Lsb + 1)
	rawField = rawField >> (uapField.Lsb - 1)
	rawField &= bitmask

	field := Field{float64(rawField) * uapField.getScale()}
	item.Fields[uapField.Name] = field

}

//DecodeIntField decodes the given field as a signed integer with potential scale
func (item *DataItem) DecodeIntField(uapField *UapField) {

	//ifthe field is repetitive, we need to skip the first byte (repetitions)
	var rawField uint64
	//TODO, there is a bug here for special cases, e.g. position msb is 64, lsb=33.. someting wierd is with negative numbers
	data := make([]byte, 8-len(item.Data))
	data = append(data, item.Data...)
	for _, b := range data {
		rawField = (rawField << 8) | uint64(b)
	}
	bitmask := generateBitMask(uapField.Msb - uapField.Lsb + 1)
	rawField = rawField >> (uapField.Lsb - 1)
	rawField &= bitmask
	var field Field
	msb := rawField >> (uapField.Msb - 1)

	if msb == 0 {
		field = Field{(float64(rawField) * uapField.getScale())}
	} else {
		rawFieldI := rawField - 1
		rawFieldI ^= bitmask
		field = Field{float64(-1.0 * float64(rawFieldI) * uapField.getScale())}

	}
	item.Fields[uapField.Name] = field

}

//DecodeHexField decodes teh given field as a hex string, e.g. for Mode S Addresses
func (item *DataItem) DecodeHexField(uapField *UapField) {
	var rawField uint64

	data := make([]byte, 8-len(item.Data))
	data = append(data, item.Data...)
	for _, b := range data {
		rawField = (rawField << 8) | uint64(b)
	}
	bitmask := generateBitMask(uapField.Msb - uapField.Lsb + 1)
	rawField = rawField >> (uapField.Lsb - 1)
	rawField &= bitmask

	bytesliye := make([]byte, len(data))
	binary.BigEndian.PutUint64(bytesliye, rawField)
	l := math.Ceil(float64((uapField.Msb - uapField.Lsb + 1) / 8))
	item.Fields[uapField.Name] = Field{hex.EncodeToString(bytesliye)[16-uint(l)*2:]}

	//theroertically, this would work IF we could be sure the full item data is hex, but it can be its not..
	//item.Fields[uapField.Name] = Field{hex.EncodeToString(item.Data)}
}

func generateBitMask(n uint8) uint64 {
	return (1 << (n)) - 1
}

func (uapField *UapField) getScale() (scale float64) {
	scale = 1
	if len(uapField.Scale) != 0 {
		expression, err := govaluate.NewEvaluableExpression(uapField.Scale)
		if err != nil {
			if err != nil {
				fmt.Println(err)
			}
		}
		scales, err := expression.Evaluate(nil)
		if err != nil {
			fmt.Println(err)
		}
		scale = scales.(float64)
	}
	return
}
