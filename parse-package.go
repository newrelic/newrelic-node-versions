package main

import (
	"blitznote.com/src/semver/v3"
	"errors"
	"fmt"
	"regexp"
)

var ErrTargetMissing = errors.New("target not found in dependencies list")

type PkgInfo struct {
	Name       string
	MinVersion string
}

func parsePackage(pkg *VersionedTestPackageJson) (*PkgInfo, error) {
	var lastVersion *semver.Range
	target := pkg.Target

	// TODO: write tests for other packages and flesh out algorithm
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
				return nil, fmt.Errorf("failed to parse version string `%s` for `%s`: %w", val.Versions, target, err)
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

	pkgInfo := &PkgInfo{
		Name:       target,
		MinVersion: lastVersion.GetLowerBoundary().String(),
	}

	return pkgInfo, nil
}

// isSpecificBelowRange is to test if a specific version string, e.g. `7.13.0`,
// precedes a range string, e.g. `>=7.16.0`. When this case is encountered,
// we want to use the _range_ as the minimum version, not the specific
// version.
func isSpecificBelowRange(specific string, rng string) bool {
	return false
}

// isSpecificVersion determines if a version string specifies a pinned version.
// That is, `1.2.3` returns `true`, but `^1.2.3` returns `false`.
func isSpecificVersion(version string) bool {
	return specificVerRegex.MatchString(version)
}

var specificVerRegex = regexp.MustCompile(`^\d+\.\d+\.\d+$`)
