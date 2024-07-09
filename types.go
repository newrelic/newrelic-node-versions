package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/cast"
)

type nrRepo struct {
	isMainRepo bool
	repoDir    string
	url        string
	branch     string
	testPath   string
}

type dirIterChan struct {
	name string
	pkg  *VersionedTestPackageJson
	err  error
}

// CloneRepoResult represents the status of Git repository clone operation.
type CloneRepoResult struct {
	// IsMainRepo indicates if the clone represents the mainline Node.js Agent
	// repository. The mainline repo includes extra configuration that is needed
	// by the tool. Since cloning happens concurrently, we don't know which
	// of the results is the mainline repo simply by index.
	IsMainRepo bool

	// Directory is the path on the file system that contains the cloned
	// repository.
	Directory string

	// TestDirectory is a string relative to Directory that contains the
	// versioned tests for the repository.
	TestDirectory string

	// Remove indicates if the Directory should be removed after all data
	// processing has completed.
	Remove bool

	// Error indicates if there was some problem during the clone operation.
	// Should be `nil` for success results.
	Error error
}

// ReleaseData represents a row of information about a package. Specifically,
// it's the final computed information to be rendered into documents.
type ReleaseData struct {
	Name                       string
	MinSupportedVersion        string
	MinSupportedVersionRelease string
	LatestVersion              string
	LatestVersionRelease       string
	MinAgentVersion            string
}

type Target struct {
	Name            string `json:"name"`
	MinAgentVersion string `json:"minAgentVersion"`
}

type VersionedTestPackageJson struct {
	Name    string            `json:"name"`
	Targets []Target          `json:"targets"`
	Version string            `json:"version"`
	Private bool              `json:"private"`
	Tests   []TestDescription `json:"tests"`
}

type TestDescription struct {
	Supported    bool              `json:"supported"`
	Comment      string            `json:"comment"`
	Engines      EnginesBlock      `json:"engines"`
	Dependencies DependenciesBlock `json:"dependencies"`
	Files        FilesBlock        `json:"files"`
}

func (td *TestDescription) UnmarshalJSON(data []byte) error {
	// We need to manually decode the test description block because we do not
	// want the `supported` key to default to `false` (the normal zero value for a
	// bool). Basically, we want to assume all test descriptors are for a
	// supported version unless explicitly indicated otherwise.
	if bytes.Compare(data, []byte("null")) == 0 {
		return nil
	}

	var rawObj map[string]*json.RawMessage
	_ = json.Unmarshal(data, &rawObj)

	*td = TestDescription{
		Supported: true,
	}
	for key, val := range rawObj {
		switch key {
		case "supported":
			switch string(*val) {
			case "false":
				td.Supported = false
			case "true":
				td.Supported = true
			}
		case "comment":
			var comment string
			err := json.Unmarshal(*val, &comment)
			if err != nil {
				return err
			}
			td.Comment = comment
		case "engines":
			var engines EnginesBlock
			err := json.Unmarshal(*val, &engines)
			if err != nil {
				return err
			}
			td.Engines = engines
		case "dependencies":
			var deps DependenciesBlock
			err := json.Unmarshal(*val, &deps)
			if err != nil {
				return err
			}
			td.Dependencies = deps
		case "files":
			var files FilesBlock
			err := json.Unmarshal(*val, &files)
			if err != nil {
				return err
			}
			td.Files = files
		}
	}

	return nil
}

type EnginesBlock struct {
	Node string `json:"node"`
}

type FilesBlock []string

type DependenciesBlock map[string]DependencyBlock

type DependencyBlock struct {
	Versions string `json:"versions"`
	Samples  int    `json:"samples"`
}

func (db *DependenciesBlock) UnmarshalJSON(data []byte) error {
	if bytes.Compare(data, []byte("null")) == 0 {
		return nil
	}

	var rawObj map[string]*json.RawMessage
	err := json.Unmarshal(data, &rawObj)
	if err != nil {
		return err
	}

	*db = make(map[string]DependencyBlock)
	for key, val := range rawObj {
		if bytes.Compare((*val)[0:1], []byte("{")) == 0 {
			// Parse a full dependency block directly.
			var block DependencyBlock
			err = json.Unmarshal(*val, &block)
			if err != nil {
				return fmt.Errorf("failed to parse dependency block: %w", err)
			}
			(*db)[key] = block
		} else {
			// Otherwise, convert a simple version string into a full
			// dependency block.
			strVal := string(*val)
			(*db)[key] = DependencyBlock{Versions: strVal[1 : len(strVal)-1]}
		}
	}

	return nil
}

func (db *DependencyBlock) UnmarshalJSON(data []byte) error {
	if bytes.Compare(data, []byte("null")) == 0 {
		return nil
	}

	var decoded map[string]*json.RawMessage
	_ = json.Unmarshal(data, &decoded)

	samples := decoded["samples"]
	if samples != nil {
		samplesBytes := []byte(*samples)
		if bytes.Compare(samplesBytes[0:1], []byte(`"`)) == 0 {
			strSamples := string(samplesBytes[1 : len(samplesBytes)-1])
			db.Samples = cast.ToInt(strSamples)
		} else {
			db.Samples = cast.ToInt(string(samplesBytes))
		}
	}

	versions := decoded["versions"]
	if versions == nil {
		return fmt.Errorf("missing versions property: %s", data)
	}
	strVersions := string(*versions)
	db.Versions = strVersions[1 : len(strVersions)-1]

	return nil
}
