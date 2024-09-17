<a href="https://opensource.newrelic.com/oss-category/#new-relic-experimental"><picture><source media="(prefers-color-scheme: dark)" srcset="https://github.com/newrelic/opensource-website/raw/main/src/images/categories/dark/Experimental.png"><source media="(prefers-color-scheme: light)" srcset="https://github.com/newrelic/opensource-website/raw/main/src/images/categories/Experimental.png"><img alt="New Relic Open Source experimental project banner." src="https://github.com/newrelic/opensource-website/raw/main/src/images/categories/Experimental.png"></picture></a>

# New Relic Node.js 3rd Party Versions
This is a utility used by the New Relic Node.js agent team.  It will clone `node-newrelic`, `newrelic-node-apollo-server-plugin` and `newrelic-node-nextjs` repos and create a 3rd party library compatibility report.

## Installation

```sh
go install github.com/newrelic/newrelic-node-versions@latest
```

### Tools

This project relies on some community tools that require extra installation:

+ [Task](https://taskfile.dev): `go install github.com/go-task/task/v3/cmd/task@latest`
+ [golangci-lint](https://golangci-lint.run): `github.com/golangci/golangci-lint/cmd/golangci-lint@latest`

## Getting Started

## Usage

```sh
./nrversions
```

Will output a report document as shown in our main repo's
[compatibility doc](https://github.com/newrelic/node-newrelic/blob/main/compatibility.md).

For optional CLI args, run:

```sh
❯ ./nrversions -help

nrversions - This tool is used to generate a document detailing the modules that
the newrelic Node.js agent instruments and the version ranges of those
modules.

The following flags are supported:


  Flags: 
       --version   Displays the program version string.
    -h --help      Displays help with available flag, subcommand, and positional value parameters.
    -ai-compat-json --a         Path to the ai-compat.json file that describes the AI Monitoring
compatibility of the agent. The default is to use the JSON file included
in the mainline agent repository.

    -no-externals --n         Disable cloning and processing of external repos. An external repo is
one that provides extra functionality to the "newrelic" module. This
allows processing a single repo with --repo-dir. The default, i.e. not
supplying this flaggy, is to process all known external repos.

    -replace-in-file --R         Specify a target file in which the results will be written. Normally,
the result is written to stdout. When this flaggy is given, the result
will be written to the specified file. The generated text will replace
all text in the file between two marker lines. The markers can be defined
through environment variables: START_MARKER and END_MARKER. Default values
are "{/* begin: compat-table */}"
and "{/* end: compat-table */}".

    -repo-dir --r         Specify a local directory that contains a Node.js instrumentation repo.
If not provided, the main agent GitHub repository will be cloned to a
local temporary directory and that will be used.

    -test-dir --t            Specify the test directory to parse the package.json files.
   If not provided, it will default to 'test/versioned'. This applies to
the repo provided by the --repo-dir flaggy.
 
    -verbose --v         Enable verbose output. As the data is being loaded and parsed various
logs will be written to stderr that should give indicators of what
is happening.
```

## Building

```sh
go build
```

## Testing

```sh
go test

```

## Contribute

We encourage your contributions to improve newrelic-node-versions! Keep in mind that when you submit your pull request, you'll need to sign the CLA via the click-through using CLA-Assistant. You only have to sign the CLA one time per project.

If you have any questions, or to execute our corporate CLA (which is required if your contribution is on behalf of a company), drop us an email at opensource@newrelic.com.

**A note about vulnerabilities**
As noted in our [security policy](./SECURITY.md), New Relic is committed to the privacy and security of our customers and their data. We believe that providing coordinated disclosure by security researchers and engaging with the security community are important means to achieve our security goals.

If you believe you have found a security vulnerability in this project or any of New Relic's products or websites, we welcome and greatly appreciate you reporting it to New Relic through [our bug bounty program](https://docs.newrelic.com/docs/security/security-privacy/information-security/report-security-vulnerabilities/).

If you would like to contribute to this project, review [these guidelines](./CONTRIBUTING.md).

To [all contributors](https://github.com/newrelic/newrelic-node-versions/graphs/contributors), we thank you!  Without your contribution, this project would not be what it is today. 

## License
New Relic Node.js 3rd Party Versions is licensed under the [Apache 2.0](http://apache.org/licenses/LICENSE-2.0.txt) License. It also uses source code from third-party libraries. You can find full details on which libraries are used and the terms under which they are licensed in the [third-party notices document.](THIRD_PARTY_NOTICES.md)
