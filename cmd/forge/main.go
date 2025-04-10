package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/BisiOlaYemi/forge/pkg/forge"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "forge",
	Short: "Forge - A modern Go web framework",
	Long: `Forge is a modern, full-stack web framework for Go — 
designed to combine developer happiness, performance, and structure.`,
}

func init() {
	// New project command
	newCmd := &cobra.Command{
		Use:   "new [project-name]",
		Short: "Create a new Forge project",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := createNewProject(args[0]); err != nil {
				fmt.Printf("Error creating project: %v\n", err)
				os.Exit(1)
			}
		},
	}

	// Generate commands
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

	// Serve command
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
	// Create a new Forge application
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

	// Create a hot reloader
	reloader, err := forge.NewHotReloader(app)
	if err != nil {
		fmt.Printf("Error creating hot reloader: %v\n", err)
		os.Exit(1)
	}

	// Start the hot reloader
	if err := reloader.Start(); err != nil {
		fmt.Printf("Error starting hot reloader: %v\n", err)
		os.Exit(1)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Stop the hot reloader
	if err := reloader.Stop(); err != nil {
		fmt.Printf("Error stopping hot reloader: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
} 