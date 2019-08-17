# pdftex

A Go library for compiling LaTeX or one of its variants (TeX, XeTeX, LuaTeX...). For the library to work,
all commands (pdflatex, xelatex...) and packages (texlive...) used must be installed in the host executing
the code. It can compile either a string with the contents of the .tex file, or a Go template that when
execeuted generates a valid .tex file. The output of the process may be an io.Reader or a pdf file saved
on disk.
