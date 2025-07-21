
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

func (tokenSuite *TokenTestSuite) TestValidateToken() {
	tests := []struct {
		name           string
		aud            string
		roles          []string
		expectedResult bool
	}{
		{
			name:           "Valid audience and allowed role",
			aud:            "ds-ceb-123-v1",
			roles:          []string{"NETWORK_BOUNDARY_RESOURCE_MANAGER"},
			expectedResult: true,
		},
		{
			name:           "Valid audience but disallowed role",
			aud:            "ds-ceb-123-v1",
			roles:          []string{"SOME_OTHER_ROLE"},
			expectedResult: false,
		},
		{
			name:           "Invalid audience",
			aud:            "unknown-env",
			roles:          []string{"NETWORK_BOUNDARY_RESOURCE_MANAGER"},
			expectedResult: false,
		},
		{
			name:           "Multiple roles - one valid",
			aud:            "ds-ceb-123-v1",
			roles:          []string{"RANDOM_ROLE", "NETWORK_BOUNDARY_RESOURCE_READER"},
			expectedResult: true,
		},
		{
			name:           "Empty roles list",
			aud:            "ds-ceb-123-v1",
			roles:          []string{},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		tokenSuite.T().Run(tt.name, func(t *testing.T) {
			valid := token.Validate(tt.aud, tt.roles)
			assert.Equal(t, tt.expectedResult, valid)
		})
	}
}
