package updater

import (
	"testing"
)

func TestParsePullOutput_NoUpdates(t *testing.T) {
	output := `
Pulling db ... done
Pulling redis ... done
Status: Image is up to date for mariadb:10.6
Status: Image is up to date for redis:7-alpine
`
	updates := ParsePullOutput(output)
	if len(updates) != 0 {
		t.Errorf("Expected 0 updates, got %d", len(updates))
	}
}

func TestParsePullOutput_WithUpdates(t *testing.T) {
	output := `
Pulling db ... done
Status: Downloaded newer image for mariadb:10.6
Pulling redis ... done
Status: Downloaded newer image for redis:7-alpine
`
	updates := ParsePullOutput(output)
	if len(updates) != 2 {
		t.Errorf("Expected 2 updates, got %d", len(updates))
	}

	// Check first update
	if updates[0].Image != "mariadb:10.6" {
		t.Errorf("Expected image mariadb:10.6, got %s", updates[0].Image)
	}
	if !updates[0].Updated {
		t.Error("Expected update to be marked as updated")
	}

	// Check second update
	if updates[1].Image != "redis:7-alpine" {
		t.Errorf("Expected image redis:7-alpine, got %s", updates[1].Image)
	}
	if !updates[1].Updated {
		t.Error("Expected update to be marked as updated")
	}
}

func TestParsePullOutput_WithDigests(t *testing.T) {
	output := `
Pulling db ... done
Digest: sha256:abc123def456789
Status: Downloaded newer image for mariadb:10.6
`
	updates := ParsePullOutput(output)
	if len(updates) == 0 {
		t.Fatal("Expected at least 1 update")
	}

	// Verify the image was detected
	if updates[0].Image != "mariadb:10.6" {
		t.Errorf("Expected image mariadb:10.6, got %s", updates[0].Image)
	}
}

func TestParsePullOutput_EmptyOutput(t *testing.T) {
	output := ""
	updates := ParsePullOutput(output)
	if len(updates) != 0 {
		t.Errorf("Expected 0 updates for empty output, got %d", len(updates))
	}
}

func TestParsePullOutput_RealWorldExample(t *testing.T) {
	output := `
Pulling kkengine (kkengine/kkengine:latest)...
latest: Pulling from kkengine/kkengine
Digest: sha256:a1b2c3d4e5f6
Status: Downloaded newer image for kkengine/kkengine:latest
Pulling db (mariadb:10.6)...
10.6: Pulling from library/mariadb
Digest: sha256:1234567890ab
Status: Image is up to date for mariadb:10.6
Pulling redis (redis:7-alpine)...
7-alpine: Pulling from library/redis
Digest: sha256:fedcba098765
Status: Downloaded newer image for redis:7-alpine
`
	updates := ParsePullOutput(output)

	// Should have 2 updates (kkengine and redis, not mariadb since it's up to date)
	if len(updates) != 2 {
		t.Errorf("Expected 2 updates, got %d", len(updates))
	}
}
