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

Will output a compatibility report like this

```
+------------------------+---------------------+----------------------------+---------------+----------------------+
| NAME                   | MINSUPPORTEDVERSION | MINSUPPORTEDVERSIONRELEASE | LATESTVERSION | LATESTVERSIONRELEASE |
+------------------------+---------------------+----------------------------+---------------+----------------------+
| @apollo/gateway        | 2.3.0               | 2023-01-25                 | 2.7.3         | 2024-04-16           |
| @apollo/server         | 4.0.0               | 2022-10-10                 | 4.10.3        | 2024-04-15           |
| @elastic/elasticsearch | 7.16.0              | 2021-12-15                 | 8.13.1        | 2024-04-09           |
| @grpc/grpc-js          | 1.4.0               | 2021-10-13                 | 1.10.6        | 2024-04-03           |
| @hapi/hapi             | 20.1.2              | 2021-03-20                 | 21.3.9        | 2024-04-09           |
| @langchain/core        | 0.1.17              | 2024-01-19                 | 0.1.58        | 2024-04-16           |
| @nestjs/cli            | 8.0.0               | 2021-07-07                 | 10.3.2        | 2024-02-07           |
| @prisma/client         | 5.0.0               | 2023-07-11                 | 5.12.1        | 2024-04-04           |
| amqplib                | 0.5.0               | 2016-11-01                 | 0.10.4        | 2024-04-11           |
| apollo-server          | 2.14.0              | 2020-05-27                 | 3.13.0        | 2023-11-14           |
| apollo-server-express  | 2.14.0              | 2020-05-27                 | 3.13.0        | 2023-11-14           |
| apollo-server-fastify  | 2.14.0              | 2020-05-27                 | 3.13.0        | 2023-11-14           |
| apollo-server-hapi     | 3.0.0               | 2021-07-07                 | 3.13.0        | 2023-11-14           |
| apollo-server-koa      | 2.14.0              | 2020-05-27                 | 3.13.0        | 2023-11-14           |
| apollo-server-lambda   | 2.14.0              | 2020-05-27                 | 3.13.0        | 2023-11-14           |
| bluebird               | 2.0.0               | 2014-06-04                 | 3.7.2         | 2019-11-28           |
| bunyan                 | 1.8.12              | 2017-08-02                 | 1.8.15        | 2021-01-08           |
| cassandra-driver       | 3.4.0               | 2018-02-05                 | 4.7.2         | 2023-09-21           |
| connect                | 2.0.0               | 2012-02-28                 | 3.7.0         | 2019-05-18           |
| director               | 1.2.0               | 2013-04-01                 | 1.2.8         | 2015-02-04           |
| express                | 4.6.0               | 2014-07-12                 | 4.19.2        | 2024-03-25           |
| fastify                | 2.0.0               | 2019-02-25                 | 4.26.2        | 2024-03-03           |
| generic-pool           | 2.4.0               | 2016-01-18                 | 3.9.0         | 2022-09-10           |
| ioredis                | 3.0.0               | 2017-05-18                 | 5.3.2         | 2023-04-15           |
| mongodb                | 2.1.0               | 2015-12-06                 | 6.5.0         | 2024-03-11           |
| mysql                  | 2.2.0               | 2014-04-27                 | 2.18.1        | 2020-01-23           |
| mysql2                 | 1.3.1               | 2017-05-31                 | 3.9.4         | 2024-04-09           |
| next                   | 13.0.0              | 2022-10-25                 | 14.2.1        | 2024-04-12           |
| openai                 | 4.0.0               | 2023-08-16                 | 4.35.0        | 2024-04-16           |
| pg                     | 8.2.0               | 2020-05-13                 | 8.11.5        | 2024-04-02           |
| pino                   | 7.0.0               | 2021-10-14                 | 8.20.0        | 2024-04-06           |
| redis                  | 2.0.0               | 2015-09-21                 | 4.6.13        | 2024-02-05           |
| restify                | 5.0.0               | 2017-06-27                 | 11.1.0        | 2023-02-24           |
| superagent             | 2.0.0               | 2016-05-29                 | 8.1.2         | 2023-08-15           |
| undici                 | 4.7.0               | 2021-09-22                 | 6.13.0        | 2024-04-12           |
| winston                | 3.0.0               | 2018-06-12                 | 3.13.0        | 2024-03-24           |
+------------------------+---------------------+----------------------------+---------------+----------------------+
```

For optional CLI args, run:

```sh
❯ ./nrversions -help

This tool is used to generate a document detailing the modules that
the newrelic Node.js agent instruments and the version ranges of those
modules.

The following flags are supported:

  -n, --no-externals           Disable cloning and processing of external repos. An external repo is
                               one that provides extra functionality to the "newrelic" module. This
                               allows processing a single repo with --repo-dir. The default, i.e. not
                               supplying this flag, is to process all known external repos.
                               
  -o, --output-format string   Specify the format to write the results as. The default is an ASCII
                               table. Possible values: "ascii" or "markdown".
                                (default "markdown")
  -r, --repo-dir string        Specify a local directory that contains a Node.js instrumentation repo.
                               If not provided, the main agent GitHub repository will be cloned to a
                               local temporary directory and that will be used.
                               
  -t, --test-dir string        Specify the test directory to parse the package.json files.
                               If not provided, it will default to 'test/versioned'. This applies to
                               the repo provided by the --repo-dir flag.
                                
  -v, --verbose                Enable verbose output. As the data is being loaded and parsed various
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
