package adapter

import (
	"fmt"
	"os"

	"github.com/layer5io/meshkit/utils"
)

// previousComponentVersion selects a copy source while excluding
// the current target, including partially generated directories.
func previousComponentVersion(root, target string) (string, error) {
	entries, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read component versions: %w", err)
	}

	var versions []string

	for _, entry := range entries {
		if !entry.IsDir() ||
			entry.Name() == target ||
			!componentFilePart.MatchString(entry.Name()) {
			continue
		}

		versions = append(versions, entry.Name())
	}

	if len(versions) == 0 {
		return "", nil
	}

	versions = utils.SortDottedStringsByDigits(versions)

	return versions[len(versions)-1], nil
}
