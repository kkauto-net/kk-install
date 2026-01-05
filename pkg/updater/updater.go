package updater

import (
	"regexp"
	"strings"
)

type ImageUpdate struct {
	Image     string
	OldDigest string
	NewDigest string
	Updated   bool
}

// ParsePullOutput parses docker-compose pull output
// Example output lines:
//
//	Pulling db ... done
//	Pulling redis ... downloading
//	kkengine Pulled
//	Status: Downloaded newer image for mariadb:10.6
func ParsePullOutput(output string) []ImageUpdate {
	var updates []ImageUpdate

	// Pattern for "Downloaded newer image"
	newerPattern := regexp.MustCompile(`Downloaded newer image for (.+)`)

	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check for "newer image" pattern - this is the most reliable indicator
		if matches := newerPattern.FindStringSubmatch(line); len(matches) > 1 {
			updates = append(updates, ImageUpdate{
				Image:   matches[1],
				Updated: true,
			})
		}
	}

	return updates
}
