package adapter

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/layer5io/meshkit/utils"
)

const incompleteGenerationMarker = ".generation-incomplete"

// previousComponentVersion excludes the current target and
// directories marked as having incomplete generation.
// Historical unmarked directories remain eligible.
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

		marker := filepath.Join(
			root, entry.Name(), incompleteGenerationMarker,
		)

		_, err := os.Lstat(marker)
		if err == nil {
			continue
		}
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("inspect generation marker: %w", err)
		}

		versions = append(versions, entry.Name())
	}

	if len(versions) == 0 {
		return "", nil
	}

	versions = utils.SortDottedStringsByDigits(versions)

	return versions[len(versions)-1], nil
}
