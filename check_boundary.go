✅ Refactoring Path-Based Authorization in Atlas2

### 🧠 Background

Atlas2 **authentication and authorization is performed via JWT**, managed at the Envoy layer using filters and external auth integrations. Previously, each service boundary (e.g., WWB, NB, CN) had its own match rule based on the **API path prefix**, and requests were routed through different **authorizers** accordingly.

This setup resulted in:

* A **tight coupling between API path and authorization logic**.
* A **complex and repetitive Envoy configuration**, with each boundary needing its own `provider_name`, path match, and validation logic.
* Difficulty supporting **optional feature boundaries**, where APIs do not follow strict URL segregation.

---

### 📌 Objective

To **decouple path-based authorization logic from the Envoy layer** and move it into the centralized `authz` webhook. This enables more flexible request handling and simplifies the overall configuration.

---

### ✂️ What Was Changed

#### 🔴 Removed:

* All **factory-specific authorizers** that enforced service-boundary logic based on API paths at the Envoy level.

  ```yaml
  {{- range $key, $val := .Values.envoy.factories }}
  - match:
      prefix: {{ $val.api_group }}
    requires:
      provider_name: {{ $val.provider_name }}-authorizer
  {{- end }}
  ```

* Associated `provider_name` definitions and OIDC configuration for each factory.

#### ✅ Retained:

* A **single default authorizer** applied to all requests:

  ```yaml
  - match:
      prefix: /
    requires:
      provider_name: default-authorizer
  ```

---

### 🚚 Where Did the Authorization Logic Go?

The **path-specific authorization logic** was shifted into the **Authorization Webhook (`authz`)**.
This allows all requests to be processed centrally, and the webhook can now enforce the correct boundary rules based on request metadata instead of URL path.

---

### 🎯 Benefits of This Refactor

* ✅ **Cleaner Envoy config** — reduces complexity and duplication.
* ✅ **Centralized logic** — authorization decisions are made in one place, simplifying updates and audits.
* ✅ **Increased flexibility** — removes dependency on strict URL path structures.
* ✅ **Supports future extensibility** — enables dynamic routing and optional feature boundaries.
