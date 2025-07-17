// Package extauth contains functions to process the envoy request and generating corresponding responses.
package extauth

import (
	"context"
	"fmt"
	"net/http"
	"encoding/json"
	"strings"

	envoyauth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	"your/module/path/validation"  // Adjust import path for your validation package
	"bitbucketde-cluster04.jpmchase.net/atlaswpfactories/atlas-controller-authnz/internal/pkg/logging"
	"bitbucketde-cluster04.jpmchase.net/atlaswpfactories/atlas-controller-authnz/internal/pkg/utils/errors"
	"bitbucketde-cluster04.jpmchase.net/atlaswpfactories/atlas-controller-authnz/internal/pkg/bootstrap"
	"bitbucketde-cluster04.jpmchase.net/atlaswpfactories/atlas-controller-authnz/internal/pkg/impersonation"
	"google.golang.org/grpc/codes"
)

type AuthorizationServer struct {
	ClientSet      interface{}      // Use your actual kubernetes.Interface type here
	Bootstrap      *bootstrap.Bootstrap
	DepscopeKeyID  string
}

func New(clientset interface{}, timeout int64, depscopeKeyID string, cachesize int) (*AuthorizationServer, error) {
	namespaceCache, err := bootstrap.NewNamespaceCache(cachesize)
	if err != nil {
		logging.GetLogger(context.Background()).Error().Msgf("Failed to create namespace cache: %v", err)
		return nil, err
	}

	clusterRoles, err := bootstrap.FetchAtlasControllerClusterRoles(clientset)
	if err != nil {
		return nil, err
	}

	return &AuthorizationServer{
		ClientSet: clientset,
		Bootstrap: &bootstrap.Bootstrap{
			ClientSet:      clientset,
			ClusterRoles:   clusterRoles,
			RubeAPITimeout: timeout,
			NamespaceCache: namespaceCache,
		},
		DepscopeKeyID: depscopeKeyID,
	}, nil
}

func (a *AuthorizationServer) Check(ctx context.Context, req *envoyauth.CheckRequest) (*envoyauth.CheckResponse, error) {
	logger := logging.GetLogger(ctx)

	filterMetadata := req.GetAttributes().GetMetadataContext().GetFilterMetadata()
	jwtDecodedValues := filterMetadata["envoy.filters.http-jwt_authn"].GetFields()["jwt-decoded"].GetStructValue().GetFields()

	headers := req.GetAttributes().GetRequest().GetHttp().GetHeaders()
	requestID := headers["x-request-id"]
	ctx = context.WithValue(ctx, logging.KeyCorrelationID, requestID)
	ctx = context.WithValue(ctx, logging.KeyRequestPath, req.GetAttributes().GetRequest().GetHttp().GetPath())
	ctx = logging.InitLogger().WithContext(ctx)
	logger = logging.GetLogger(ctx)

	method := req.GetAttributes().GetRequest().GetHttp().GetMethod()
	path := req.GetAttributes().GetRequest().GetHttp().GetPath()

	// Shortcut for some GET requests
	if method == http.MethodGet && a.Bootstrap.GetPathLength(path) == 4 {
		return &envoyauth.CheckResponse{
			Status: &status.Status{
				Code: int32(codes.OK),
			},
			HttpResponse: &envoyauth.CheckResponse_OkResponse{
				OkResponse: &envoyauth.OkHttpResponse{},
			},
		}, nil
	}

	scopedRolesVal, ok := jwtDecodedValues["ScopedRoles"]
	if !ok {
		customErr := errors.New(
			errors.ScopedRoleMissing,
			"scoped role claim missing in the token",
			errors.ScopedRoleMissingSolution,
		)
		logger.Error().Msgf("an error occurred: %v", customErr)
		return errorResponseGenerator("an error occurred:", http.StatusForbidden, customErr), nil
	}

	depscopeToRoleMap, err := token.GroupByDepScope(ctx, scopedRolesVal, a.DepscopeKeyID)
	if err != nil {
		customErr := errors.New(
			errors.InvalidToken,
			"invalid token in the request",
			errors.InvalidTokenSolution,
		)
		logger.Error().Msgf("an error occurred: %v", customErr)
		return errorResponseGenerator("an error occurred:", http.StatusForbidden, customErr), nil
	}

	namespace, err := a.Bootstrap.NamespaceAndDepScopeFromRequest(ctx, path, depscopeToRoleMap)
	if err != nil {
		return errorResponseGenerator("an error occurred:", http.StatusForbidden, err), nil
	}

	// NEW: Boundary Validation using roles from the token map and the request path
	// Flatten all roles from depscopeToRoleMap into a []string for validation
	allRoles := make([]string, 0)
	for _, roles := range depscopeToRoleMap {
		allRoles = append(allRoles, roles...)
	}

	if err := validation.ValidateBoundaryAccess(path, allRoles); err != nil {
		logger.Error().Msgf("boundary validation failed: %v", err)
		return errorResponseGenerator(fmt.Sprintf("access denied: %v", err), http.StatusForbidden, err), nil
	}

	headersToAdd := impersonation.RequestHeaders(namespace, depscopeToRoleMap)
	logger.Debug().Msgf("headers added: %+v", headersToAdd)
	logger.Info().Msgf("bootstrapping complete for namespace %s", namespace)

	return &envoyauth.CheckResponse{
		Status: &status.Status{
			Code: int32(codes.OK),
		},
		HttpResponse: &envoyauth.CheckResponse_OkResponse{
			OkResponse: &envoyauth.OkHttpResponse{
				Headers: headersToAdd,
			},
		},
	}, nil
}

func errorResponseGenerator(body string, httpResponseCode int, err error) *envoyauth.CheckResponse {
	return &envoyauth.CheckResponse{
		Status: &status.Status{
			Code: int32(codes.Unknown),
		},
		HttpResponse: &envoyauth.CheckResponse_DeniedResponse{
			DeniedResponse: &envoyauth.DeniedHttpResponse{
				Status: &envoyauth.HttpStatus{
					Code: envoyauth.StatusCode(httpResponseCode),
				},
				Body: fmt.Sprintf("%s %v", body, err),
			},
		},
	}
}


_______________
package validation

import (
	"fmt"
	"strings"
)

// BoundaryType represents the API boundaries for role-based access control.
type BoundaryType string

const (
	PortfolioBoundary          BoundaryType = "portfolio"
	NetworkBoundary            BoundaryType = "network"
	WholesaleWorkloadBoundary  BoundaryType = "wholesale_workload"
	WorkloadBoundary           BoundaryType = "workload"
)

// ErrUnauthorizedBoundaryAccess indicates missing required role prefix for access.
type ErrUnauthorizedBoundaryAccess struct {
	RequiredPrefix string
}

func (e *ErrUnauthorizedBoundaryAccess) Error() string {
	return fmt.Sprintf("missing required role with prefix %q", e.RequiredPrefix)
}

// ValidateBoundaryAccess checks if the user's roles authorize access to the boundary
// derived from the request path. Returns nil if authorized, or error if not.
func ValidateBoundaryAccess(path string, roles []string) error {
	boundary := extractBoundaryFromPath(path)

	// If no known boundary is matched, allow access by default.
	if boundary == "" {
		return nil
	}

	expectedPrefix := buildRolePrefix(boundary)

	for _, role := range roles {
		if strings.HasPrefix(role, expectedPrefix) {
			return nil // Authorized
		}
	}

	return &ErrUnauthorizedBoundaryAccess{RequiredPrefix: expectedPrefix}
}

// extractBoundaryFromPath determines boundary type by matching path prefix
// with expected API groups. Next character after prefix (if any) must be
// '.', '/' or end of string for a valid match.
func extractBoundaryFromPath(path string) BoundaryType {
	matchesPrefix := func(prefix string) bool {
		if !strings.HasPrefix(path, prefix) {
			return false
		}
		if len(path) == len(prefix) {
			return true
		}
		nextChar := path[len(prefix)]
		return nextChar == '.' || nextChar == '/'
	}

	switch {
	case matchesPrefix("/apis/portfolio"):
		return PortfolioBoundary
	case matchesPrefix("/apis/network"):
		return NetworkBoundary
	case matchesPrefix("/apis/workload"):
		return WholesaleWorkloadBoundary
	case matchesPrefix("/apis/marketplace"):
		return WorkloadBoundary
	default:
		return ""
	}
}

// buildRolePrefix constructs the role prefix string used for role validation.
func buildRolePrefix(boundary BoundaryType) string {
	return strings.ToUpper(string(boundary)) + "_BOUNDARY_"
}

