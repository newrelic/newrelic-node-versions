package main

import (
	"blitznote.com/src/semver/v3"
	"errors"
	"fmt"
)

var ErrTargetMissing = errors.New("target not found in dependencies list")

type PkgInfo struct {
	Name       string
	MinVersion string
}

func parsePackage(pkg *VersionedTestPackageJson) ([]PkgInfo, error) {
	var lastVersion *semver.Range
	targets := pkg.Targets

	results := make([]PkgInfo, 0)
	for _, target := range targets {
		for _, test := range pkg.Tests {
			if test.Supported == false {
				continue
			}

			for key, val := range test.Dependencies {
				if key != target {
					continue
				}

				currentVersion, err := semver.NewRange([]byte(val.Versions))
				if err != nil {
					return nil, fmt.Errorf("failed to parse version string `%s` for `%s`: %w", val.Versions, targets, err)
				}

				if lastVersion == nil {
					lastVersion = &currentVersion
					continue
				}

				if currentVersion.GetLowerBoundary().Less(*(lastVersion.GetLowerBoundary())) == true {
					lastVersion = &currentVersion
				}
			}
		}

		if lastVersion == nil {
			return nil, fmt.Errorf("%s: %w", pkg.Name, ErrTargetMissing)
		}

		pkgInfo := PkgInfo{
			Name:       target,
			MinVersion: lastVersion.GetLowerBoundary().String(),
		}
		results = append(results, pkgInfo)
	}

	return results, nil
}
