package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/cast"
)

type VersionedTestPackageJson struct {
	Name    string            `json:"name"`
	Version string            `json:"version"`
	Private bool              `json:"private"`
	Tests   []TestDescription `json:"tests"`
}

type TestDescription struct {
	Engines      EnginesBlock      `json:"engines"`
	Dependencies DependenciesBlock `json:"dependencies"`
	Files        FilesBlock        `json:"files"`
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
			var block DependencyBlock
			err = json.Unmarshal(*val, &block)
			if err != nil {
				return fmt.Errorf("failed to parse dependency block: %w", err)
			}
			(*db)[key] = block
		} else {
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
	err := json.Unmarshal(data, &decoded)
	if err != nil {
		return fmt.Errorf("failed to decode dependency block: %w", err)
	}

	samples := []byte(*decoded["samples"])
	if samples != nil {
		if bytes.Compare(samples[0:1], []byte(`"`)) == 0 {
			strSamples := string(samples[1 : len(samples)-1])
			db.Samples = cast.ToInt(strSamples)
		} else {
			db.Samples = cast.ToInt(string(samples))
		}
	}

	strVersions := string(*decoded["versions"])
	db.Versions = strVersions[1 : len(strVersions)-1]

	return nil
}
