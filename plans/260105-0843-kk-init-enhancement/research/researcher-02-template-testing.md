# Research Report: Go text/template Testing Best Practices

## Executive Summary
Testing Go templates requires comprehensive validation of rendering output, syntax correctness, and edge cases. Best practices include table-driven tests with fixtures, golden file comparison, and programmatic validation of generated configs (TOML, YAML). For embedded templates (`go:embed`), use test helpers that parse embedded FS and validate all template combinations.

## Research Methodology
- Sources consulted: Go documentation, testing patterns, validation tools
- Focus: Template rendering validation, config syntax checking, test fixtures

## Key Findings

### 1. Testing Template Rendering
**Table-driven tests**: Define test cases with various Config inputs and expected outputs.

```go
func TestRenderTemplate(t *testing.T) {
    tests := []struct {
        name string
        cfg  Config
        want string
    }{
        {
            name: "basic config",
            cfg:  Config{Domain: "example.com", DBPassword: "pass123"},
            want: "expected output...",
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var buf bytes.Buffer
            err := RenderTemplate("template", tt.cfg, &buf)
            if err != nil {
                t.Fatal(err)
            }
            if got := buf.String(); got != tt.want {
                t.Errorf("got %q, want %q", got, tt.want)
            }
        })
    }
}
```

### 2. Golden File Approach
Store expected outputs in `testdata/golden/` directory:
- `testdata/golden/Caddyfile.golden`
- `testdata/golden/kkfiler.toml.golden`

Compare rendered output against golden files:

```go
func TestRenderGolden(t *testing.T) {
    cfg := Config{Domain: "test.com"}
    var buf bytes.Buffer
    RenderTemplate("Caddyfile", cfg, &buf)

    golden := filepath.Join("testdata", "golden", "Caddyfile.golden")
    want, _ := os.ReadFile(golden)

    if diff := cmp.Diff(string(want), buf.String()); diff != "" {
        t.Errorf("mismatch (-want +got):\n%s", diff)
    }
}
```

### 3. Validation Tools

**TOML validation** (for kkfiler.toml):
```go
import "github.com/BurntSushi/toml"

func ValidateTOML(content string) error {
    var v interface{}
    _, err := toml.Decode(content, &v)
    return err
}
```

**YAML validation** (for docker-compose.yml):
```go
import "gopkg.in/yaml.v3"

func ValidateYAML(content string) error {
    var v interface{}
    return yaml.Unmarshal([]byte(content), &v)
}
```

**Caddyfile validation**:
- Use `github.com/caddyserver/caddy/v2/caddyconfig/caddyfile` adapter
- Or simple syntax checks (braces matching, directive validation)

### 4. Testing Embedded Templates

```go
//go:embed *.tmpl
var templateFS embed.FS

func TestAllTemplatesExist(t *testing.T) {
    required := []string{
        "Caddyfile.tmpl",
        "kkfiler.toml.tmpl",
        "kkphp.conf.tmpl",
        "docker-compose.yml.tmpl",
        "env.tmpl",
    }

    for _, name := range required {
        _, err := templateFS.ReadFile(name)
        if err != nil {
            t.Errorf("template %s not found: %v", name, err)
        }
    }
}
```

### 5. Config Combinations Testing

Test all combinations of EnableSeaweedFS and EnableCaddy:

```go
func TestAllCombinations(t *testing.T) {
    combinations := []struct {
        seaweed bool
        caddy   bool
    }{
        {false, false},
        {true, false},
        {false, true},
        {true, true},
    }

    for _, combo := range combinations {
        cfg := Config{
            EnableSeaweedFS: combo.seaweed,
            EnableCaddy:     combo.caddy,
            // ... other fields
        }

        // Test docker-compose.yml renders correctly
        // Test .env renders correctly
        // Test optional files render only when enabled
    }
}
```

## Implementation Recommendations for kkcli

1. **Test Structure**:
```
pkg/templates/
├── *.tmpl
├── embed.go
├── embed_test.go
└── testdata/
    ├── golden/
    │   ├── Caddyfile.golden
    │   ├── kkfiler.toml.golden
    │   └── docker-compose.yml.golden
    └── fixtures/
        └── config.go (test configs)
```

2. **Test Coverage**:
   - ✅ All templates exist and are parseable
   - ✅ All Config combinations render without error
   - ✅ Generated YAML/TOML/Caddyfile syntax is valid
   - ✅ Template variables are correctly substituted
   - ✅ Conditional rendering (SeaweedFS/Caddy) works
   - ✅ File permissions are correctly set (.env = 0600)

3. **Validation Libraries**:
   - `gopkg.in/yaml.v3` - YAML validation (already in go.mod)
   - `github.com/BurntSushi/toml` - TOML validation
   - Custom Caddyfile parser or regex-based validation

4. **CI Integration**:
   - Run template tests in GitHub Actions
   - Fail build if template syntax invalid
   - Use golden file updates on breaking changes

## Sources
- Go text/template docs: https://pkg.go.dev/text/template
- Testing embedded files: https://pkg.go.dev/embed
- Table-driven tests: https://dave.cheney.net/2019/05/07/prefer-table-driven-tests
- Golden files: https://github.com/sebdah/goldie

Unresolved questions: None.
