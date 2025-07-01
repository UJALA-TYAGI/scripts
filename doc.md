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
