package pdftex

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"github.com/pkg/errors"
)

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

	filename := "sourcefile"
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
