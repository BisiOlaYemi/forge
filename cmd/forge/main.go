package main

import (
	"fmt"
	"os"

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
			createNewProject(args[0])
		},
	}

	// Generate commands
	makeControllerCmd := &cobra.Command{
		Use:   "make:controller [name]",
		Short: "Generate a new controller",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			generateController(args[0])
		},
	}

	makeModelCmd := &cobra.Command{
		Use:   "make:model [name]",
		Short: "Generate a new model",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			generateModel(args[0])
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

func createNewProject(name string) {
	fmt.Printf("Creating new Forge project: %s\n", name)
	// TODO: Implement project scaffolding
}

func generateController(name string) {
	fmt.Printf("Generating controller: %s\n", name)
	// TODO: Implement controller generation
}

func generateModel(name string) {
	fmt.Printf("Generating model: %s\n", name)
	// TODO: Implement model generation
}

func startServer() {
	fmt.Println("Starting Forge development server...")
	// TODO: Implement server startup
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
} 