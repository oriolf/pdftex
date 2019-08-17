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

	bareTex := `
\documentclass[12pt]{article}

\begin{document}
    \section{Sample document 2}

	Sample content\ldots

\end{document}
`
	err = pdftex.New().Input(bareTex).Compile().Save("file2.pdf")
	if err != nil {
		log.Fatalln("Could not compile bare tex:", err)
	}
}
