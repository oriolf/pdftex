package pdftex

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
	"time"
)

type pdfGenerator struct {
	err             error
	command         Command
	templatesFolder string
	templateName    string
	fileSystem      fs.FS
	funcs           template.FuncMap
	data            interface{}
	template        *template.Template
	copyFiles       bool
	input           string
	output          string
}

type Command string

const (
	PDFLaTeX Command = "pdflatex"
	XeLaTeX  Command = "xelatex"
)

func New() *pdfGenerator {
	return &pdfGenerator{
		command:         PDFLaTeX,
		templatesFolder: "templates",
		templateName:    "template.tmpl",
	}
}

func (pdf *pdfGenerator) Command(c Command) *pdfGenerator {
	pdf.command = c
	return pdf
}

func (pdf *pdfGenerator) TemplatesFolder(folder string) *pdfGenerator {
	pdf.templatesFolder = folder
	return pdf
}

func (pdf *pdfGenerator) TemplateName(name string) *pdfGenerator {
	pdf.templateName = name
	return pdf
}

func (pdf *pdfGenerator) FileSystem(fsys fs.FS) *pdfGenerator {
	pdf.fileSystem = fsys
	return pdf
}

func (pdf *pdfGenerator) Funcs(funcs template.FuncMap) *pdfGenerator {
	pdf.funcs = funcs
	return pdf
}

func (pdf *pdfGenerator) Data(data interface{}) *pdfGenerator {
	pdf.data = data
	return pdf
}

func (pdf *pdfGenerator) Template(tmpl *template.Template) *pdfGenerator {
	pdf.template = tmpl
	return pdf
}

func (pdf *pdfGenerator) CopyFiles() *pdfGenerator {
	pdf.copyFiles = true
	return pdf
}

func (pdf *pdfGenerator) Input(input string) *pdfGenerator {
	pdf.input = input
	return pdf
}

func (pdf *pdfGenerator) Err() error     { return pdf.err }
func (pdf *pdfGenerator) Output() string { return pdf.output }
func (pdf *pdfGenerator) Save(filename string) error {
	if pdf.err != nil {
		return pdf.err
	}

	if pdf.output == "" {
		return errors.New("compilation has not been executed yet")
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create file %s: %w", filename, err)
	}
	defer f.Close()

	if _, err := f.Write([]byte(pdf.output)); err != nil {
		return fmt.Errorf("could not write output to file: %w", err)
	}

	return nil
}

func (pdf *pdfGenerator) Compile() *pdfGenerator {
	if pdf.err != nil {
		return pdf
	}

	if pdf.input == "" {
		pdf.input, pdf.err = pdf.compileTemplate()
	}

	pdf.output, pdf.err = pdf.compileRawInput()
	return pdf
}

func (pdf *pdfGenerator) compileRawInput() (string, error) {
	if pdf.err != nil {
		return "", pdf.err
	}

	foldername := createFolderName()
	if err := os.Mkdir(foldername, os.ModePerm); err != nil {
		return "", fmt.Errorf("could not mkdir tmp folder: %w", err)
	}

	f, err := os.Create(filepath.Join(foldername, "file.tex"))
	if err != nil {
		return "", fmt.Errorf("could not create file.tex: %w", err)
	}

	if _, err := f.Write([]byte(pdf.input)); err != nil {
		f.Close()
		return "", fmt.Errorf("could not write file.tex contents: %w", err)
	}
	f.Close()

	return pdf.executeCompilationAndClean(foldername)
}

func (pdf *pdfGenerator) executeCompilationAndClean(foldername string) (string, error) {
	if pdf.copyFiles {
		files, err := ioutil.ReadDir(pdf.templatesFolder)
		if err != nil {
			return "", fmt.Errorf("could not read templates folder: %w", err)
		}

		for _, file := range files {
			if err := exec.Command("cp", filepath.Join(pdf.templatesFolder, file.Name()), foldername).Run(); err != nil {
				return "", fmt.Errorf("could not copy templates folder contents to tmp folder: %w", err)
			}
		}
	}

	com := fmt.Sprintf("cd %s && %s file.tex", foldername, string(pdf.command))
	if err := exec.Command("bash", "-c", com).Run(); err != nil {
		return "", fmt.Errorf("could not %s file.tex: %w", pdf.command, err)
	}

	output, err := ioutil.ReadFile(filepath.Join(foldername, "file.pdf"))
	if err != nil {
		return "", fmt.Errorf("error reading file.pdf: %w", err)
	}

	if err := os.RemoveAll(foldername); err != nil {
		return "", fmt.Errorf("could not rm -r tmp folder: %w", err)
	}

	return string(output), nil
}

func (pdf *pdfGenerator) compileTemplate() (string, error) {
	if pdf.err != nil {
		return "", pdf.err
	}

	if pdf.template == nil {
		tmpl := template.New(pdf.templateName).Funcs(pdf.funcs)
		templateName := filepath.Join(pdf.templatesFolder, pdf.templateName)
		var err error
		if pdf.fileSystem != nil {
			tmpl, err = tmpl.ParseFS(pdf.fileSystem, templateName)
		} else {
			tmpl, err = tmpl.ParseFiles(templateName)
		}
		if err != nil {
			return "", fmt.Errorf("could not create template: %w", err)
		}
		pdf.template = tmpl
	}

	var buffer bytes.Buffer
	if err := pdf.template.Execute(&buffer, pdf.data); err != nil {
		return "", fmt.Errorf("could not execute template: %w", err)
	}

	return buffer.String(), nil
}

func createFolderName() string { return "tmp" + fmt.Sprintf("%d", time.Now().UnixNano()) }
func createFileName() string   { return "sourcefile" + fmt.Sprintf("%d", time.Now().UnixNano()) }
