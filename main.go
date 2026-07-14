package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	if err := run(os.Args, os.Stdout, os.Stderr); err != nil {
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) < 2 {
		printUsage(stderr)
		return errors.New("missing command")
	}

	command := args[1]
	switch command {
	case "serve":
		serveCmd := flag.NewFlagSet("serve", flag.ContinueOnError)
		serveCmd.SetOutput(stderr)
		port := serveCmd.String("port", "8080", "Port to serve presentation on")
		serveCmd.Usage = func() {
			fmt.Fprintln(serveCmd.Output(), "Usage: gophern serve [-port 8080] <file.md>")
			fmt.Fprintln(serveCmd.Output(), "Options:")
			serveCmd.PrintDefaults()
		}
		if err := serveCmd.Parse(args[2:]); err != nil {
			return err
		}
		if serveCmd.NArg() < 1 {
			fmt.Fprintln(stderr, "Error: Markdown file path required for serve command")
			return errors.New("missing markdown file")
		}
		markdownFile := serveCmd.Arg(0)
		fmt.Fprintf(stdout, "Serving %s on port %s...\n", markdownFile, *port)
		// Handled in later tasks
		return nil

	case "export":
		exportCmd := flag.NewFlagSet("export", flag.ContinueOnError)
		exportCmd.SetOutput(stderr)
		output := exportCmd.String("o", "presentation.html", "Output HTML file path")
		exportCmd.Usage = func() {
			fmt.Fprintln(exportCmd.Output(), "Usage: gophern export [-o output.html] <file.md>")
			fmt.Fprintln(exportCmd.Output(), "Options:")
			exportCmd.PrintDefaults()
		}
		if err := exportCmd.Parse(args[2:]); err != nil {
			return err
		}
		if exportCmd.NArg() < 1 {
			fmt.Fprintln(stderr, "Error: Markdown file path required for export command")
			return errors.New("missing markdown file")
		}
		markdownFile := exportCmd.Arg(0)
		fmt.Fprintf(stdout, "Exporting %s to %s...\n", markdownFile, *output)
		// Handled in later tasks
		return nil

	default:
		printUsage(stderr)
		return fmt.Errorf("unknown command: %s", command)
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: gophern <command> [arguments]")
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  serve [-port 8080] <file.md>  Start the presentation server")
	fmt.Fprintln(w, "  export [-o output.html] <file.md> Export to a single HTML file")
}
