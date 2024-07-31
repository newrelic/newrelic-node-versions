package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dusted-go/logging/prettylog"
	"github.com/spf13/afero"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync"

	_ "embed"

	"blitznote.com/src/semver/v3"
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jedib0t/go-pretty/v6/table"
	flag "github.com/spf13/pflag"
)

//go:embed tmpl/preamble.md
var docPreamble string

var agentRepo = nrRepo{
	url:        `https://github.com/newrelic/node-newrelic.git`,
	branch:     `main`,
	testPath:   `test/versioned`,
	isMainRepo: true,
}
var externalsRepos = []nrRepo{
	{url: `https://github.com/newrelic/newrelic-node-apollo-server-plugin.git`, branch: `main`, testPath: `tests/versioned`},
}

var columHeaders = map[string]string{
	"Name":                `Package name`,
	"MinSupportedVersion": `Minimum supported version`,
	"LatestVersion":       `Latest supported version`,
	"MinAgentVersion":     `Introduced in*`,
}

var appFS = afero.NewOsFs()

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

	var repos []nrRepo
	if flags.repoDir != "" {
		var testDir string
		repoDir := flags.repoDir
		if flags.testDir != "" {
			testDir = flags.testDir
		} else {
			testDir = "test/versioned"
		}
		var testRepo = nrRepo{repoDir: repoDir, testPath: testDir, isMainRepo: true}
		repos = []nrRepo{testRepo}
	} else {
		repos = []nrRepo{agentRepo}
	}

	if flags.noExternals == false {
		repos = append(repos, externalsRepos...)
	}

	logger.Info("cloning repositories")
	cloneResults := cloneRepos(repos, logger)
	logger.Info("repository cloning complete")

	testDirs := make([]string, 0)
	for _, cloneResult := range cloneResults {
		if cloneResult.Error != nil {
			logger.Error(cloneResult.Error.Error())
			continue
		}

		versionedTestsDir := filepath.Join(cloneResult.Directory, cloneResult.TestDirectory)
		logger.Debug("adding test dir", "dir", versionedTestsDir)
		testDirs = append(testDirs, versionedTestsDir)
	}

	logger.Info("processing data")
	defer func() {
		cleanupTempDirs(cloneResults, logger)
	}()
	data := processVersionedTestDirs(testDirs, logger)

	aiCompatInputFile := flags.aiCompatJsonFile
	mainRepoClone := cloneResults[slices.IndexFunc(cloneResults, func(s CloneRepoResult) bool {
		return s.IsMainRepo == true
	})]
	if aiCompatInputFile == "" {
		aiCompatInputFile = path.Join(mainRepoClone.Directory, "ai-support.json")
	}
	aiCompatDoc := strings.Builder{}
	err = RenderAiCompatDoc(aiCompatInputFile, &aiCompatDoc)
	if err != nil {
		return fmt.Errorf("failed to process ai compat doc: %w", err)
	}
	logger.Info("data processing complete")

	var writeDest io.Writer
	if flags.replaceInFile != "" {
		writeDest = &strings.Builder{}
	} else {
		writeDest = os.Stdout
	}

	slices.SortFunc(data, releaseDataSorter)
	prunedData := pruneData(data)
	renderAsMarkdown(prunedData, writeDest)
	io.WriteString(writeDest, "\n"+aiCompatDoc.String())

	if flags.replaceInFile != "" {
		content := writeDest.(*strings.Builder).String()
		err = ReplaceInFile(flags.replaceInFile, content, flags.startMarker, flags.endMarker)
		if err != nil {
			return err
		}
	}

	logger.Info("done")

	return nil
}

// cleanupTempDirs removes any temporary directories marked for removal that
// were created during cloning.
func cleanupTempDirs(cloneResults []CloneRepoResult, logger *slog.Logger) {
	for _, cloneResult := range cloneResults {
		if cloneResult.Remove == false {
			logger.Debug("not removing directory " + cloneResult.Directory)
			continue
		}
		logger.Debug("removing directory " + cloneResult.Directory)
		_ = appFS.RemoveAll(cloneResult.Directory)
	}
}

// processVersionedTestDirs iterates through all versioned test directories,
// looking for versioned `package.json` files, and processes what it finds
// into release data for each found supported module.
func processVersionedTestDirs(testDirs []string, logger *slog.Logger) []ReleaseData {
	wg := sync.WaitGroup{}
	results := make([]ReleaseData, 0)

	for _, versionedTestsDir := range testDirs {
		iterChan := make(chan dirIterChan)
		go iterateTestDir(versionedTestsDir, iterChan)

		npm := NewNpmClient(WithLogger(logger))
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

				// TODO: this was a hard exit. How should we handle catastrophic errors?
				logger.Error(err.Error())
				continue
			}

			for _, info := range pkgInfos {
				wg.Add(1)
				go func(info PkgInfo) {
					defer wg.Done()
					releaseData, err := buildReleaseData(info, npm)
					if err != nil {
						logger.Error(err.Error())
						return
					}
					results = append(results, *releaseData)
				}(info)
			}
		}
	}

	wg.Wait()
	return results
}

func buildLogger(verbose bool) *slog.Logger {
	dest := prettylog.WithDestinationWriter(os.Stderr)
	level := slog.LevelInfo

	if verbose == true {
		level = slog.LevelDebug
	}

	handler := prettylog.New(
		&slog.HandlerOptions{Level: level},
		dest,
	)

	return slog.New(handler)
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
		MinAgentVersion:            info.MinAgentVersion,
	}

	return result, nil
}

func openPackageJson(path string) (*os.File, error) {
	pkgJsonFilePath := filepath.Join(path, "package.json")
	return os.Open(pkgJsonFilePath)
}

func iterateTestDir(dir string, iterChan chan dirIterChan) {
	// try to parse package.json from root
	// in this case it is the only result
	// so end early
	rootPkgJson, _ := openPackageJson(dir)
	if rootPkgJson != nil {
		pkg, err := readPackageJson(rootPkgJson)
		if err != nil {
			iterChan <- dirIterChan{
				name: dir,
				err:  fmt.Errorf("failed to read package.json for `%s`: %w", dir, err),
			}
		}
		iterChan <- dirIterChan{name: dir, pkg: pkg}
		close(iterChan)
		return
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		iterChan <- dirIterChan{err: fmt.Errorf("failed to read directory `%s`: %w", dir, err)}
		close(iterChan)
		return
	}

	for _, entry := range entries {
		testDir := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			pkgJsonFile, err := openPackageJson(testDir)
			if err != nil {
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
	}

	close(iterChan)
}

// readPackageJson reads a file as a versioned `package.json`.
func readPackageJson(pkgJsonFile io.Reader) (*VersionedTestPackageJson, error) {
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

// cloneRepos clones multiple repositories at once but does not return until
// all repositories have been cloned.
func cloneRepos(repos []nrRepo, logger *slog.Logger) []CloneRepoResult {
	wg := sync.WaitGroup{}
	result := make([]CloneRepoResult, 0)
	for _, repo := range repos {
		wg.Add(1)
		go func(r nrRepo) {
			defer wg.Done()
			cloneResult := cloneRepo(r, logger)
			cloneResult.IsMainRepo = r.isMainRepo
			result = append(result, cloneResult)
		}(repo)
	}
	wg.Wait()
	return result
}

// cloneRepo clones a remote repository. If a local directory is specified
// in the repo description, then cloning is skipped and only a result object
// is returned.
func cloneRepo(repo nrRepo, logger *slog.Logger) CloneRepoResult {
	if repo.repoDir != "" {
		return CloneRepoResult{
			Directory:     repo.repoDir,
			TestDirectory: repo.testPath,
			Remove:        false,
		}
	}

	repoDir, err := afero.TempDir(appFS, "", "newrelic")
	if err != nil {
		return CloneRepoResult{
			Error: fmt.Errorf("failed to create temporary directory: %w", err),
		}
	}

	logger.Debug("cloning repo", "url", repo.url)
	_, err = git.PlainClone(repoDir, false, &git.CloneOptions{
		URL:           repo.url,
		ReferenceName: plumbing.ReferenceName(repo.branch),
		Depth:         1,
	})
	if err != nil {
		return CloneRepoResult{
			Error: fmt.Errorf("failed to clone repo `%s`: %w", repo.url, err),
		}
	}

	return CloneRepoResult{
		Directory:     repoDir,
		TestDirectory: repo.testPath,
		Remove:        true,
	}
}

// releaseDataSorter is a sorting function for [ReleaseData]. It is meant to
// be used by [slices.SortFunc].
func releaseDataSorter(a ReleaseData, b ReleaseData) int {
	if a.Name == b.Name {
		return 0
	}
	switch a.Name > b.Name {
	case true:
		return 1
	default:
		return -1
	}
}

// pruneData removes duplicate entries from the data set. A duplicate entry
// is one in which the [ReleaseData.Name] is equal. The entry with the lowest
// [ReleaseData.MinSupportedVersion] will be kept. The data should be sorted
// prior to being pruned.
func pruneData(data []ReleaseData) []ReleaseData {
	result := make([]ReleaseData, 0)
	for i := 0; i < len(data); {
		if i == len(data)-1 {
			// only one result, just assign to result
			result = append(result, data[i])
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

// renderAsMarkdown renders the collected data as a Markdown table. This is
// intended to be used when generating output to be embedded in one of docs
// locations (or maybe to be fed into pandoc to generate a PDF in order to
// email it to a customer).
func renderAsMarkdown(data []ReleaseData, writer io.Writer) {
	outputTable := releaseDataToTable(data)
	io.WriteString(writer, docPreamble)
	io.WriteString(writer, "\n")
	io.WriteString(writer, outputTable.RenderMarkdown())
	io.WriteString(writer, "\n\n")
	io.WriteString(
		writer,
		heredoc.Docf(`
			*When package is not specified, support is within the %snewrelic%s package.
		`, "`", "`"),
	)
}

// releaseDataToTable builds the tabular data structure from the discovered
// supported modules data.
func releaseDataToTable(data []ReleaseData) table.Writer {
	outputTable := table.NewWriter()

	keys := make([]string, 0)
	header := table.Row{}
	rv := reflect.ValueOf(ReleaseData{})
	rt := rv.Type()
	for i := 0; i < rv.NumField(); i += 1 {
		var value = rt.Field(i).Name
		if columnHeader, ok := columHeaders[value]; ok {
			header = append(header, columnHeader)
			keys = append(keys, value)
		}
	}
	outputTable.AppendHeader(header)

	for _, info := range data {
		row := table.Row{}
		rv = reflect.ValueOf(info)
		for _, key := range keys {
			value := rv.FieldByName(key).Interface().(string)
			if key == "Name" {
				value = fmt.Sprintf("`%s`", value)
			} else if key == "MinAgentVersion" && strings.HasPrefix(value, "@") == true {
				value = fmt.Sprintf("`%s`", value)
			}
			row = append(row, value)
		}
		outputTable.AppendRow(row)
	}

	return outputTable
}
