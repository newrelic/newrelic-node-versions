package main

import (
	"fmt"
	"slices"

	"github.com/MakeNowJust/heredoc/v2"
	flag "github.com/spf13/pflag"
)

type appFlags struct {
	noExternals  bool
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

	fs.BoolVarP(
		&flags.noExternals,
		"no-externals",
		"n",
		false,
		heredoc.Doc(`
			Disable cloning and processing of external repos. An external repo is
			one that provides extra functionality to the "newrelic" module. This
			allows processing a single repo with --repo-dir. The default, i.e. not
			supplying this flag, is to process all known external repos.
		`),
	)

	flags.outputFormat = NewStringEnumValue(
		[]string{"ascii", "markdown"},
		"markdown",
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
			Specify a local directory that contains a Node.js instrumentation repo.
			If not provided, the main agent GitHub repository will be cloned to a
			local temporary directory and that will be used.
		`),
	)

	fs.StringVarP(
		&flags.testDir,
		"test-dir",
		"t",
		"",
		heredoc.Doc(`
      Specify the test directory to parse the package.json files.
      If not provided, it will default to 'test/versioned'. This applies to
			the repo provided by the --repo-dir flag.
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

// NewStringEnumValue creates a new [StringEnumValue] with a defined set of
// allowed values and a sets the initial value to a default value (`def`).
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
