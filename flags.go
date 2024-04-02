package main

import (
	"fmt"
	"github.com/MakeNowJust/heredoc/v2"
	flag "github.com/spf13/pflag"
)

type appFlags struct {
	repoDir string
	verbose bool
}

var flags = appFlags{}

var usageText = heredoc.Doc(`
	This tool is used to generate a document detailing the modules that
	the newrelic Node.js agent instruments and the version ranges of those
	modules.

	The following flags are supported:
`)

func createAndParseFlags(args []string) error {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.Usage = func() {
		printUsage(fs.FlagUsages())
	}

	fs.StringVarP(
		&flags.repoDir,
		"repo-dir",
		"r",
		"",
		heredoc.Doc(`
			Specify a local directory that contains the node-newrelic repo.
			If not provided, the GitHub repository will be cloned to a local temporary
			directory and that will be used.
		`),
	)

	fs.BoolVarP(
		&flags.verbose,
		"verbose",
		"v",
		false,
		heredoc.Doc(`
			Enable verbose output. As the data is being loaded and parsed various
			logs will be written to stderr that should give indicators of what
			is happening.
		`),
	)

	// TODO: add flags for generating different formats:
	// 1. markdown (for GitHub repo)
	// 2. docs site
	// They should also generate new PRs for the respective repos.

	return fs.Parse(args[1:])
}

func printUsage(help string) {
	fmt.Println(usageText)
	fmt.Println(help)
}
