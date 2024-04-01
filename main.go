package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/jedib0t/go-pretty/v6/table"
	flag "github.com/spf13/pflag"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

const nrRepo = `https://github.com/newrelic/node-newrelic.git`

type dirIterChan struct {
	name string
	pkg  *VersionedTestPackageJson
	err  error
}

// TODO: use npm api to fully construct https://newrelic.atlassian.net/wiki/spaces/INST/pages/3269656658/Node+Compatibility+Research
type resultsTableRow struct {
	name                string
	minSupportedVersion string
}

func main() {
	err := run(os.Args)
	if err != nil {
		fmt.Printf("app error: %v", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	err := createAndParseFlags(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	var repoDir string
	if flags.repoDir != "" {
		repoDir = flags.repoDir
	} else {
		rd, err := cloneRepo()
		if err != nil {
			return err
		}
		repoDir = rd
	}

	fmt.Println("Processing data ...")
	versionedTestsDir := filepath.Join(repoDir, "test", "versioned")

	iterChan := make(chan dirIterChan)
	go iterateTestDir(versionedTestsDir, iterChan)

	// The issue I have with this approach is that we don't get a single error
	// that we can terminate the program with.
	outputTable := table.NewWriter()
	outputTable.AppendHeader(table.Row{"Name", "MinVersion"})
	for result := range iterChan {
		if result.err != nil {
			fmt.Println(result.err)
			continue
		}

		pkgInfo, err := parsePackage(result.pkg)
		if err != nil {
			if errors.Is(err, ErrTargetMissing) {
				continue
			}
			return err
		}
		outputTable.AppendRow(table.Row{pkgInfo.Name, pkgInfo.MinVersion})
	}

	outputTable.SortBy([]table.SortBy{{Name: "Name", Mode: table.Asc}})
	fmt.Println(outputTable.Render())

	return nil
}

func iterateTestDir(dir string, iterChan chan dirIterChan) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		iterChan <- dirIterChan{err: fmt.Errorf("failed to read directory `%s`: %w", dir, err)}
		close(iterChan)
		return
	}

	for _, entry := range entries {
		testDir := filepath.Join(dir, entry.Name())
		pkgJsonFilePath := filepath.Join(dir, entry.Name(), "package.json")
		pkgJsonFile, err := os.Open(pkgJsonFilePath)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				// Some versioned tests, e.g. the "restify" tests, have subdirectories
				// that split the tests into multiple suites. We determine this by
				// recognizing that a `testdir/package.json` does not exist.
				iterateTestDir(testDir, iterChan)
				return
			}
			iterChan <- dirIterChan{
				name: entry.Name(),
				err:  fmt.Errorf("could not find package.json in `%s`: %w", testDir, err),
			}
			continue
		}

		pkg, err := readPackageJson(pkgJsonFile)
		if err != nil {
			iterChan <- dirIterChan{
				name: entry.Name(),
				err:  fmt.Errorf("failed to read package.json for `%s`: %w", entry.Name(), err),
			}
			continue
		}
		iterChan <- dirIterChan{name: entry.Name(), pkg: pkg}
	}

	close(iterChan)
}

func readPackageJson(pkgJsonFile *os.File) (*VersionedTestPackageJson, error) {
	data, err := io.ReadAll(pkgJsonFile)
	if err != nil {
		return nil, err
	}

	var vtpj VersionedTestPackageJson
	err = json.Unmarshal(data, &vtpj)
	if err != nil {
		return nil, err
	}

	return &vtpj, nil
}

func cloneRepo() (string, error) {
	repoDir, err := os.MkdirTemp("", "newrelic")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(repoDir)

	fmt.Println("Cloning repository ...")
	_, err = git.PlainClone(repoDir, false, &git.CloneOptions{
		URL:   nrRepo,
		Depth: 1,
	})
	if err != nil {
		return "", fmt.Errorf("failed to clone newrelic repo: %w", err)
	}

	return repoDir, nil
}
