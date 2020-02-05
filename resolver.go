package middleware

import (
	"errors"
	"fmt"
	"github.com/scylladb/go-set/strset"
	"net/http"
)

var (
	ErrCircularDependency = errors.New("unable to resolve, circular dependency")
	ErrMissingDependency = errors.New("missing dependency")
)

type Resolver struct {
	middlewareByName map[Name]ResolvableMiddleware
	orderedMiddleware []ResolvableMiddleware
}


// VerifyCorrectOrdering verifies that the middleware passed are in the correct order based off their dependencies
func VerifyCorrectOrdering(middlewares ...ResolvableMiddleware) bool {
	resolved := strset.New()
	for _, middleware := range middlewares {
		for _, deps := range middleware.Dependencies() {
			if !resolved.Has(string(deps)) {
				return false
			}
		}
		resolved.Add(string(middleware.Name()))
	}
	return true
}

// NewResolver creates a resolver that correctly orders the middleware passed.
// It checks that all dependencies are satisfied and checks for cycles both of which result in an error
func NewResolver(mw ResolvableMiddleware, middleware ...ResolvableMiddleware) (*Resolver, error){
	middleware = append(middleware, mw)
	resolver := Resolver{middlewareByName:make(map[Name]ResolvableMiddleware)}
	allMiddleware := strset.NewWithSize(len(middleware) + 1)
	for _, middleware := range middleware {
		resolver.middlewareByName[middleware.Name()] = middleware
		allMiddleware.Add(string(middleware.Name()))
	}

	resolved := &orderedMap{names: make(map[Name]struct{})}
	unresolved := strset.NewWithSize(len(middleware))
	if err := resolver.resolve(middleware[0], resolved, unresolved); err != nil {
		return nil, err
	}

	chain := []ResolvableMiddleware{}
	for _, name := range resolved.Keys() {
		chain = append(chain, resolver.middlewareByName[name])
		allMiddleware.Remove(string(name))
	}
	// there might be some middleware that were not depended on
	// so add those at the end
	for _, name := range allMiddleware.List() {
		chain = append(chain, resolver.middlewareByName[Name(name)])
	}

	resolver.orderedMiddleware = chain

	return &resolver, nil
}

// Setup creates the chained ordered middleware function from the middleware passed in the constructor
func (r *Resolver) Setup() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		for i := range r.orderedMiddleware {
			next = r.orderedMiddleware[len(r.orderedMiddleware)-1-i].Middleware()(next)
		}

		return next
	}
}

// OrderedMiddleware returns the resolved order for the middlewares based off their dependencies
func (r *Resolver) OrderedMiddleware() []ResolvableMiddleware {
	return r.orderedMiddleware
}


// resolve runs a topological sort to determine the dependencies of the passed middleware
func (r *Resolver) resolve(node ResolvableMiddleware, resolved *orderedMap, unresolved *strset.Set) error {
	unresolved.Add(string(node.Name()))
	for _, edge := range node.Dependencies() {
		if !resolved.Has(edge) {
			if unresolved.Has(string(edge)) {
				return fmt.Errorf("%q -> %q: %w", node.Name(), edge, ErrCircularDependency)
			}
		}
		next, ok := r.middlewareByName[edge]
		if !ok {
			return fmt.Errorf("unresolaveable middleware %q: %w", edge, ErrMissingDependency)
		}

		if err := r.resolve(next, resolved, unresolved); err != nil {
			return err
		}
	}
	resolved.Add(node)
	unresolved.Remove(string(node.Name()))
	return nil
}