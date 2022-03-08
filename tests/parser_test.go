package asterix

import (
	"fmt"
	"github.com/chrisdo/asterix"
	"testing"
)

func TestDecodePosition(t *testing.T) {
	//do a couple of Integer field tests
	/*
	    * -180.0,90.0 - FE00000001000000
	   0.0,0.0 - 0000000000000000
	   -180.0,-90.0 - FE000000FF000000
	   -50.1235,19.1235 - FF716D2F0036654E
	   90.0,90.0 - 0100000001000000
	   180.0,1.0 - 020000000002D82D
	   180.0,-90.0 - 02000000FF000000
	   180.0,90.0 - 0200000001000000
	*/

	type testData struct {
		lat  float64
		lon  float64
		data []byte
	}
	var testSet = make([]testData, 0, 8)
	testSet = append(testSet, testData{-180.0, 90.0, []byte{
		0xFE, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00}})
	testSet = append(testSet, testData{0.0, 0.0, []byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}})
	testSet = append(testSet, testData{-180.0, -90.0, []byte{
		0xFE, 0x00, 0x00, 0x00, 0xFF, 0x00, 0x00, 0x00}})
	testSet = append(testSet, testData{-50.1235, 19.1235, []byte{
		0xFF, 0x71, 0x6D, 0x2F, 0x00, 0x36, 0x65, 0x4E}})
	testSet = append(testSet, testData{19.1235, 19.1235, []byte{
		0x00, 0x36, 0x65, 0x4E, 0x00, 0x36, 0x65, 0x4E}})
	testSet = append(testSet, testData{90.0, 90.0, []byte{
		0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00}})
	testSet = append(testSet, testData{180.0, 1.0, []byte{
		0x02, 0x00, 0x00, 0x00, 0x00, 0x02, 0xD8, 0x2D}})
	testSet = append(testSet, testData{180.0, -90.0, []byte{
		0x02, 0x00, 0x00, 0x00, 0xFF, 0x00, 0x00, 0x00}})
	testSet = append(testSet, testData{180.0, 90.0, []byte{
		0x02, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00}})

	var item = asterix.UapDataItem{}
	item.Fields = make([]*asterix.UapField, 2)

	var field = asterix.UapField{}
	field.Name = "Latitude"
	field.Format = "int"
	field.Lsb = 33
	field.Msb = 64
	field.Scale = "180/33554432"
	field.Octed = 1
	item.Fields[0] = &field
	var fieldLon = asterix.UapField{}
	fieldLon.Name = "Longitude"
	fieldLon.Format = "int"
	fieldLon.Lsb = 1
	fieldLon.Msb = 32
	fieldLon.Scale = "180/33554432"
	fieldLon.Octed = 1
	item.Fields[1] = &fieldLon

	for _, v := range testSet {
		var dataItem = asterix.DataItem{Data: v.data}
		dataItem.Fields = make(map[string]asterix.Field)

		dataItem.DecodeIntField(&field)
		dataItem.DecodeIntField(&fieldLon)

		if (len(dataItem.Fields)) != 2 {
			t.Errorf("Field was not added%s", "")
		}

		if dataItem.Fields[field.Name].Value != v.lat {
			t.Errorf("Wrong Latitude Field Value, was %f, but wanted %f", dataItem.Fields[field.Name].Value, v.lat)
		}
		if dataItem.Fields[fieldLon.Name].Value != v.lon {
			t.Errorf("Wrong Longitude Field Value, was %f, but wanted %f", dataItem.Fields[fieldLon.Name].Value, v.lon)
		}
		fmt.Printf("%v\n", dataItem)

	}

}

func TestDecodeIntField(t *testing.T) {

	dataItem := asterix.DataItem{Data: []byte{
		0x00, 0xff}}
	item := asterix.UapDataItem{}
	dataItem.Fields = make(map[string]asterix.Field)

	item.Fields = make([]*asterix.UapField, 1)

	field := asterix.UapField{}
	field.Name = "testField"
	field.Format = "int"
	field.Lsb = 1
	field.Msb = 8
	field.Scale = "0.5"
	field.Octed = 1
	item.Fields[0] = &field

	dataItem.DecodeIntField(&field)

	if (len(dataItem.Fields)) != 1 {
		t.Errorf("Field was not added%s", "")
	}
	if dataItem.Fields[field.Name].Value != -0.5 {
		t.Errorf("Wrong DECODING INT Field Value, was %s, but wanted -0.5", dataItem.Fields[field.Name].Value)
	}
}

func TestDecodeUIntField(t *testing.T) {

	dataItem := asterix.DataItem{Data: []byte{
		0x00, 0x01}}
	item := asterix.UapDataItem{}
	dataItem.Fields = make(map[string]asterix.Field)

	item.Fields = make([]*asterix.UapField, 1)

	field := asterix.UapField{}
	field.Name = "testField"
	field.Format = "uint"
	field.Lsb = 1
	field.Msb = 8
	field.Scale = "0.5"
	field.Octed = 1
	item.Fields[0] = &field

	dataItem.DecodeUintField(&field)
	if dataItem.Fields[field.Name].Value != 0.5 {
		t.Errorf("Wrong Decoding UINT Field Value, was %s, but wanted 0.5", dataItem.Fields[field.Name].Value)
	}
}
func TestDecodeIcao6StrField(t *testing.T) {

	//callsign: 202CC371C32CE0 // shall be  K      L      M      1      0      2      3     _
	var result = "KLM1023 "
	dataItem := asterix.DataItem{Data: []byte{
		0x2c, 0xc3, 0x71, 0xc3, 0x2c, 0xe0}}
	item := asterix.UapDataItem{}
	dataItem.Fields = make(map[string]asterix.Field)

	item.Fields = make([]*asterix.UapField, 1)

	field := asterix.UapField{}
	field.Name = "testField"
	field.Format = "icao6str"
	field.Lsb = 1
	field.Msb = 48
	field.Octed = 1
	item.Fields[0] = &field

	dataItem.DecodeIcao6StrField(&field)
	if dataItem.Fields[field.Name].Value != result {
		t.Errorf("Wrong Decoding ICAO 6 STR Field Value, was %s, but wanted %s", dataItem.Fields[field.Name].Value, result)
	}

}

func TestDecodeAsciiField(t *testing.T) {

	var result = "B737"
	dataItem := asterix.DataItem{Data: []byte{
		0x42, 0x37, 0x33, 0x37}}
	item := asterix.UapDataItem{}
	dataItem.Fields = make(map[string]asterix.Field)

	item.Fields = make([]*asterix.UapField, 1)

	field := asterix.UapField{}
	field.Name = "testField"
	field.Format = "ascii"
	field.Lsb = 1
	field.Msb = 32
	field.Octed = 1
	item.Fields[0] = &field

	dataItem.DecodeASCIIField(&field)
	if dataItem.Fields[field.Name].Value != result {
		t.Errorf("Wrong Decoding ASCII Field Value, was %s, but wanted %s", dataItem.Fields[field.Name].Value, result)
	}
}

func TestDecodeAsciiFieldAgain(t *testing.T) {

	var result = "1234567"
	dataItem := asterix.DataItem{Data: []byte{
		0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37}}
	item := asterix.UapDataItem{}
	dataItem.Fields = make(map[string]asterix.Field)

	item.Fields = make([]*asterix.UapField, 1)

	field := asterix.UapField{}
	field.Name = "testField"
	field.Format = "ascii"
	field.Lsb = 1
	field.Msb = 56
	field.Octed = 1
	item.Fields[0] = &field

	dataItem.DecodeASCIIField(&field)
	if dataItem.Fields[field.Name].Value != result {
		t.Errorf("Wrong Decoding ASCII Field Value, was %s, but wanted %s", dataItem.Fields[field.Name].Value, result)
	}
}
func TestDecodeHexField(t *testing.T) {

	dataItem := asterix.DataItem{Data: []byte{
		0xc0, 0xff, 0xee}}
	item := asterix.UapDataItem{}
	dataItem.Fields = make(map[string]asterix.Field)

	item.Fields = make([]*asterix.UapField, 1)

	field := asterix.UapField{}
	field.Name = "testField"
	field.Format = "hex"
	field.Lsb = 1
	field.Msb = 24
	field.Octed = 1
	item.Fields[0] = &field

	dataItem.DecodeHexField(&field)

	if (len(dataItem.Fields)) != 1 {
		t.Errorf("Field was not added%s", "")
	}
	if dataItem.Fields[field.Name].Value != "c0ffee" {
		t.Errorf("Wrong DECODING HEX Field Value, was %s, but wanted c0ffee", dataItem.Fields[field.Name].Value)
	}
}
