package middleware

import "net/http"

type Name string

type ResolvableMiddleware interface {
	Name() Name
	Dependencies() []Name
	Middleware() func(http.Handler) http.Handler
}

type resolvableMiddleware struct {
	mw func(http.Handler) http.Handler
	name Name
	dependencies []Name
}

var _ ResolvableMiddleware = &resolvableMiddleware{}

func NewResolvableMiddleware(name Name, mw func(http.Handler) http.Handler, deps ...Name) ResolvableMiddleware {
	return &resolvableMiddleware{
		mw: mw,
		name: name,
		dependencies: deps,
	}
}

func (h *resolvableMiddleware)  Middleware() func(http.Handler) http.Handler {
	return h.mw
}

func (h *resolvableMiddleware) Name() Name {
	return h.name
}

func (h *resolvableMiddleware) Dependencies() []Name {
	return h.dependencies
}