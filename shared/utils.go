package shared

import (
	"strconv"
	"strings"
)

type Version struct {
	Major int
	Minor int
	Patch int
}

// AtLeastHaVersion checks if the version is at least the given major, minor, and patch.
func AtLeastHaVersion(version string, major, minor int, patch ...int) bool {
	versions := strings.Split(version, ".")

	haMajor, _ := strconv.Atoi(versions[0])
	haMinor, _ := strconv.Atoi(versions[1])
	haPatch, _ := strconv.Atoi(versions[2])

	if len(patch) == 0 {
		return haMajor > major || (haMajor == major && haMinor >= minor)
	}

	p := patch[0]

	return haMajor > major ||
		(haMajor == major && haMinor > minor) ||
		(haMajor == major && haMinor == minor && haPatch >= p)
}
