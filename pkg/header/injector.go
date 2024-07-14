package header

import (
	"encoding/base64"
	"fmt"
	"net/http"

	middlewareapi "github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/middleware"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/options"
	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/options/util"
)

type Injector interface {
	Inject(http.Header, *middlewareapi.RequestScope)
}

type injector struct {
	valueInjectors []valueInjector
}

func (i injector) Inject(header http.Header, scope *middlewareapi.RequestScope) {
	for _, injector := range i.valueInjectors {
		injector.inject(header, scope)
	}
}

func NewInjector(headers []options.Header) (Injector, error) {
	injectors := []valueInjector{}
	for _, header := range headers {
		for _, value := range header.Values {
			injector, err := newValueinjector(header.Name, value)
			if err != nil {
				return nil, fmt.Errorf("error building injector for header %q: %v", header.Name, err)
			}
			injectors = append(injectors, injector)
		}
	}

	return &injector{valueInjectors: injectors}, nil
}

type valueInjector interface {
	inject(http.Header, *middlewareapi.RequestScope)
}

func newValueinjector(name string, value options.HeaderValue) (valueInjector, error) {
	switch {
	case value.SecretSource != nil && value.ClaimSource == nil:
		return newSecretInjector(name, value.SecretSource)
	case value.SecretSource == nil && value.ClaimSource != nil:
		return newClaimInjector(name, value.ClaimSource)
	case value.ScopeSource != nil:
		return newScopeInjector(name, value.ScopeSource)
	default:
		return nil, fmt.Errorf("header %q value has multiple entries: only one entry per value is allowed", name)
	}
}

type injectorFunc struct {
	injectFunc func(http.Header, *middlewareapi.RequestScope)
}

func (i *injectorFunc) inject(header http.Header, scope *middlewareapi.RequestScope) {
	i.injectFunc(header, scope)
}

func newInjectorFunc(injectFunc func(header http.Header, scope *middlewareapi.RequestScope)) valueInjector {
	return &injectorFunc{injectFunc: injectFunc}
}

func newSecretInjector(name string, source *options.SecretSource) (valueInjector, error) {
	value, err := util.GetSecretValue(source)
	if err != nil {
		return nil, fmt.Errorf("error getting secret value: %v", err)
	}

	return newInjectorFunc(func(header http.Header, _ *middlewareapi.RequestScope) {
		header.Add(name, string(value))
	}), nil
}

func newClaimInjector(name string, source *options.ClaimSource) (valueInjector, error) {
	switch {
	case source.BasicAuthPassword != nil:
		password, err := util.GetSecretValue(source.BasicAuthPassword)
		if err != nil {
			return nil, fmt.Errorf("error loading basicAuthPassword: %v", err)
		}
		return newInjectorFunc(func(header http.Header, scope *middlewareapi.RequestScope) {
			session := scope.Session
			claimValues := session.GetClaim(source.Claim)
			for _, claim := range claimValues {
				if claim == "" {
					continue
				}
				auth := claim + ":" + string(password)
				header.Add(name, fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(auth))))
			}
		}), nil
	case source.Prefix != "":
		return newInjectorFunc(func(header http.Header, scope *middlewareapi.RequestScope) {
			session := scope.Session
			claimValues := session.GetClaim(source.Claim)
			for _, claim := range claimValues {
				if claim == "" {
					continue
				}
				header.Add(name, source.Prefix+claim)
			}
		}), nil
	default:
		return newInjectorFunc(func(header http.Header, scope *middlewareapi.RequestScope) {
			session := scope.Session
			claimValues := session.GetClaim(source.Claim)
			for _, claim := range claimValues {
				if claim == "" {
					continue
				}
				header.Add(name, claim)
			}
		}), nil
	}
}

func newScopeInjector(name string, source *options.ScopeSource) (valueInjector, error) {
	return newInjectorFunc(func(header http.Header, scope *middlewareapi.RequestScope) {
		header.Add(name, scope.GetRequestScopeField(source.Field))
	}), nil
}
