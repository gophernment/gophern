package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsageAndExit()
	}

	command := os.Args[1]
	switch command {
	case "serve":
		serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
		port := serveCmd.String("port", "8080", "Port to serve presentation on")
		serveCmd.Parse(os.Args[2:])
		if serveCmd.NArg() < 1 {
			fmt.Println("Error: Markdown file path required for serve command")
			os.Exit(1)
		}
		markdownFile := serveCmd.Arg(0)
		fmt.Printf("Serving %s on port %s...\n", markdownFile, *port)
		// Handled in later tasks
	case "export":
		exportCmd := flag.NewFlagSet("export", flag.ExitOnError)
		output := exportCmd.String("o", "presentation.html", "Output HTML file path")
		exportCmd.Parse(os.Args[2:])
		if exportCmd.NArg() < 1 {
			fmt.Println("Error: Markdown file path required for export command")
			os.Exit(1)
		}
		markdownFile := exportCmd.Arg(0)
		fmt.Printf("Exporting %s to %s...\n", markdownFile, *output)
		// Handled in later tasks
	default:
		printUsageAndExit()
	}
}

func printUsageAndExit() {
	fmt.Println("Usage: gophern <command> [arguments]")
	fmt.Println("Commands:")
	fmt.Println("  serve <file.md> [--port 8080]  Start the presentation server")
	fmt.Println("  export <file.md> [-o output.html] Export to a single HTML file")
	os.Exit(1)
}
