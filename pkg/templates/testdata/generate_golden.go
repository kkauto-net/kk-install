// +build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kkauto-net/kk-install/pkg/templates"
)

func main() {
	cfg := templates.Config{
		EnableSeaweedFS: true,
		EnableCaddy:     true,
		DBPassword:      "test_db_pass",
		DBRootPassword:  "test_db_root_pass",
		RedisPassword:   "test_redis_pass",
		Domain:          "example.com",
	}

	goldenDir := "golden"
	os.MkdirAll(goldenDir, 0755)

	// Generate docker-compose.yml.golden
	err := templates.RenderTemplate("docker-compose.yml", cfg, filepath.Join(goldenDir, "docker-compose.yml.golden"))
	if err != nil {
		fmt.Printf("Error rendering docker-compose.yml: %v\n", err)
		os.Exit(1)
	}

	// Generate env.golden
	err = templates.RenderTemplate("env", cfg, filepath.Join(goldenDir, "env.golden"))
	if err != nil {
		fmt.Printf("Error rendering env: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Golden files generated successfully!")
}
