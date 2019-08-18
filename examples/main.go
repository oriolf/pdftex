package main

import (
	"log"

	"github.com/oriolf/pdftex"
)

func main() {
	bareTex := `
\documentclass[12pt]{article}

\begin{document}
    \section{Sample document 2}

	Sample content\ldots

\end{document}
`
	if err := pdftex.New().Input(bareTex).Compile().Save("file.pdf"); err != nil {
		log.Fatalln("Could not compile bare tex:", err)
	}

	data := []string{"a", "b", "c"}
	if err := pdftex.New().Data(data).Compile().Save("file2.pdf"); err != nil {
		log.Fatalln("Could not compile folder:", err)
	}

	if err := pdftex.New().TemplatesFolder("templates2").CopyFiles().Compile().Save("file3.pdf"); err != nil {
		log.Fatalln("Could not compile folder with files:", err)
	}
}
