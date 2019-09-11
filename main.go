package main

import (
	"bufio"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/urfave/cli"
)

const (
	maxLineLength = 1024
	templateRepo  = "https://github.com/lnsp/hive-templates"
	updateRepo    = "pull"
	copyRepo      = "clone"
	gitPath       = "git"
	goPath        = "go"
	goRunArg      = "run"
	serviceFile   = "service.go"
	serviceFolder = "service"
	methodFile    = "method.go"
	methodFolder  = "methods"
	runtimeFile   = "runtime.go"
	dockerFile    = "Dockerfile"
	aboutFolder   = "about"
	aboutFile     = "about.go"
)

var (
	autogenPath, _  = ioutil.TempDir("", "hive-autogen")
	templatePath, _ = ioutil.TempDir("", "hive-template")
)

var srcPath = filepath.Join(os.Getenv("GOPATH"), "src")
var input = bufio.NewReader(os.Stdin)

func getInput(text string) (string, error) {
	fmt.Print(text)
	buffer, _, err := input.ReadLine()
	if err != nil {
		return "", err
	}
	return string(buffer), nil
}

type methodHeader struct {
	Name, Service string
}

type serviceHeader struct {
	Name, Path, Version string
	Methods             []methodHeader
}

func updateTemplates() error {
	err := os.MkdirAll(templatePath, 0744)
	if err != nil {
		return err
	}
	// Download template
	cloneCmd := exec.Command(gitPath, copyRepo, templateRepo, templatePath)
	cloneCmd.Stdout = os.Stdout
	err = cloneCmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func copyFile(srcPath, destPath string) error {
	srcFile, err := os.Create(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err = io.Copy(destFile, srcFile); err != nil {
		return err
	}

	return nil
}

func actionNew(c *cli.Context) error {
	serviceName, err := getInput("Enter service name: ")
	if err != nil {
		return err
	}

	servicePath, err := getInput("Enter service path: ")
	if err != nil {
		return err
	}

	serviceVersion, err := getInput("Enter service version: ")
	if err != nil {
		return err
	}

	service := serviceHeader{
		Name:    serviceName,
		Path:    servicePath,
		Version: serviceVersion,
		Methods: make([]methodHeader, 0),
	}

	addMethod, err := getInput("Do you want to add a method? [Y/n] ")
	if err != nil {
		return err
	}

	for addMethod != "n" {
		methodName, err := getInput("Enter method name: ")
		if err != nil {
			return err
		}
		methodName = strings.Title(methodName)

		service.Methods = append(service.Methods, methodHeader{
			Name:    methodName,
			Service: service.Name,
		})

		addMethod, err = getInput("Do you want to add a method? [Y/n] ")
		if err != nil {
			return err
		}
	}

	updateTemplates()

	serviceTemplate := template.Must(template.ParseFiles(
		filepath.Join(templatePath, serviceFolder, serviceFile),
	))

	methodTemplate := template.Must(template.ParseFiles(
		filepath.Join(templatePath, methodFolder, methodFile),
	))

	runtimeTemplate := template.Must(template.ParseFiles(
		filepath.Join(templatePath, runtimeFile),
	))

	if _, err := os.Stat(filepath.Join(srcPath, service.Path)); err == nil {
		fmt.Println("Service already exists at the specified path.")
		return nil
	}

	// Create base path
	if err := os.MkdirAll(filepath.Join(
		srcPath,
		service.Path,
	), 0744); err != nil {
		return err
	}

	// Create service package
	if err := os.MkdirAll(filepath.Join(
		srcPath,
		service.Path,
		serviceFolder,
	), 0744); err != nil {
		return err
	}

	// Create methods package
	if err := os.MkdirAll(filepath.Join(
		srcPath,
		service.Path,
		methodFolder,
	), 0744); err != nil {
		return err
	}

	if err := writeTemplateToFile(filepath.Join(
		srcPath,
		service.Path,
		serviceFolder,
		serviceFile,
	), serviceTemplate, service); err != nil {
		return err
	}

	for _, method := range service.Methods {
		if err := writeTemplateToFile(filepath.Join(
			srcPath,
			service.Path,
			methodFolder,
			strings.ToLower(method.Name)+".go",
		), methodTemplate, method); err != nil {
			return err
		}
	}

	if err := writeTemplateToFile(filepath.Join(
		srcPath,
		service.Path,
		runtimeFile,
	), runtimeTemplate, service); err != nil {
		return err
	}

	if err := copyFile(filepath.Join(
		templatePath,
		dockerFile,
	), filepath.Join(
		srcPath,
		service.Path,
		dockerFile,
	)); err != nil {
		return err
	}

	fmt.Println("Successfully generated service.")
	return nil
}

func writeTemplateToFile(path string, tmpl *template.Template, val interface{}) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	err = tmpl.Execute(file, val)
	if err != nil {
		return err
	}
	return nil
}

func actionAbout(c *cli.Context) error {
	err := updateTemplates()
	if err != nil {
		return errors.New("failed to update templates: " + err.Error())
	}

	service := struct {
		Name, Path string
	}{}
	service.Name = filepath.Base(c.Args()[0])
	service.Path = c.Args()[0]

	// Store generated about file
	aboutTemplate := template.Must(template.ParseFiles(
		filepath.Join(templatePath, aboutFolder, aboutFile),
	))

	// Create autogen directory
	err = os.MkdirAll(autogenPath, 0744)
	if err != nil {
		return errors.New("failed to create directory: " + err.Error())
	}

	generatedName := filepath.Join(autogenPath, service.Name+".go")

	// Write about file
	if err := writeTemplateToFile(generatedName, aboutTemplate, service); err != nil {
		return errors.New("failed to write template: " + err.Error())
	}

	// Run "go run"
	runCmd := exec.Command(goPath, goRunArg, generatedName)
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr
	if err := runCmd.Run(); err != nil {
		return errors.New("execution failed: " + err.Error())
	}

	return nil
}

func actionBuild(c *cli.Context) error {
	return nil
}

func actionDeploy(c *cli.Context) error {
	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "Hive"
	app.Usage = "Microservice administration tool"
	app.Version = "0.1.0"

	app.Commands = []cli.Command{
		{
			Name:    "new",
			Aliases: []string{"n"},
			Usage:   "Create a new Hive microservice",
			Action:  actionNew,
		},

		{
			Name:    "about",
			Aliases: []string{"a"},
			Usage:   "Display information about the service",
			Action: func(c *cli.Context) error {
				err := actionAbout(c)
				if err != nil {
					fmt.Println(err)
				}

				return nil
			},
		}, /*
			{
				Name:    "build",
				Aliases: []string{"b"},
				Usage:   "Build and test the microservice",
				Action:  actionBuild,
			},
			{
				Name:    "deploy",
				Aliases: []string{"d"},
				Usage:   "Deploy the microservice",
				Action:  actionDeploy,
			},
		*/
	}

	app.Run(os.Args)
}
