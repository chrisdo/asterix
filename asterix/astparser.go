package asterix

import (
	"bufio"
	"fmt"
)

type Edition struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
}

type AsterixDef struct {
	Category int     `json:"number"`
	Name     string  `json:"title"`
	Version  Edition `json:"edition"`
	Date     string  `json:"date"`
	Preamble string  `json:"preamble"`
}

type UapCroatia struct {
	Items map[int]string `json:"items"`
	Type  string         `json:"type"`
}

type Definition struct {
	general   AsterixDef
	uap       UapCroatia
	catalogue Catalogue
}

type Catalogue map[string]CatalogueItem

type CatalogueItem struct {
}

//ParseAsterixDef is the starting point to read asteris defintions in json format from
//https://zoranbosnjak.github.io/asterix-specs/index.html
//THIS IS NOT WORKING YET
func ParseAsterixDef(s *bufio.Scanner) (AsterixDef, error) {
	asterixDef := AsterixDef{}
	for s.Scan() {
		fmt.Println(s.Text())
	}
	return asterixDef, nil
}

func ParseItemsDef(s *bufio.Scanner) {

}
