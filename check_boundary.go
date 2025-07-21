Yes, you can **absolutely implement this without changing `main.go`**, by following this refined plan:

---

### ✅ Goal

* Keep the **audience-role mapping** in a config file ✅
* Load it **internally inside the `token` package** (no changes to `main.go`) ✅
* Keep the rest of the code untouched — just validate after token decoding ✅

---

### ✅ Updated Plan (Minimal Changes, Fully Isolated)

1. Create a config file: `audience_roles.yaml`
2. Load the config from within the `token` package (once, lazily)
3. Implement the role validation in `token.ValidateRolesForAudience`
4. Call that validation from `extauth.Check()` (you’re already modifying that part)

---

### 🗂️ 1. `audience_roles.yaml`

Same as before:

```yaml
audience_roles:
  JPMC:URI:RS-112081-139591-Atlas2PBBeta-PROD:
    - PORTFOLIO_BOUNDARY_RESOURCE_MANAGER
    - PORTFOLIO_BOUNDARY_RESOURCE_READER
  JPMC:URI:RS-112081-139592-Atlas2NBBeta-PROD:
    - NETWORK_BOUNDARY_RESOURCE_MANAGER
    - NETWORK_BOUNDARY_RESOURCE_READER
  JPMC:URI:RS-90467-141027-Atlas2RWBBeta-PROD:
    - RETAIL_WORKLOAD_BOUNDARY_RESOURCE_MANAGER
    - RETAIL_WORKLOAD_BOUNDARY_RESOURCE_READER
  JPMC:URI:RS-112081-150439-Atlas2WWBBeta-PROD:
    - WHOLESALE_BOUNDARY_RESOURCE_MANAGER
    - WHOLESALE_BOUNDARY_RESOURCE_READER
```

---

### 🧠 2. Inside `token/config.go` — Lazy YAML Loader

```go
package token

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v2"
)

type audienceRoleConfig struct {
	AudienceRoles map[string][]string `yaml:"audience_roles"`
}

var (
	once              sync.Once
	audienceRoles     map[string][]string
	loadConfigErr     error
	configFilePath    = "/etc/authnz/audience_roles.yaml" // Or hardcode based on your deployment
)

func getAudienceRoleMap() (map[string][]string, error) {
	once.Do(func() {
		absPath, _ := filepath.Abs(configFilePath)
		data, err := os.ReadFile(absPath)
		if err != nil {
			loadConfigErr = fmt.Errorf("failed to read audience-role config: %w", err)
			return
		}

		var config audienceRoleConfig
		err = yaml.Unmarshal(data, &config)
		if err != nil {
			loadConfigErr = fmt.Errorf("failed to parse audience-role config: %w", err)
			return
		}

		audienceRoles = config.AudienceRoles
	})
	return audienceRoles, loadConfigErr
}
```

---

### ✅ 3. In `token.go` — Role Validation Logic

```go
func ValidateRolesForAudience(aud string, roles []string) error {
	roleMap, err := getAudienceRoleMap()
	if err != nil {
		return errors.NewInternalError(errors.InvalidAudience, fmt.Sprintf("audience-role config error: %v", err))
	}

	allowed, ok := roleMap[aud]
	if !ok {
		return errors.NewInternalError(errors.InvalidAudience, fmt.Sprintf("audience %s is not recognized", aud))
	}

	allowedSet := make(map[string]struct{})
	for _, role := range allowed {
		allowedSet[role] = struct{}{}
	}

	for _, r := range roles {
		if _, ok := allowedSet[r]; !ok {
			return errors.NewInternalError(errors.RoleAudienceMismatch, fmt.Sprintf("role %s is not allowed for audience %s", r, aud))
		}
	}

	return nil
}
```

---

### 🧩 4. In `extauth.go` — Add This One Line

This is the **only integration point**:

```go
// Already done earlier
aud := jwtDecodedValues["aud"].GetStringValue()

// Get roles from depscopeToRoleMap
var allRoles []string
for _, roles := range depscopeToRoleMap {
    allRoles = append(allRoles, roles...)
}

// Add this validation
if err := token.ValidateRolesForAudience(aud, allRoles); err != nil {
    logger.Error().Msgf("audience-role validation failed: %v", err)
    return errorResponseGenerator("audience-role validation failed:", http.StatusForbidden, err), nil
}
```

---

### ✅ Outcome

* ✔ No changes to `main.go`
* ✔ Configurable via YAML
* ✔ Automatically loaded once, internally
* ✔ Easy to update audience-role mappings without code changes

---

Would you like me to give you a test YAML and a mock unit test to validate this flow too?
