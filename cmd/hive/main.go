package main

import (
	"bufio"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/urfave/cli.v1"
)

const (
	maxLineLength = 1024
	autogenPath   = "/tmp/hive/autogen"
	templatePath  = "/tmp/hive/templates"
	templateRepo  = "https://github.com/lnsp/hive-templates"
	updateRepo    = "pull"
	copyRepo      = "clone"
	gitPath       = "git"
	goPath        = "go"
	goRunArg      = "run"
	serviceFile   = "service.go"
	methodFile    = "method.go"
	runtimeFile   = "runtime.go"
	runtimePath   = "runtime"
	dockerFile    = "Dockerfile"
	aboutFile     = "about.go"
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
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		err = os.MkdirAll(templatePath, 0744)
		if err != nil {
			return err
		}
		// Download templates
		cloneCmd := exec.Command(gitPath, copyRepo, templateRepo, templatePath)
		cloneCmd.Stderr = os.Stderr
		err = cloneCmd.Run()
		if err != nil {
			return err
		}
	}

	updateCmd := exec.Command(gitPath, updateRepo)
	updateCmd.Dir = templatePath
	updateCmd.Stderr = os.Stderr
	err := updateCmd.Run()
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
		filepath.Join(templatePath, serviceFile),
	))

	methodTemplate := template.Must(template.ParseFiles(
		filepath.Join(templatePath, methodFile),
	))

	runtimeTemplate := template.Must(template.ParseFiles(
		filepath.Join(templatePath, runtimePath, runtimeFile),
	))

	if _, err := os.Stat(filepath.Join(srcPath, service.Path)); err == nil {
		fmt.Println("Service already exists at the specified path.")
		return nil
	}

	if err := os.MkdirAll(filepath.Join(
		srcPath,
		service.Path,
	), 0744); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(
		srcPath,
		service.Path,
		runtimePath,
	), 0744); err != nil {
		return err
	}

	if err := writeTemplateToFile(filepath.Join(
		srcPath,
		service.Path,
		serviceFile,
	), serviceTemplate, service); err != nil {
		return err
	}

	for _, method := range service.Methods {
		if err := writeTemplateToFile(filepath.Join(
			srcPath,
			service.Path,
			strings.ToLower(method.Name)+".go",
		), methodTemplate, method); err != nil {
			return err
		}
	}

	if err := writeTemplateToFile(filepath.Join(
		srcPath,
		service.Path,
		runtimePath,
		runtimeFile,
	), runtimeTemplate, service); err != nil {
		return err
	}

	if err := copyFile(filepath.Join(
		templatePath,
		runtimePath,
		dockerFile,
	), filepath.Join(
		srcPath,
		service.Path,
		runtimePath,
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
		return err
	}

	service := struct {
		Name, Path string
	}{}
	service.Name = filepath.Base(c.Args()[0])
	service.Path = c.Args()[0]

	// Store generated about file
	aboutTemplate := template.Must(template.ParseFiles(
		filepath.Join(templatePath, aboutFile),
	))

	// Create autogen directory
	err = os.MkdirAll(autogenPath, 0744)
	if err != nil {
		return err
	}

	// Write about file
	if err := writeTemplateToFile(filepath.Join(
		autogenPath,
		aboutFile,
	), aboutTemplate, service); err != nil {
		return err
	}

	// Run "go run"
	runCmd := exec.Command(goPath, goRunArg, filepath.Join(autogenPath, aboutFile))
	runCmd.Stdout = os.Stdout
	if err := runCmd.Run(); err != nil {
		return err
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
			Action:  actionAbout,
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
