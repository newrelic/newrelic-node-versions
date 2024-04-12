package main

import (
	"blitznote.com/src/semver/v3"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/go-git/go-git/v5"
	"github.com/jedib0t/go-pretty/v6/table"
	flag "github.com/spf13/pflag"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sync"
)

const nrRepo = `https://github.com/newrelic/node-newrelic.git`

type dirIterChan struct {
	name string
	pkg  *VersionedTestPackageJson
	err  error
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

	logger := buildLogger(flags.verbose)

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

	logger.Debug("Processing data ...")
	versionedTestsDir := filepath.Join(repoDir, "test", "versioned")

	iterChan := make(chan dirIterChan)
	go iterateTestDir(versionedTestsDir, iterChan)

	npm := NewNpmClient()
	wg := sync.WaitGroup{}
	data := make([]ReleaseData, 0)
	for result := range iterChan {
		if result.err != nil {
			logger.Error(result.err.Error())
			continue
		}

		pkgInfos, err := parsePackage(result.pkg)
		if err != nil {
			if errors.Is(err, ErrTargetMissing) {
				logger.Debug(err.Error())
				continue
			}
			return err
		}

		// TODO: handle errors better. Probably refactor into something like the dirIter goroutine
		for _, info := range pkgInfos {
			wg.Add(1)
			go func(info PkgInfo) {
				defer wg.Done()
				logger.Debug("getting detailed package info", "package", info.Name)
				releaseData, err := buildReleaseData(info, npm)
				if err != nil {
					logger.Error(err.Error())
					return
				}
				data = append(data, *releaseData)
			}(info)
		}
	}

	wg.Wait()

	slices.SortFunc(data, func(a ReleaseData, b ReleaseData) int {
		if a.Name == b.Name {
			return 0
		}
		switch a.Name > b.Name {
		case true:
			return 1
		default:
			return -1
		}
	})
	prunedData := pruneData(data)
	switch flags.outputFormat.String() {
	default:
		renderAsAscii(prunedData, os.Stdout)
	case "ascii":
		renderAsAscii(prunedData, os.Stdout)
	case "markdown":
		renderAsMarkdown(prunedData, os.Stdout)
	}

	return nil
}

func buildLogger(verbose bool) *slog.Logger {
	if verbose == true {
		return slog.New(
			// TODO: replace with https://github.com/dusted-go/logging/issues/3
			slog.NewTextHandler(
				os.Stderr,
				&slog.HandlerOptions{Level: slog.LevelDebug},
			),
		)
	}
	return slog.New(
		slog.NewTextHandler(
			os.Stderr,
			&slog.HandlerOptions{Level: slog.LevelError},
		),
	)
}

func buildReleaseData(info PkgInfo, npm *NpmClient) (*ReleaseData, error) {
	latest, err := npm.GetLatest(info.Name)
	if err != nil {
		return nil, err
	}

	detailedInfo, err := npm.GetDetailedInfo(info.Name)
	if err != nil {
		return nil, err
	}

	minReleaseDate := detailedInfo.Time[info.MinVersion]
	latestReleaseDate := detailedInfo.Time[latest]

	result := &ReleaseData{
		Name:                       info.Name,
		MinSupportedVersion:        info.MinVersion,
		MinSupportedVersionRelease: minReleaseDate.ToFullDate().ToString(),
		LatestVersion:              latest,
		LatestVersionRelease:       latestReleaseDate.ToFullDate().ToString(),
	}

	return result, nil
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

// pruneData removes duplicate entries from the data set. A duplicate entry
// is one in which the [ReleaseData.Name] is equal. The entry with the lowest
// [ReleaseData.MinSupportedVersion] will be kept. The data should be sorted
// prior to being pruned.
func pruneData(data []ReleaseData) []ReleaseData {
	result := make([]ReleaseData, 0)
	for i := 0; i < len(data); {
		if i == len(data)-1 {
			break
		}

		a := data[i]
		b := data[i+1]
		if a.Name != b.Name {
			result = append(result, a)
			i += 1
			continue
		}

		verA, _ := semver.NewVersion([]byte(a.MinSupportedVersion))
		verB, _ := semver.NewVersion([]byte(b.MinSupportedVersion))
		if verA.Less(verB) {
			result = append(result, a)
		} else {
			result = append(result, b)
		}
		i += 2
	}

	return result
}

// renderAsAscii renders the collected data as an ASCII table. This is
// intended to be used when generating local CLI output during testing.
func renderAsAscii(data []ReleaseData, writer io.Writer) {
	outputTable := releaseDataToTable(data)
	io.WriteString(writer, outputTable.Render())
}

func renderAsMarkdown(data []ReleaseData, writer io.Writer) {
	outputTable := releaseDataToTable(data)
	io.WriteString(
		writer,
		heredoc.Docf(`
			## Instrumented Modules

			The following table lists the modules that the %[1]newrelic%[1] Node.js
			agent instruments, along with the minimum version of the module the agent
			supports, the release date of that minimum version, and the version plus
			release date of the most recent version (as of the time this document
			was generated).
		`, "`"),
	)
	io.WriteString(writer, "\n")
	io.WriteString(writer, outputTable.RenderMarkdown())
}

func releaseDataToTable(data []ReleaseData) table.Writer {
	outputTable := table.NewWriter()

	keys := make([]string, 0)
	header := table.Row{}
	rv := reflect.ValueOf(ReleaseData{})
	rt := rv.Type()
	for i := 0; i < rv.NumField(); i += 1 {
		header = append(header, rt.Field(i).Name)
		keys = append(keys, rt.Field(i).Name)
	}
	outputTable.AppendHeader(header)

	for _, info := range data {
		row := table.Row{}
		rv = reflect.ValueOf(info)
		for _, key := range keys {
			row = append(row, rv.FieldByName(key).Interface())
		}
		outputTable.AppendRow(row)
	}

	return outputTable
}
