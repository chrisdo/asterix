package asterix

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"time"
)

func readFspec(r *bytes.Reader) (fspec []uint8, fspecBytes []byte) {
	fspecByte := byte(1) //initialize so that first for condition is fullfilled

	currentFrn := uint8(0)

	for fspecByte&1 == 1 {
		fspecByte, _ = r.ReadByte()
		fspecBytes = append(fspecBytes, fspecByte)
		for i := byte(0x80); i > 1; i /= 2 {
			currentFrn++
			if fspecByte&i > 1 {
				fspec = append(fspec, currentFrn)
			}
		}

	}
	return
}

func (ab *Block) readRecords(uap *UAP) (numberOfRecords int) {
	var records []Record
	if ab.Data == nil && len(ab.Data) == 0 {
		return 0
	}

	//from byte number 3 (first is cat, then 2 bytes length)
	r := bytes.NewReader(ab.Data[3:ab.Length])

	for r.Len() > 0 {
		fspec, _ := readFspec(r)
		log.Printf("Record fspec: %v", fspec)

		var rec Record
		rec.Items = make(map[string]DataItem, len(fspec))
		for _, frn := range fspec {
			//look up frn for category and version and read from r the specified size of bytes
			rec.Items[uap.ItemsByFrn[frn].ID], _ = readByFormat(r, uap.ItemsByFrn[frn], frn)
		}
		for k, v := range rec.Items {
			log.Printf("Item %s: %v", k, v)
		}
		records = append(records, rec)
	}

	ab.Records = records
	return len(records)

}

func readByFormat(r *bytes.Reader, uapDataItem *UapDataItem, frn uint8) (DataItem, error) {
	switch uapDataItem.Format {

	case fixed:
		return readFixedLengthDataItem(r, uapDataItem, frn), nil
	case variable:
		return readVariableLengthDataItem(r, uapDataItem, frn), nil
	case compound:
		return readCompoundDataItem(r, uapDataItem, frn), nil
	case explicit:
		//e.g. special Purpose FIeld or extensions
		return readExplicitDataItem(r, uapDataItem, frn), nil
	case repetitive:
		return readRepetitiveLengthDataItem(r, uapDataItem, frn), nil
	default:
		//ignore
	}
	return DataItem{}, nil
}

func (d *DataItem) readByFieldType(uap *UapDataItem) {
	d.Fields = make(map[string]Field) //for each Fuield in the uap, use the bitReader and decode it according to its field type
	for _, uapField := range uap.Fields {
		if len(uapField.Format) == 0 || uapField.Format == uintType {
			d.DecodeUintField(uapField)
		} else if uapField.Format == intType {
			d.DecodeIntField(uapField)
		} else if uapField.Format == icao6StrType {
			d.DecodeIcao6StrField(uapField)
		} else if uapField.Format == hexType {
			d.DecodeHexField(uapField)
		} else if uapField.Format == ascii {
			d.DecodeASCIIField(uapField)
		}
	}
}

func readCompoundDataItem(r *bytes.Reader, di *UapDataItem, frn uint8) (dataItem DataItem) {

	//first we need to read fspec
	fspec, _ := readFspec(r)
	//i create a DataItem
	//that has a map of stringto Fields
	//from the subfields i get back a DataItem. I can add all Fields as: ItemId#subFrn#FieldName: Field
	//Uff, not so clear right now
	log.Printf("Compund Item %s (FRN: %d) fspec: %v", di.ID, frn, fspec)
	dataItem.Fields = make(map[string]Field)
	for _, subFrn := range fspec {

		subItem, err := readByFormat(r, di.SubFields[subFrn-1], subFrn)
		if err != nil {
			fmt.Println(err)
		}

		for k, v := range subItem.Fields {

			dataItem.Fields[fmt.Sprintf("%s#%d#%s", di.ID, subFrn, k)] = v
		}
	}
	return
}

func readExplicitDataItem(r *bytes.Reader, uap *UapDataItem, frn uint8) (dataItem DataItem) {
	len, err := r.ReadByte()
	if err != nil {

	}
	b := make([]byte, len-1)
	r.Read(b)
	dataItem.Data = append(b, len)
	//then fspec, and then treat like compound, OR: be able to load from uap mapping this document
	return
}

func readRepetitiveLengthDataItem(r *bytes.Reader, udi *UapDataItem, frn uint8) (dataItem DataItem) {
	//first byte: number of repetititions
	dataItem.Fields = make(map[string]Field)

	numRep, _ := r.ReadByte()
	data := make([]byte, numRep*udi.Length+1)
	data = append(data, numRep)
	for rep := 0; rep < int(numRep); rep++ {
		di := readFixedLengthDataItem(r, udi, frn)
		//somhow append the []byte from di to data
		data = append(data[:], dataItem.Data[:]...) //this i do not understand, see here: https://stackoverflow.com/questions/8461462/how-can-i-use-go-append-with-two-byte-slices-or-arrays

		for k, v := range di.Fields {
			dataItem.Fields[fmt.Sprintf("%s#%d#%s", udi.ID, rep, k)] = v
		}
	}
	return
}

func readVariableLengthDataItem(r *bytes.Reader, uap *UapDataItem, frn uint8) DataItem {
	//read first length, could be e.g. 2 bytes or just 1
	b := make([]byte, uap.Length)
	//read at least those first bytes
	r.Read(b)

	//then if LSB ==1, we can read one more byte until LSB=0
	//TODO: we need to add a maxLength property to the xml andtake this into account here to compensate for errors like cat 21 v2.1
	if b[len(b)-1]&0x01 == 1 {
		next := 1
		for next&0x01 == 1 {
			next, err := r.ReadByte()

			if err != nil {
				b = append(b, next)
			}
			if next&0x01 == 0 {
				break
			}
		}
	}

	dataItem := DataItem{Data: b}
	dataItem.readByFieldType(uap)

	log.Printf("Read Variable Length item frn %d with ID %s with length %d and data %v", frn, uap.ID, len(b), dataItem)
	return dataItem
}

func readFixedLengthDataItem(r *bytes.Reader, uap *UapDataItem, frn uint8) DataItem {
	b := make([]byte, uap.Length)
	r.Read(b)
	dataItem := DataItem{Data: b}
	dataItem.readByFieldType(uap)

	log.Printf("Read Fixed Item frn %d ID %s data %v", frn, uap.ID, dataItem)
	return dataItem
}

//NewBlock returns a new asterix.Block based on the dictionary of this data.
//A dataBLock is always of a certain category and version, therefore a ParserCOnfiguration has
//to be passed in order to determine the version to be used to parse the data
//The key of the dictionary is used to lookup the Uap version used to decode the block
func (dict UapDictionary) NewBlock(data []byte, parserConfig ParserConfiguration) (*Block, error) {
	asterixBlock := Block{Cat: data[0], Length: binary.BigEndian.Uint16(data[1:3]), ReceivedAt: time.Now()}
	asterixBlock.Data = data //[3:asterixBlock.Length]
	if _, ok := dict[asterixBlock.Cat]; !ok {
		return &asterixBlock, fmt.Errorf("Category %d not present in dictionary, can not decode", asterixBlock.Cat)
	}
	if _, ok := parserConfig[asterixBlock.Cat]; !ok {
		return &asterixBlock, fmt.Errorf("No Version configured for Category %d, can not decode", asterixBlock.Cat)
	}
	asterixBlock.readRecords(dict[asterixBlock.Cat][parserConfig[asterixBlock.Cat]])
	return &asterixBlock, nil
}

//what do I need to know about a Data Item
//FRN and ID
//Length: Fixed, Variable (fixed first part + extensionLength (can be 1, 2..)), Repetitive, Compound, special
//fixed: clear
//variable: fixed length first part (e.g. 1 or 2 byte), then extension of variable length
//Repetitive: has given length, but a multiplikator REP field as first byte, so total = REP*length
//Compund: first bytes are fspec, followed by fields with any of the above data (only fixed, variable or repetitive)
//special: used for special purpose or extension, first field is one byte length, then fspec, then fields
