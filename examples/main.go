package main

import (
	"fmt"
	"log"
	"sync"

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

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			filename := fmt.Sprintf("file_%d.pdf", i)
			if err := pdftex.New().TemplatesFolder("templates2").CopyFiles().Compile().Save(filename); err != nil {
				log.Fatalln("Error when compiling in parallel:", err)
			}
		}()
	}

	wg.Wait()
}
