package main

import (
	"fmt"
	"slices"

	"github.com/MakeNowJust/heredoc/v2"
	flag "github.com/spf13/pflag"
)

type appFlags struct {
	outputFormat *StringEnumValue
	repoDir      string
	testDir      string
	verbose      bool
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

	flags.outputFormat = NewStringEnumValue(
		[]string{"ascii", "markdown"},
		"ascii",
	)
	fs.VarP(
		flags.outputFormat,
		"output-format",
		"o",
		heredoc.Doc(`
			Specify the format to write the results as. The default is an ASCII
			table. Possible values: "ascii" or "markdown".
		`),
	)

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

	fs.StringVarP(
		&flags.testDir,
		"test-dir",
		"t",
		"",
		heredoc.Doc(`
      Specify the test directory to parse the package.json files.
      If not provided, it will default to 'test/versioned'.
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

// StringEnumValue is a custom [flag.Value] implementation that
// allows a specific set of string values.
// See https://github.com/spf13/pflag/issues/236#issuecomment-931600452
type StringEnumValue struct {
	allowed []string
	value   string
}

func NewStringEnumValue(allowed []string, def string) *StringEnumValue {
	return &StringEnumValue{
		allowed: allowed,
		value:   def,
	}
}

func (sev *StringEnumValue) String() string {
	return sev.value
}

func (sev *StringEnumValue) Set(val string) error {
	if slices.Contains(sev.allowed, val) == false {
		return fmt.Errorf("%s is not an allowed value", val)
	}

	sev.value = val

	return nil
}

func (sev *StringEnumValue) Type() string {
	return "string"
}
