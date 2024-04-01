package main

import (
	"blitznote.com/src/semver/v3"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type PkgInfo struct {
	Name       string
	MinVersion string
}

func parsePackage(versionedDirName string, pkg VersionedTestPackageJson) (*PkgInfo, error) {
	var pkgName string
	var lastVersion *semver.Range

	// TODO: write tests for other packages and flesh out algorithm
	for _, test := range pkg.Tests {
		if test.Supported == false {
			continue
		}

		for key, val := range test.Dependencies {
			if depMatchesTarget(versionedDirName, key) == false {
				continue
			}

			if lastVersion == nil {
				v, err := semver.NewRange([]byte(val.Versions))
				if err != nil {
					return nil, err
				}
				lastVersion = &v
				pkgName = key
				continue
			}
		}
	}

	if lastVersion == nil {
		return nil, errors.New("failed to find matching dependency in package.json")
	}

	pkgInfo := &PkgInfo{
		Name:       pkgName,
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

func depMatchesTarget(target string, depName string) bool {
	if target == depName {
		return true
	}

	if strings.HasPrefix(depName, fmt.Sprintf("@%s/", target)) == true {
		return true
	}

	if strings.Contains(depName, target) == true {
		return true
	}

	return false
}

// isSpecificVersion determines if a version string specifies a pinned version.
// That is, `1.2.3` returns `true`, but `^1.2.3` returns `false`.
func isSpecificVersion(version string) bool {
	return specificVerRegex.MatchString(version)
}

var specificVerRegex = regexp.MustCompile(`^\d+\.\d+\.\d+$`)
