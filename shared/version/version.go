package version

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Version represents a semantic version with Major, Minor, and Patch components.
type Version struct {
	major int
	minor int
	patch int
}

// MarshalJSON converts the Version to a JSON string in the format "Major.Minor.Patch".
func (v Version) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

// UnmarshalJSON unmarshals a JSON string into a Version.
func (v *Version) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("failed to unmarshal version: %w", err)
	}

	parsedVersion, err := Parse(str)
	if err != nil {
		return err
	}

	*v = parsedVersion

	return nil
}

// String converts the Version to a string in "Major.Minor.Patch" format.
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.major, v.minor, v.patch)
}

// Parse parses a version string in "Major.Minor.Patch" format and returns a new Version.
func Parse(version string) (Version, error) {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid version: %q, expected format Major.Minor.Patch", version)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return Version{}, fmt.Errorf("invalid major version: %w", err)
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return Version{}, fmt.Errorf("invalid minor version: %w", err)
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return Version{}, fmt.Errorf("invalid patch version: %w", err)
	}

	return Version{major: major, minor: minor, patch: patch}, nil
}

// Minimum checks if the version is at least the given major, minor, and patch.
// Minor and Patch are optional; if omitted, they default to 0.
func (v Version) Minimum(major int, versions ...int) bool {
	var minor, patch int

	if len(versions) > 0 {
		minor = versions[0]
	}

	if len(versions) > 1 {
		patch = versions[1]
	}

	return v.major > major ||
		(v.major == major && v.minor > minor) ||
		(v.major == major && v.minor == minor && v.patch >= patch)
}
