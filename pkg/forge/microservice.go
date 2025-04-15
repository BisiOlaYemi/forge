package forge

import (
	"fmt"
	"os"
)

func CreateNewProject(projectName string) {
	fmt.Println("Creating new Forge project:", projectName)
	os.MkdirAll(projectName+"/controllers", os.ModePerm)
	os.MkdirAll(projectName+"/models", os.ModePerm)
	os.MkdirAll(projectName+"/routes", os.ModePerm)
	os.MkdirAll(projectName+"/services", os.ModePerm)
	fmt.Println("Project scaffold created.")
}

func GenerateController(name string, service string) {
	var path string
	if service != "" {
		path = fmt.Sprintf("services/%s/controllers", service)
	} else {
		path = "controllers"
	}
	os.MkdirAll(path, os.ModePerm)
	controllerFile := fmt.Sprintf("%s/%sController.go", path, name)
	content := fmt.Sprintf(`package controllers

import "fmt"

func %sController() {
	fmt.Println("%s controller logic here")
}`, name, name)
	os.WriteFile(controllerFile, []byte(content), 0644)
	fmt.Println("Controller created at:", controllerFile)
}

func GenerateModel(name string, service string) {
	var path string
	if service != "" {
		path = fmt.Sprintf("services/%s/models", service)
	} else {
		path = "models"
	}
	os.MkdirAll(path, os.ModePerm)
	modelFile := fmt.Sprintf("%s/%s.go", path, name)
	content := fmt.Sprintf(`package models

// %s represents the model structure
type %s struct {
	ID   int
	Name string
}`, name, name)
	os.WriteFile(modelFile, []byte(content), 0644)
	fmt.Println("Model created at:", modelFile)
}

func CreateMicroservice(serviceName string) {
	base := fmt.Sprintf("services/%s", serviceName)
	folders := []string{
		base + "/controllers",
		base + "/models",
		base + "/routes",
		base + "/config",
	}
	for _, folder := range folders {
		os.MkdirAll(folder, os.ModePerm)
	}
	mainFile := base + "/main.go"
	mainContent := fmt.Sprintf(`package main

import "fmt"

func main() {
	fmt.Println("%s microservice started")
}`, serviceName)
	os.WriteFile(mainFile, []byte(mainContent), 0644)
	fmt.Println("Microservice scaffold created at:", base)
}
