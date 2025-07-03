# Detailed Kubernetes API Flow for kubectl get on Custom Resources

## Purpose

This document provides a comprehensive overview of the Kubernetes API server endpoints that are internally called when performing a <code>kubectl get</code> on any Custom Resource (CR). It covers the full discovery and retrieval process, explaining each API endpoint, its role, request/response details, and related considerations.

This is intended for:

- Developers working with Kubernetes APIs  
- Platform engineers managing API gateways or proxies (e.g., Envoy)  
- Troubleshooters debugging CLI or API access issues  
- API consumers wanting to understand backend calls.

## Prerequisites & Assumptions

- The resource being fetched is a **namespaced Custom Resource** defined via a CRD.  
- Kubernetes cluster supports the standard API discovery endpoints.  
- The user running <code>kubectl</code> has necessary RBAC permissions for discovery and resource reading.

## Overview of API Call Sequence

The <code>kubectl get [resource] -n [namespace]</code> command triggers a series of REST calls to the Kubernetes API server:

| Step | Endpoint Path                                                        | Purpose                                                          |
|-------|--------------------------------------------------------------------|------------------------------------------------------------------|
| 1     | /openapi/v2                                                        | Retrieve OpenAPI schema for all registered API resources         |
| 2     | /api                                                               | List core API groups and versions (like v1)                      |
| 3     | /apis                                                              | List all named API groups (including CRDs)                       |
| 4     | /apis/[group]/[version]                                            | Retrieve resource metadata for a specific group and version      |
| 5     | /apis/[group]/[version]/namespaces/[namespace]/[resource]   | List actual custom resource instances in the specified namespace |


] For **cluster-scoped** resources, the ```/namespaces/[namespace]``` segment is omitted.

## Detailed Explanation of Each Step

### 1. GET /openapi/v2

- **Description**: Returns the OpenAPI v2 schema that describes all resource types (built-in and custom) supported by the API server.  
- **Role**: Enables the `kubectl` client to know resource field structures, validation rules, and output formatting details.  
- **Typical Request Headers**: Includes authentication headers, content type `application/json`.
- **Response Type**: A JSON object conforming to the OpenAPI v2 (Swagger 2.0) specification.
  - Contains `swagger` version string (usually "2.0")
  - `info` metadata about the API
  - `paths` describing all available API endpoints and their supported HTTP methods
  - `definitions` listing all resource schemas (built-in and custom) with their fields, types, validation, and descriptions
- **Response Content**: Large JSON schema describing every resource's fields, types, descriptions.  
- **Notes**:  
  - This request is **optional** and skipped if `kubectl` is run with explicit output format flags like `-o json`.  
  - Failure to get this may cause `kubectl` to fail with errors like `unknown type`.
 
### 2. GET /api

- **Description**: Lists all the **core API groups** (usually just `v1`).
- **Role**: Part of API discovery — confirms availability of core resources.
- **Response Example**: A JSON object listing core API versions.

### 3. GET /apis

- **Description**: Lists all registered named API groups, including:  
  - Kubernetes built-in groups like `metrics.k8s.io`, `networking.k8s.io`.  
  - All CRD groups for e.g. `portfolio.wholesale.atlas.aws.jpmchase.net`.  
- **Role**: Enables the client to map resource short names to full API group and version.  
- **Response Example**: A JSON list of API groups with their versions and preferred versions.

### 4. GET /apis/[group]/[version]

- **Description**: Retrieves detailed metadata about all resource types available in the specified API group and version.  
- **Role**: Used by `kubectl` to:  
  - Confirm the plural name of the resource (e.g., `awsportfolioboundaries`).  
  - Determine if the resource is namespaced or cluster-scoped.  
  - Identify supported operations (`verbs`) such as `get`, `list`, `watch`, `create`, `patch`, `delete`.  
- **Response Example**: JSON list of resource types, including kind, namespaced flag, supported verbs, and short names.

### 5. GET /apis/[group]/[version]/namespaces/[namespace]/[resource]

- **Description**: Fetches the actual list of custom resource instances in the specified namespace.  
- **Role**: Returns the resource data that `kubectl get` will display.  
- **Response**: JSON object containing `items` array with resource instances, each with metadata and spec fields.

## Permissions and RBAC

- The user or service account making the request **must have RBAC permissions** for:  
  - `get`, `list`, and `watch` verbs on the custom resource.  
  - `get` on API discovery endpoints like `/apis` and `/api`.  
- Without correct permissions, the requests will fail with `403 Forbidden`.

## Debugging and Observability

- Use `kubectl get [resource] -n [namespace] --v=9` to view detailed API request logs.  
- Use API server audit logs (if enabled) to trace the calls.  
- For network debugging, check proxy logs (e.g., Envoy) to verify allowed paths and routing.  
- Check for errors in:  
  - OpenAPI schema fetch (`/openapi/v2`)  
  - API group discovery (`/apis`, `/api`)  
  - Resource metadata fetch (`/apis/[group]/[version]`)  
  - Resource list fetch (`/apis/.../namespaces/...`)  

## Additional Notes

- For **cluster-scoped** custom resources, remove the `/namespaces/[namespace]` segment in step 5.  
- If `kubectl` is run with output flags (`-o json`, `-o yaml`), the OpenAPI call (`/openapi/v2`) is usually skipped.  
- Custom resources can have multiple versions; discovery via `/apis/[group]` informs which versions exist.  
- If API aggregation or API extension servers are present, similar discovery calls occur for those.

## Summary Table

| Step | API Endpoint                                              | Description                                 |
|-------|------------------------------------------------------------|---------------------------------------------|
| 1     | /openapi/v2                                               | Fetch OpenAPI schemas for output formatting |
| 2     | /api                                                      | List core API groups (e.g., v1)             |
| 3     | /apis                                                     | List all named API groups, including CRDs   |
| 4     | /apis/[group]/[version]                                 | Get resource metadata for the group/version |
| 5     | /apis/[group]/[version]/namespaces/[namespace]/[resource] | List custom resource instances in namespace |

## Conclusion

The seemingly simple `kubectl get` command relies on a chain of internal Kubernetes API calls to:

- Discover API groups and resource versions.  
- Retrieve resource metadata and schema.  
- Fetch the actual custom resource instances.

Understanding this flow helps in:

- Properly configuring API proxies, ingress, and security controls.  
- Debugging `kubectl` CLI or permission issues.  
- Optimizing API server interactions.  
- Building reliable tooling and dashboards that interact with Kubernetes resources.
- 

# Kubernetes API Documentation

This document serves as a comprehensive guide to interact with resources in the Atlas2.0 environment. It explains all available API calls in detail, including what each call does, when to use it, how to use it, required headers, expected payloads, and the format of the response.

## 🔐 1. Authentication

All API calls require authentication via a Bearer token.

Required Header:
- Authorization: Bearer <your_token_here>

Tokens must be securely stored and rotated periodically as per your security policy.

## 📦 2. Supported Resource Types

This documentation applies to all internal AWS boundaries managed as Kubernetes custom resources. These include:

| Resource Kind         | API Group                                                   | Version  |
|----------------------|-------------------------------------------------------------|----------|
| AWSPortfolioBoundary | portfolio.wholesale.atlas.aws.jpmchase.net                 | v1alpha1 |
| AWSWorkloadBoundary  | workload.wholesale.atlas.aws.jpmchase.net                  | v1alpha1 |
| AWSNetworkBoundary   | network.wholesale.atlas.aws.jpmchase.net                   | v1alpha1 |

All operations described below are applicable to any of the above resources. For demonstration purposes, examples will use the AWSPortfolioBoundary resource.

## 🔁 3. General API Operations

> Note: Each API operation listed below includes an expected example response to help understand the return structure and how to parse or validate the output.

### A. List Resources

Purpose: Retrieve all existing resources of a kind in a given namespace.
Method: GET
URL:
```{{api_url}}/{{ingress_path}}/apis/<api_group>/v1alpha1/namespaces/{{namespace}}/<resource_kind>```
Headers:
- Authorization: Bearer <token>

Expected Response (200):
```{ "apiVersion": "v1", "items": [ { "metadata": { "name": "example-boundary", "namespace": "ds-test" }, "spec": { ... } } ], "kind": "List", "metadata": { "resourceVersion": "123456" } }```

### B. Get a Resource by Name

Purpose: Fetch a single resource by its name.
Method: GET
URL:
```{{api_url}}/{{ingress_path}}/apis/<api_group>/v1alpha1/namespaces/{{namespace}}/<resource_kind>/<resource_name>```

Expected Response (200):
```{ "apiVersion": "portfolio.wholesale.atlas.aws.jpmchase.net/v1alpha1", "kind": "AWSPortfolioBoundary", "metadata": { "name": "example-boundary", "namespace": "ds-test" }, "spec": { ... }, "status": { "conditions": [ { "type": "Ready", "status": "True" } ] } }```

### C. Create a Resource

Purpose: Create a new custom resource in the given namespace.
Method: POST
URL:
```{{api_url}}/{{ingress_path}}/apis/<api_group>/v1alpha1/namespaces/{{namespace}}/<resource_kind>```
Headers:
- Authorization: Bearer <token>
- Content-Type: application/yaml
Optional Query Param: ?dryRun=All (for validation only)

### D. Update a Resource

Purpose: Replace the full definition of an existing resource.
Method: PUT
URL:
```{{api_url}}/{{ingress_path}}/apis/<api_group>/v1alpha1/namespaces/{{namespace}}/<resource_kind>/<resource_name>```
Headers:
- Authorization: Bearer <token>
- Content-Type: application/yaml

### E. Patch a Resource

Purpose: Partially update the resource specification.
Method: PATCH
URL:
```{{api_url}}/{{ingress_path}}/apis/<api_group>/v1alpha1/namespaces/{{namespace}}/<resource_kind>/<resource_name>```
Headers:
- Authorization: Bearer <token>
- Content-Type: application/merge-patch+json

### F. Delete a Resource

Purpose: Remove the resource from the cluster.
Method: DELETE
URL:
```{{api_url}}/{{ingress_path}}/apis/<api_group>/v1alpha1/namespaces/{{namespace}}/<resource_kind>/<resource_name>```
Headers:
- Authorization: Bearer <token>

### G. Watch Resource Events

Purpose: Stream changes to the resource in real-time.
Method: GET
URL:
```{{api_url}}/{{ingress_path}}/apis/<api_group>/v1alpha1/watch/namespaces/{{namespace}}/<resource_kind>```
Headers:
- Authorization: Bearer <token>

## 🧾 4. Example Payload (AWSPortfolioBoundary)

```apiVersion: portfolio.wholesale.atlas.aws.jpmchase.net/v1alpha1
kind: AWSPortfolioBoundary
metadata:
  name: "example-boundary"
  namespace: "ds-test"
spec:
  deploymentScope:
    jrn: "jrn:jpm:iep-test:::depscope:example-jrn"
  enabledRegions:
    - "us-east-1"
  orgUnitName: "example-org-unit"
  parent:
    jrn: "jrn:jpm:atlas-dev:example-parent"
  SCPS:
    - jrn: "jrn:jpm:atlas-dev:example-scp"
  trustedscopes:
    - name: example-trust
      type: Portfolio
      attributes:
        - key: jpmc-sealappid
          value: "112081"
        - key: jpmc-sealdepid
          value: "115004"
  satisfiedcontrols:
    pci: "jrn:jpm:atlas:::pci:cat3"
    dataclassification: "jrn:jpm:atlas:::dataclass:confidential"
    connectivity: "jrn:jpm:atlas:::connect:internal"
    soc: "jrn:jpm:atlas:::soc:none"
    hitrust: "jrn:jpm:atlas:::hitrust:none"
    jurisdiction: "jrn:jpm:atlas:::jur:default"
```

## 📥 5. Example Responses

Successful Create Response (201)
```{ "apiVersion": "portfolio.wholesale.atlas.aws.jpmchase.net/v1alpha1", "kind": "AWSPortfolioBoundary", "metadata": { "name": "example-boundary", "namespace": "ds-test", "uid": "abc12345-6789-0123-defg-hijk45678901", "creationTimestamp": "2025-07-01T10:30:00Z" }, "spec": { ... }, "status": { "conditions": [ { "type": "Ready", "status": "True", "lastUpdateTime": "2025-07-01T10:30:05Z" } ] } }```

Error Response (422)
```{ "kind": "Status", "apiVersion": "v1", "metadata": {}, "status": "Failure", "message": "AWSPortfolioBoundary 'example-boundary' is invalid: spec.enabledRegions: Required value", "reason": "Invalid", "code": 422 }```

## ❗ 6. Error Handling

| HTTP Code | Meaning               | Action                                 |
|-----------|-----------------------|----------------------------------------|
| 401       | Unauthorized          | Check Bearer token                     |
| 403       | Forbidden             | You do not have permissions            |
| 404       | Not Found             | Resource doesn’t exist                 |
| 409       | Conflict              | Duplicate name or versioning issue     |
| 422       | Unprocessable Entity  | Payload validation failed              |
| 500       | Internal Server Error | Check logs or report to admins         |

## ✅ Final Notes

- Replace <api_group> and <resource_kind> for different resources:
  - Portfolio: portfolio.wholesale.atlas.aws.jpmchase.net, awsportfolioboundaries
  - Workload: workload.wholesale.atlas.aws.jpmchase.net, awsworkloadboundaries
  - Network: network.wholesale.atlas.aws.jpmchase.net, awsnetworkboundaries
- Always validate YAML against CRDs.
- Use dryRun=All before applying changes.
