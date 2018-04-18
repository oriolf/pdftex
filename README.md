# pdftex

A Go library for compiling LaTeX and XeTeX. The library depends on pdflatex, xelatex commands and all packages included in 
the tex file to be installed in the host executing the code. It can compile either a string or a Go template. The output of 
the "Compile" functions is a single PDF file.
