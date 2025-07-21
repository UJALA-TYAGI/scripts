
## ✅ Quick Summary of Your Goal

In `extauth.go`, during request processing, **you want to reject requests where:**

* `aud` and `scopedRoles` in the JWT token do not align with the allowed audience-role mapping.

You already have:

* Token parsing and extraction logic in `token.go`
* `decodeToken()` that gives you structured token fields like `aud`, `role`, etc.
* You want the validation **only once**, and efficiently.

---

## 🎯 Recommendation: **Call `ValidateAudienceRoleMapping()` from `token.go` inside `extauth.go`**

### Why?

| Criteria                          | Validate inside `token.go` | Validate inline in `extauth.go` |
| --------------------------------- | -------------------------- | ------------------------------- |
| Centralized logic (DRY)           | ✅ Yes                      | ❌ Duplicates logic              |
| Reuse in other services           | ✅ Yes                      | ❌ No                            |
| Ease of testing and unit coverage | ✅ Yes (clean func)         | ❌ Mixed with request logic      |
| Separation of concerns            | ✅ Token validation only    | ❌ Breaks modular boundaries     |
| Simplicity in `extauth.go`        | ✅ Yes                      | ❌ More cluttered code           |
| Config loaded once (singleton)    | ✅ Already handled          | ❌ Would need re-setup           |

---

## ✅ What You Should Do

### Step 1: In `token.go`, export the validator

Make sure `ValidateAudienceRoleMapping` is **exported**:

```go
func ValidateAudienceRoleMapping(token map[string]string) error { ... }
```

---

### Step 2: In `extauth.go`, extract `aud` and `scopedRoles` from `jwtDecodedValues`

```go
aud, audOk := jwtDecodedValues["aud"]
scopedRoles, roleOk := jwtDecodedValues["ScopedRoles"]
```

> Assuming `jwtDecodedValues` is a map of JWT fields.

---

### Step 3: Convert to `map[string]string` and call the validator

```go
if audOk && roleOk {
    tokenMap := map[string]string{
        "aud":  aud.GetStringValue(),
        "role": scopedRoles.GetStringValue(),
    }

    if err := token.ValidateAudienceRoleMapping(tokenMap); err != nil {
        customErr := errors.New(
            errors.RoleNotPermittedForAudience,
            fmt.Sprintf("audience-role mismatch: %v", err),
            errors.RoleNotPermittedForAudienceSolution,
        )
        logger.Error().Msgf("audience-role validation failed: %v", customErr)
        return errorResponseGenerator("authorization denied", http.StatusForbidden, customErr), nil
    }
}
```

---

### 🔄 Where to Place This in `extauth.go`:

Just **after extracting `jwtDecodedValues`** and **before calling `GroupByDepScope`**, so that you block early:

```go
// Extract aud and role from token
aud, audOk := jwtDecodedValues["aud"]
scopedRoles, roleOk := jwtDecodedValues["ScopedRoles"]

if audOk && roleOk {
    tokenMap := map[string]string{
        "aud":  aud.GetStringValue(),
        "role": scopedRoles.GetStringValue(),
    }

    if err := token.ValidateAudienceRoleMapping(tokenMap); err != nil {
        customErr := errors.New(
            errors.RoleNotPermittedForAudience,
            fmt.Sprintf("audience-role mismatch: %v", err),
            errors.RoleNotPermittedForAudienceSolution,
        )
        logger.Error().Msgf("audience-role validation failed: %v", customErr)
        return errorResponseGenerator("authorization denied", http.StatusForbidden, customErr), nil
    }
}
```

---

package token_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"your_project/pkg/token"
	"your_project/pkg/mapper"
	"your_project/internal/pkg/utils/errors"
)

type TokenTestSuite struct {
	suite.Suite
}

func TestTokenSuite(t *testing.T) {
	suite.Run(t, new(TokenTestSuite))
}

func (s *TokenTestSuite) SetupTest() {
	mockMapping := map[string][]string{
		"JPMC:URI:RS-000001-000001-TestEnv1-PROD": {
			"PORTFOLIO_BOUNDARY_RESOURCE_MANAGER",
			"PORTFOLIO_BOUNDARY_RESOURCE_READER",
		},
		"JPMC:URI:R5-000002-000002-TestEnv2-PROD": {
			"NETWORK_BOUNDARY_RESOURCE_MANAGER",
			"NETWORK_BOUNDARY_RESOURCE_READER",
		},
		"JPMC:URI:R5-000003-000003-TestEnv3-PROD": {
			"RETAIL_WORKLOAD_BOUNDARY_RESOURCE_MANAGER",
			"RETAIL_WORKLOAD_BOUNDARY_RESOURCE_READER",
		},
		"JPMC:URI:R5-000004-000004-TestEnv4-PROD": {
			"WHOLESALE_BOUNDARY_RESOURCE_MANAGER",
			"WHOLESALE_BOUNDARY_RESOURCE_READER",
		},
	}
	mapper.SetTestMapping(mockMapping)
}

func (s *TokenTestSuite) TestValidateAudienceRoleMapping() {
	tests := []struct {
		name    string
		aud     string
		roles   []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Valid role for TestEnv1",
			aud:     "JPMC:URI:RS-000001-000001-TestEnv1-PROD",
			roles:   []string{"PORTFOLIO_BOUNDARY_RESOURCE_MANAGER"},
			wantErr: false,
		},
		{
			name:    "Valid role among multiple roles for TestEnv2",
			aud:     "JPMC:URI:R5-000002-000002-TestEnv2-PROD",
			roles:   []string{"SOME_OTHER_ROLE", "NETWORK_BOUNDARY_RESOURCE_READER"},
			wantErr: false,
		},
		{
			name:    "Unknown audience",
			aud:     "JPMC:URI:UNKNOWN-AUDIENCE-PROD",
			roles:   []string{"PORTFOLIO_BOUNDARY_RESOURCE_MANAGER"},
			wantErr: true,
			errMsg:  "not found in config",
		},
		{
			name:    "No permitted role for valid audience",
			aud:     "JPMC:URI:R5-000002-000002-TestEnv2-PROD",
			roles:   []string{"UNAUTHORIZED_ROLE"},
			wantErr: true,
			errMsg:  "not allowed",
		},
		{
			name:    "Valid audience but empty roles list",
			aud:     "JPMC:URI:RS-000001-000001-TestEnv1-PROD",
			roles:   []string{},
			wantErr: true,
			errMsg:  "not allowed",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := token.ValidateAudienceRoleMapping(tt.aud, tt.roles)
			if tt.wantErr {
				s.Error(err)
				if tt.errMsg != "" {
					s.Contains(err.Error(), tt.errMsg)
				}
			} else {
				s.NoError(err)
			}
		})
	}
}



