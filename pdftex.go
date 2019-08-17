package pdftex

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/pkg/errors"
)

type pdfGenerator struct {
	err             error
	command         command
	templatesFolder string
	templateName    string
	funcs           template.FuncMap
	data            interface{}
	template        *template.Template
	input           string
	output          string
}

type command string

const (
	PDFLaTeX command = "pdflatex"
	XeLaTeX  command = "xelatex"
)

func New() *pdfGenerator {
	return &pdfGenerator{
		command:         PDFLaTeX,
		templatesFolder: "templates",
		templateName:    "template.tmpl",
	}
}

func (pdf *pdfGenerator) Command(c command) *pdfGenerator {
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

func (pdf *pdfGenerator) Funcs(funcs template.FuncMap) *pdfGenerator {
	pdf.funcs = funcs
	return pdf
}

func (pdf *pdfGenerator) Data(data interface{}) *pdfGenerator {
	pdf.data = data
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
		return errors.Wrapf(err, "could not create file %s", filename)
	}
	defer f.Close()

	if _, err := f.Write([]byte(pdf.output)); err != nil {
		return errors.Wrap(err, "could not write output to file")
	}

	return nil
}

func (pdf *pdfGenerator) Compile() *pdfGenerator {
	if pdf.err != nil {
		return pdf
	}

	if pdf.input != "" {
		pdf.output, pdf.err = pdf.compileRawInput()
		return pdf
	}

	pdf.output, pdf.err = pdf.compileTemplate()
	return pdf
}

func (pdf *pdfGenerator) compileRawInput() (string, error) {
	foldername := createFolderName()
	if err := os.Mkdir(foldername, os.ModePerm); err != nil {
		return "", errors.Wrap(err, "could not mkdir tmp folder")
	}

	f, err := os.Create(filepath.Join(foldername, "file.tex"))
	if err != nil {
		return "", errors.Wrap(err, "could not create file.tex")
	}

	if _, err := f.Write([]byte(pdf.input)); err != nil {
		f.Close()
		return "", errors.Wrap(err, "could not write file.tex contents")
	}
	f.Close()

	return pdf.executeCompilationAndClean(foldername)
}

func (pdf *pdfGenerator) executeCompilationAndClean(foldername string) (string, error) {
	if err := os.Chdir(foldername); err != nil {
		return "", errors.Wrap(err, "could not cd tmp folder")
	}

	if err := exec.Command(string(pdf.command), "file.tex").Run(); err != nil {
		return "", errors.Wrapf(err, "could not %s file.tex", pdf.command)
	}

	output, err := ioutil.ReadFile("file.pdf")
	if err != nil {
		return "", errors.Wrap(err, "error reading file.pdf")
	}

	if err := os.Chdir(".."); err != nil {
		return "", errors.Wrap(err, "could not cd ..")
	}

	if err := os.RemoveAll(foldername); err != nil {
		return "", errors.Wrap(err, "could not rm -r tmp folder")
	}

	return string(output), nil
}

// TODO implement
func (pdf *pdfGenerator) compileTemplate() (string, error) { return "", errors.New("not implemented") }

func CompileLatex(filename, text string) error { return compileTex(filename, text, "pdflatex") }
func CompileXetex(filename, text string) error { return compileTex(filename, text, "xelatex") }

func compileTex(filename, text, command string) error {
	f, foldername, err := createTmpFolderAndFile(filename)
	if err != nil {
		return errors.Wrap(err, "could not create tmp folder")
	}

	if _, err := f.Write([]byte(text)); err != nil {
		f.Close()
		return errors.Wrap(err, "could not write text to tex file")
	}
	f.Close()

	return executeCompilationAndClean(filename, command, foldername)
}

func CompileLatexTemplate(filename, templatename string, data interface{}, funcs template.FuncMap) error {
	return compileTexTemplate(filename, templatename, "pdflatex", data, funcs)
}
func CompileXetexTemplate(filename, templatename string, data interface{}, funcs template.FuncMap) error {
	return compileTexTemplate(filename, templatename, "xelatex", data, funcs)
}

func compileTexTemplate(filename, templatename, command string, data interface{}, funcs template.FuncMap) error {
	tmpl, err := template.New(templatename + ".tmpl").Funcs(funcs).ParseFiles(templatename + ".tmpl")
	if err != nil {
		return errors.Wrap(err, "could not create template")
	}

	f, foldername, err := createTmpFolderAndFile(filename)
	if err != nil {
		f.Close()
		return errors.Wrap(err, "could not create tmp folder")
	}

	if err := tmpl.Execute(f, data); err != nil {
		f.Close()
		return errors.Wrap(err, "could not execute template")
	}
	f.Close()

	return executeCompilationAndClean(filename, command, foldername)
}

func executeCompilationAndClean(filename, command, foldername string) error {
	if err := os.Chdir(foldername); err != nil {
		return errors.Wrap(err, "could not cd tmp folder")
	}

	if err := exec.Command(command, filename+".tex").Run(); err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not %s file.tex", command))
	}

	if err := os.Chdir(".."); err != nil {
		return errors.Wrap(err, "could not cd ..")
	}

	if err := os.Rename(foldername+"/"+filename+".pdf", filename+".pdf"); err != nil {
		return errors.Wrap(err, "could not mv file.pdf")
	}

	return errors.Wrap(os.RemoveAll(foldername), "could not rm -r tmp folder")
}

func createTmpFolderAndFile(filename string) (*os.File, string, error) {
	foldername := createFolderName()
	if err := os.Mkdir(foldername, os.ModePerm); err != nil {
		return nil, "", errors.Wrap(err, "could not mkdir tmp folder")
	}

	f, err := os.Create(foldername + "/" + filename + ".tex")
	if err != nil {
		return nil, "", errors.Wrap(err, "could not create .tex file")
	}

	return f, foldername, nil
}

func createFolderName() string { return "tmp" + fmt.Sprintf("%d", time.Now().UnixNano()) }
func createFileName() string   { return "sourcefile" + fmt.Sprintf("%d", time.Now().UnixNano()) }

func CompileLatexTemplateReader(templatename string, data interface{}, funcs template.FuncMap) (io.Reader, error) {
	return compileTexTemplateReader("pdflatex", templatename, data, funcs)
}

func CompileXetexTemplateReader(templatename string, data interface{}, funcs template.FuncMap) (io.Reader, error) {
	return compileTexTemplateReader("xelatex", templatename, data, funcs)
}

func compileTexTemplateReader(command, templatename string, data interface{}, funcs template.FuncMap) (io.Reader, error) {
	tmpl, err := template.New(templatename).Funcs(funcs).ParseFiles(templatename)
	if err != nil {
		return nil, errors.Wrap(err, "could not create template")
	}

	filename := createFileName()
	f, foldername, err := createTmpFolderAndFile(filename)
	if err != nil {
		f.Close()
		return nil, errors.Wrap(err, "could not create tmp folder")
	}

	if err := tmpl.Execute(f, data); err != nil {
		f.Close()
		return nil, errors.Wrap(err, "could not execute template")
	}
	f.Close()

	return executeCompilationReader(filename, command, foldername)
}

func executeCompilationReader(filename, command, foldername string) (io.Reader, error) {
	if err := os.Chdir(foldername); err != nil {
		return nil, errors.Wrap(err, "could not cd tmp folder")
	}

	if err := exec.Command(command, filename+".tex").Run(); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("could not %s file.tex", command))
	}

	if err := os.Chdir(".."); err != nil {
		return nil, errors.Wrap(err, "could not cd ..")
	}

	f, err := os.Open(foldername + "/" + filename + ".pdf")
	if err != nil {
		return nil, errors.Wrap(err, "could not open final pdf")
	}
	defer f.Close()
	defer os.RemoveAll(foldername)

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.Wrap(err, "could not read pdf")
	}

	return strings.NewReader(string(b)), nil
}

func CompileLatexFolder(folder string, template *template.Template, data interface{}) (io.Reader, error) {
	return compileTexFolder("pdflatex", folder, template, data)
}

func CompileXetexFolder(folder string, template *template.Template, data interface{}) (io.Reader, error) {
	return compileTexFolder("xelatex", folder, template, data)
}

func compileTexFolder(command, folder string, tmpl *template.Template, data interface{}) (io.Reader, error) {
	filename := createFileName()
	f, foldername, err := createTmpFolderAndFile(filename)
	if err != nil {
		f.Close()
		return nil, errors.Wrap(err, "could not create tmp folder")
	}

	files, err := ioutil.ReadDir(folder)
	if err != nil {
		f.Close()
		return nil, errors.Wrap(err, "could not read folder")
	}

	for _, file := range files {
		if err := exec.Command("cp", filepath.Join(folder, file.Name()), foldername).Run(); err != nil {
			f.Close()
			return nil, errors.Wrap(err, "could not copy folder contents to tmp folder")
		}
	}

	if err := tmpl.Execute(f, data); err != nil {
		f.Close()
		return nil, errors.Wrap(err, "could not execute template")
	}
	f.Close()

	return executeCompilationReader(filename, command, foldername)
}
