package policy

import (
	"errors"
	"fmt"
	"github.com/keel-hq/keel/util/version"
	"strings"

	"github.com/Masterminds/semver"
)

// SemverPolicyType - policy type
type SemverPolicyType int

var (
	ErrNoMajorMinorPatchElementsFound = errors.New("No Major.Minor.Patch elements found")
)

// available policies
const (
	SemverPolicyTypeNone SemverPolicyType = iota
	SemverPolicyTypeAll
	SemverPolicyTypeMajor
	SemverPolicyTypeMinor
	SemverPolicyTypePatch
)

func (t SemverPolicyType) String() string {
	switch t {
	case SemverPolicyTypeNone:
		return "none"
	case SemverPolicyTypeAll:
		return "all"
	case SemverPolicyTypeMajor:
		return "major"
	case SemverPolicyTypeMinor:
		return "minor"
	case SemverPolicyTypePatch:
		return "patch"
	default:
		return ""
	}
}

func NewSemverPolicy(spt SemverPolicyType) *SemverPolicy {
	return &SemverPolicy{
		spt: spt,
	}
}

type SemverPolicy struct {
	spt SemverPolicyType
}

func (sp *SemverPolicy) ShouldUpdate(current, new string) (bool, error) {
	return shouldUpdate(sp.spt, current, new)
}

func (sp *SemverPolicy) Name() string {
	return sp.spt.String()
}

func (sp *SemverPolicy) Type() PolicyType { return PolicyTypeSemver }

func shouldUpdate(spt SemverPolicyType, current, new string) (bool, error) {
	if current == "latest" {
		return true, nil
	}

	currentSumoModifiedSemver := version.GetSemverFromSumoVersion(current)
	newSumoModifiedSemver := version.GetSemverFromSumoVersion(new)

	parts := strings.SplitN(newSumoModifiedSemver, ".", 3)
	if len(parts) != 3 {
		return false, ErrNoMajorMinorPatchElementsFound
	}

	currentVersion, err := semver.NewVersion(currentSumoModifiedSemver)
	if err != nil {
		return false, fmt.Errorf("failed to parse current version: %s", err)
	}

	newVersion, err := semver.NewVersion(newSumoModifiedSemver)
	if err != nil {
		return false, fmt.Errorf("failed to parse new version: %s", err)
	}

	if currentVersion.Prerelease() != newVersion.Prerelease() && spt != SemverPolicyTypeAll && !version.IsSumoVersion(new){
		return false, nil
	}

	// new version is not higher than current - do nothing
	if !currentVersion.LessThan(newVersion) {
		return false, nil
	}

	switch spt {
	case SemverPolicyTypeAll, SemverPolicyTypeMajor:
		return true, nil
	case SemverPolicyTypeMinor:
		return newVersion.Major() == currentVersion.Major(), nil
	case SemverPolicyTypePatch:
		return newVersion.Major() == currentVersion.Major() && newVersion.Minor() == currentVersion.Minor(), nil
	}
	return false, nil
}
