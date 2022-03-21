package asterix

import (
	"encoding/hex"
	"fmt"
	"io"
	"text/template"
)

//ToFormattedString writes a formatted output to the given writer using the given template
func (ab *Block) ToFormattedString(w io.Writer, templateFile string) {
	funcMap := template.FuncMap{
		"ToHex": hex.EncodeToString,
	}

	tmpl, err := template.New("block.tmpl").Funcs(funcMap).ParseFiles(templateFile)
	if err != nil {
		fmt.Printf("An error occured while reading template: %v", err)
	}

	err = tmpl.Execute(w, ab)
	if err != nil {
		fmt.Printf("An error occured while teplanting: %v", err)
	}

}
