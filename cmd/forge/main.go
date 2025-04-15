package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/BisiOlaYemi/forge/pkg/forge"
	"github.com/spf13/cobra"
	"github.com/fatih/color"
)

var rootCmd = &cobra.Command{
	Use:   "forge",
	Short: "Forge - A modern Go web framework",
	Long: `Forge is a modern, full-stack web framework for Go — 
designed to combine developer happiness, performance, and structure.`,
}

func init() {
	
	newCmd := &cobra.Command{
		Use:   "new [project-name]",
		Short: "Create a new Forge project",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := createNewProject(args[0]); err != nil {
				fmt.Printf("Error creating project: %v\n", err)
				os.Exit(1)
			}
			installSuccessMessage()
		},
	}

	
	makeControllerCmd := &cobra.Command{
		Use:   "make:controller [name]",
		Short: "Generate a new controller",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := generateController(args[0]); err != nil {
				fmt.Printf("Error generating controller: %v\n", err)
				os.Exit(1)
			}
		},
	}

	makeModelCmd := &cobra.Command{
		Use:   "make:model [name]",
		Short: "Generate a new model",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := generateModel(args[0]); err != nil {
				fmt.Printf("Error generating model: %v\n", err)
				os.Exit(1)
			}
		},
	}

	
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the development server",
		Run: func(cmd *cobra.Command, args []string) {
			startServer()
		},
	}

	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(makeControllerCmd)
	rootCmd.AddCommand(makeModelCmd)
	rootCmd.AddCommand(serveCmd)
}

func startServer() {
	
	app, err := forge.New(&forge.Config{
		Name:        "Forge App",
		Version:     "1.0.0",
		Description: "A Forge application",
		Server: forge.ServerConfig{
			Host:     "localhost",
			Port:     3000,
			BasePath: "/",
		},
	})
	if err != nil {
		fmt.Printf("Error creating application: %v\n", err)
		os.Exit(1)
	}

	
	reloader, err := forge.NewHotReloader(app)
	if err != nil {
		fmt.Printf("Error creating hot reloader: %v\n", err)
		os.Exit(1)
	}

	
	if err := reloader.Start(); err != nil {
		fmt.Printf("Error starting hot reloader: %v\n", err)
		os.Exit(1)
	}

	
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	
	if err := reloader.Stop(); err != nil {
		fmt.Printf("Error stopping hot reloader: %v\n", err)
		os.Exit(1)
	}
}

func installSuccessMessage() {
    cyan := color.New(color.FgCyan).SprintFunc()
    green := color.New(color.FgGreen).SprintFunc()
    yellow := color.New(color.FgYellow).SprintFunc()
    bold := color.New(color.Bold).SprintFunc()

    fmt.Println("\n✨ " + bold("Welcome to Forge: The GoPowerhouse Web Framework!") + " ✨")
    fmt.Println("🔨 Created with passion by " + green("Yemi Ogunrinde"))
    fmt.Println(cyan("\nLet’s build something amazing together! 🚀"))
    fmt.Println(yellow("Happy Coding! 💻"))
    fmt.Println("☕ Like it? " + bold("Buy me a coffee") + " at: https://buymeacoffee.com/yemiogunrinde\n")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
} 