package main

import (
	"io"
	"log"
	"os"
	"text/template"

	"github.com/oriolf/pdftex"
)

func main() {
	tmpl := template.Must(template.New("template.tmpl").ParseFiles("templates/template.tmpl"))
	r, err := pdftex.CompileLatexFolder("templates", tmpl, []string{"a", "b", "c"})
	if err != nil {
		log.Fatalln("Could not compile:", err)
	}

	f, _ := os.Create("file.pdf")
	defer f.Close()
	io.Copy(f, r)
}
