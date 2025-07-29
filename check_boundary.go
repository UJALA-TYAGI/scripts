
### 🛤️ Options Considered

During this refactor, we evaluated two primary approaches for decoupling authorization from strict path-based matching:

#### Option 1: **Move Validation to Authz Webhook** ✅ *(Chosen Approach)*

* Perform audience validation in the centralized `authz` webhook instead of Envoy.
* The webhook checks token claims and request metadata to determine access.
* Results in a **cleaner Envoy configuration** and **centralized access logic**.
* Slightly increases request hops (Envoy → webhook), but offers flexibility and maintainability.
* Scales better across CRDs and supports more dynamic routing needs.

#### Option 2: **Enumerate Full URL Paths in CRDs for Envoy Match**

* Maintain validation within Envoy by mapping **specific URL paths to audience values** at the CRD level (rather than grouping by API prefix).
* Would require onboarding **each CRD individually** and updating Envoy rules accordingly.
* Reduces reliance on the webhook, potentially minimizing incorrect audience-resource mismatches.
* But significantly increases **templating complexity** and is harder to manage and evolve.

---

### 📌 Final Decision

We are proceeding with **Option 1: validation logic in the centralized `authz` webhook**.
This ensures **scalability, reduced configuration burden, and centralized governance**.
