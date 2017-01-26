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
	templatePath  = "/tmp/hive/templates"
	templateRepo  = "https://github.com/lnsp/hive-templates"
	updateRepo    = "pull"
	copyRepo      = "clone"
	gitPath       = "git"
	serviceFile   = "service.go"
	methodFile    = "method.go"
	runtimeFile   = "runtime.go"
	runtimePath   = "runtime"
	dockerFile    = "Dockerfile"
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

	fmt.Println("Updating service templates ...")

	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		err = os.MkdirAll(templatePath, 0744)
		if err != nil {
			return err
		}
		// Download templates
		cloneCmd := exec.Command(gitPath, copyRepo, templateRepo, templatePath)
		cloneCmd.Stdout = os.Stdout
		cloneCmd.Stderr = os.Stderr
		err = cloneCmd.Run()
		if err != nil {
			return err
		}
	}

	updateCmd := exec.Command(gitPath, updateRepo)
	updateCmd.Dir = templatePath
	updateCmd.Stdout = os.Stdout
	updateCmd.Stderr = os.Stderr
	err = updateCmd.Run()
	if err != nil {
		return err
	}

	fmt.Println("Loading templates ...")
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

	err = os.MkdirAll(filepath.Join(
		srcPath,
		service.Path,
	), 0744)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Join(
		srcPath,
		service.Path,
		runtimePath,
	), 0744)
	if err != nil {
		return err
	}

	generatedService, err := os.Create(filepath.Join(
		srcPath,
		service.Path,
		serviceFile,
	))
	if err != nil {
		return err
	}
	defer generatedService.Close()

	serviceTemplate.Execute(generatedService, service)

	for _, method := range service.Methods {
		methodFileName := strings.ToLower(method.Name) + ".go"
		generatedMethod, err := os.Create(filepath.Join(
			srcPath,
			service.Path,
			methodFileName,
		))
		if err != nil {
			return err
		}
		defer generatedMethod.Close()
		methodTemplate.Execute(generatedMethod, method)
	}

	generatedRuntime, err := os.Create(filepath.Join(
		srcPath,
		service.Path,
		runtimePath,
		runtimeFile,
	))
	if err != nil {
		return err
	}
	defer generatedRuntime.Close()

	runtimeTemplate.Execute(generatedRuntime, service)

	srcDockerfile, err := os.Create(filepath.Join(
		templatePath,
		runtimePath,
		dockerFile,
	))
	if err != nil {
		return err
	}
	defer srcDockerfile.Close()

	copiedDockerfile, err := os.Create(filepath.Join(
		srcPath,
		service.Path,
		runtimePath,
		dockerFile,
	))
	if err != nil {
		return err
	}
	defer copiedDockerfile.Close()

	_, err = io.Copy(copiedDockerfile, srcDockerfile)
	if err != nil {
		return err
	}

	fmt.Println("Successfully generated service.")
	return nil
}

func actionAbout(c *cli.Context) error {
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
		/*
			{
				Name:    "about",
				Aliases: []string{"a"},
				Usage:   "Display information about the service",
				Action:  actionAbout,
			},
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
