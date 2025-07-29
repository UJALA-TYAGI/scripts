
## ✅ Refactoring Audience-Based Authorization

### 🧠 Background

Atlas2 **authentication and authorization is performed via JWT**, managed at the Envoy layer using filters and external auth integrations. Each service boundary (e.g., WWB, NB, CN) previously had its own audience defined within Envoy, and requests were routed through **separate authorizers** depending on the path.

This setup resulted in:

* A **tight coupling between URL path and authorization logic**.
* A **complex and repetitive Envoy configuration**, with each boundary needing its own `provider_name`, `audience`, and match block.
* Challenges with **optional feature boundaries**, where requests across APIs do not necessarily follow a rigid path structure.

---

### 📌 Objective

To **decouple audience validation from the Envoy layer** and shift it into the centralized `authz` webhook. This makes the system more flexible, maintainable, and aligned with future needs like dynamic routing and optional resource boundaries.

---

### ✂️ What Was Changed

#### 🔴 Removed:

* All **factory-specific authorizers** that enforced boundary-based audiences at the Envoy level.

  ```yaml
  {{- range $key, $val := .Values.envoy.factories }}
  - match:
      prefix: {{ $val.api_group }}
    requires:
      provider_name: {{ $val.provider_name }}-authorizer
  {{- end }}
  ```

* Associated Envoy config for individual audiences and remote JWKS.

#### ✅ Retained:

* A **single default authorizer** for all API requests:

  ```yaml
  - match:
      prefix: /
    requires:
      provider_name: default-authorizer
  ```

---

### 🚚 Where Did the Audience Validation Go?

The audience validation logic was **shifted to the Authorization Webhook (`authz`)**.
This allows validation to be performed in **application logic** instead of relying on hardcoded path-based rules in Envoy.

---

### 🎯 Benefits of This Refactor

* ✅ **Cleaner Envoy configuration** — reduces duplication and simplifies maintenance.
* ✅ **Flexible routing support** — removes tight coupling between API paths and audiences.
* ✅ **Centralized logic** — easier to evolve, test, and extend from one place.
* ✅ **Future readiness** — paves the way for optional and dynamic boundary support.
