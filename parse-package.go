package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"blitznote.com/src/semver/v3"
)

var ErrTargetMissing = errors.New("targets not found in dependencies list")

// Covers strings like `>=1.0.0 < 2`, which should be `>=1.0.0 <2`.
var spaceBeforeDigitRegex = regexp.MustCompile(`([<>=]{1})(\s+)(\d)`)

const max_range = "<=999.999.999"

type PkgInfo struct {
	Name            string
	MinVersion      string
	MinAgentVersion string
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

			// We need to find the minimum version of the target module by looking
			// through the dependencies list and the inspecting the semver range
			// strings associated with it.
			for key, val := range test.Dependencies {
				if key != target.Name {
					continue
				}

				// The semver library does not parse strings like `>1.0.0 <2.0.0 || >3.0.0`.
				// So we need to split it up and normalize the pieces into range strings
				// it can understand.
				var currentVersion semver.Range
				rangeStrings := strings.Split(val.Versions, "||")
				for k, v := range rangeStrings {
					// Oh, Go, why no slices.Map?
					rangeStrings[k] = normalizeRangeString(v)
				}

				if len(rangeStrings) == 1 {
					r, err := semver.NewRange([]byte(rangeStrings[0]))
					if err != nil {
						return nil, fmt.Errorf("failed to parse version string `%s` (from `%s`) for `%s`: %w", rangeStrings[0], val.Versions, targets, err)
					}
					currentVersion = r
				} else {
					ranges := make([]semver.Range, 0)
					for _, rangeString := range rangeStrings {
						r, err := semver.NewRange([]byte(rangeString))
						if err != nil {
							return nil, fmt.Errorf("failed to parse version string `%s` (from `%s`) for `%s`: %w", rangeString, val.Versions, targets, err)
						}
						ranges = append(ranges, r)
					}
					currentVersion = ranges[0]
					for _, r := range ranges[1:] {
						if isRangeLower(r, currentVersion) == true {
							currentVersion = r
						}
					}
				}

				if lastVersion == nil {
					lastVersion = &currentVersion
					continue
				}

				if isRangeLower(currentVersion, *lastVersion) == true {
					lastVersion = &currentVersion
				}
			}
		}

		if lastVersion == nil {
			return nil, fmt.Errorf("%s: %w", pkg.Name, ErrTargetMissing)
		}

		minVersion := lastVersion.GetLowerBoundary()
		if minVersion == nil {
			// This happens when the version is set to "latest" (or "*").
			v, _ := semver.NewVersion([]byte("0.0.0"))
			minVersion = &v
		}

		pkgInfo := PkgInfo{
			Name:            target.Name,
			MinVersion:      minVersion.String(),
			MinAgentVersion: target.MinAgentVersion,
		}
		results = append(results, pkgInfo)
	}

	return results, nil
}

// normalizeRangeString massages range strings into a format that the
// semver library recognizes as a valid range string.
func normalizeRangeString(input string) string {
	result := strings.TrimSpace(input)

	if result == "latest" {
		return max_range
	}

	// Given ">= 4 < 5", normalize to ">=4 <5".
	if spaceBeforeDigitRegex.MatchString(result) {
		result = spaceBeforeDigitRegex.ReplaceAllString(result, "${1}${3}")
	}

	return result
}

// isRangeLower is used to determine if the range `check` represents a range
// with a lower minimum version than the range represented by `against`.
func isRangeLower(check semver.Range, against semver.Range) bool {
	checkLower := check.GetLowerBoundary()
	checkUpper := check.GetUpperBoundary()
	if checkLower == nil && checkUpper == nil {
		// The range seems to have been `*`. So it allows for any version.
		return true
	}

	againstLower := against.GetLowerBoundary()
	againstUpper := against.GetUpperBoundary()
	if againstLower == nil && againstUpper == nil {
		// The range is unbounded. So it covers all version strings.
		// Therefore, check would always be contained by against.
		return false
	}

	if checkLower == nil && againstLower != nil {
		// check has an upper bound, but allows for anything below it.
		// So we want to determine if its upper bound exceeds the lower bound
		// of against.
		return checkUpper.Less(*againstLower)
	}

	if againstLower == nil {
		// There's no lower boundary, e.g. a range like `<=1.0.0`.
		return false
	}

	return checkLower.Less(*againstLower)
}
