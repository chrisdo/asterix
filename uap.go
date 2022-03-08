package asterix

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

//DataItemFormat is a predefined format of a Data Item such as fixed or variable
type DataItemFormat string

//FieldFormat defines the way a field is decoded, e.g. if the value is treated as uint or as int
type FieldFormat string

//UapDictionary is a map of maps containing for each category (primary key) different versions (secondary key) of UAP
type UapDictionary map[uint8]map[string]*UAP

const (
	fixed      DataItemFormat = "fixed"
	variable   DataItemFormat = "variable"
	compound   DataItemFormat = "compound"
	explicit   DataItemFormat = "explicit"
	repetitive DataItemFormat = "repetitive"

	intType      FieldFormat = "int"
	uintType     FieldFormat = "uint"
	hexType      FieldFormat = "hex"
	icao6StrType FieldFormat = "icao6str"
	ascii        FieldFormat = "ascii"
)

//ParserConfiguration is a mappin of a asterix category to a version
type ParserConfiguration map[uint8]string

//Category holds a UAP
type Category struct {
	Uap *UAP `xml:"category"`
}

//UAP (User Application Profile) is the structure defining the content of an Asterix Record
//of a certain category and version.
type UAP struct {
	Cat        uint8          `xml:"cat,attr"`
	Version    string         `xml:"version,attr"`
	Items      []*UapDataItem `xml:"dataitem"`
	ItemsByID  map[string]*UapDataItem
	ItemsByFrn map[uint8]*UapDataItem
}

func (uap *UAP) String() string {
	return fmt.Sprintf("Category: %d, Version: %s, Items: %v", uap.Cat, uap.Version, uap.Items)
}

/*UapDataItem defines the general structure of an Item, meaning its type, length and Name*/
type UapDataItem struct {
	ID         string         `xml:"id,attr"`
	Frn        uint8          `xml:"frn,attr"`
	Format     DataItemFormat `xml:"format,attr"`
	Length     uint8          `xml:"length,attr"`
	Definition string         `xml:"definition"`
	Name       string         `xml:"name"`
	Fields     []*UapField    `xml:"field"` //this works but hav it as a slice
	SubFields  []*UapDataItem `xml:"subfield"`
}

func (uapd *UapDataItem) String() string {
	return fmt.Sprintf("%s (%s) @ FRN=%d with type=%s and fields: %v and subfields: %v", uapd.ID, uapd.Name, uapd.Frn, uapd.Format, uapd.Fields, uapd.SubFields)
}

/*UapField is the lowest structure of an Item. It contains the necessary data to decode
values based on its position and length inside the data of an Item
*/
type UapField struct {
	Octed  uint8       `xml:"octed,attr"`
	Msb    uint8       `xml:"msb,attr"`
	Lsb    uint8       `xml:"lsb,attr"`
	Name   string      `xml:"name"`
	Format FieldFormat `xml:"format"`
	Scale  string      `xml:"scale"`
}

func (uapf *UapField) String() string {
	return fmt.Sprintf("%s @ Octed %d [%d:%d] with format %s @ scale=%s", uapf.Name, uapf.Octed, uapf.Msb, uapf.Lsb, uapf.Format, uapf.Scale)
}

//decode fields, a field has a stringer, returing its name, value and raw value
//write stringer for Uap stuff

//NewDictionary reads in the xml specifications on the given folder and returns a dictionary, which
//maps a combination of category and version as key in the type of "cat:version" to a UAP
func NewDictionary(folder string) (uapDictionary UapDictionary, err error) {
	//TODO: we could also look for a settings file that contains some setup ,e.g. what files to read
	uapDictionary = make(map[uint8]map[string]*UAP, 20)
	err = filepath.Walk(folder, uapDictionary.walkFiles)

	return uapDictionary, err
}

//TODO: for a later idea of using asterix data: the tracker could br configured with a mapping of Item IDs to
//things like position, speed, heading. Then it would be easy to be independent of providing a trajectory data

//walkfiles called by walk, implementing the filepath.WalkFunc signature
func (dict UapDictionary) walkFiles(path string, info os.FileInfo, err error) error {

	if !info.IsDir() {
		fmt.Printf("Reading %s\n", info.Name())
		fp, err := filepath.Abs(path)
		file, err := os.Open(fp)
		if err != nil {
			return err
		}

		defer file.Close()

		// Read the uap file
		acat, err := ReadCategory(file)
		if err != nil {
			return err

		}
		if dict[acat.Uap.Cat] == nil {
			dict[acat.Uap.Cat] = make(map[string]*UAP)
		}
		dict[acat.Uap.Cat][acat.Uap.Version] = acat.Uap

	}
	return nil

}

//ReadCategory reads the User Application Profile from a XML file
func ReadCategory(reader io.Reader) (*Category, error) {
	var acat Category

	if err := xml.NewDecoder(reader).Decode(&acat); err != nil {
		return nil, err
	}

	fmt.Printf("read new Uap of category %d and version %s\n", acat.Uap.Cat, acat.Uap.Version)
	fmt.Printf("%+v\n", acat.Uap.Items)

	//for a faster lookup we create thosse two helpers
	acat.Uap.ItemsByID = make(map[string]*UapDataItem, len(acat.Uap.Items))
	acat.Uap.ItemsByFrn = make(map[uint8]*UapDataItem, len(acat.Uap.Items))

	for _, i := range acat.Uap.Items {
		fmt.Println(i)
		fmt.Println("")
		acat.Uap.ItemsByID[i.ID] = i
		acat.Uap.ItemsByFrn[i.Frn] = i
	}

	return &acat, nil
}
