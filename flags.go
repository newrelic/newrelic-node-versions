package main

import (
	"fmt"
	"github.com/MakeNowJust/heredoc/v2"
	flag "github.com/spf13/pflag"
	"os"
)

type appFlags struct {
	aiCompatJsonFile string
	noExternals      bool
	replaceInFile    string
	repoDir          string
	testDir          string
	verbose          bool

	startMarker string
	endMarker   string
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
		&flags.aiCompatJsonFile,
		"ai-compat-json",
		"a",
		"",
		heredoc.Doc(`
			Path to the ai-compat.json file that describes the AI Monitoring
			compatibility of the agent. The default is to use the JSON file included
			in the mainline agent repository.
		`),
	)

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

	fs.StringVarP(
		&flags.replaceInFile,
		"replace-in-file",
		"R",
		"",
		heredoc.Doc(`
			Specify a target file in which the results will be written. Normally,
			the result is written to stdout. When this flag is given, the result
			will be written to the specified file. The generated text will replace
			all text in the file between two marker lines. The markers can be defined
			through environment variables: START_MARKER and END_MARKER. Default values
			are "{/* begin: compat-table */}"
			and "{/* end: compat-table */}".
		`,
		),
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

	readEnvironment()
	return fs.Parse(args[1:])
}

func readEnvironment() {
	marker, envIsSet := os.LookupEnv("START_MARKER")
	if envIsSet == true {
		flags.startMarker = marker
	} else {
		flags.startMarker = "{/* begin: compat-table */}"
	}

	marker, envIsSet = os.LookupEnv("END_MARKER")
	if envIsSet == true {
		flags.endMarker = marker
	} else {
		flags.endMarker = "{/* end: compat-table */}"
	}
}

func printUsage(help string) {
	fmt.Println(usageText)
	fmt.Println(help)
}
