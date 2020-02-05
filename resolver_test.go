package middleware

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVerifyCorrectOrdering(t *testing.T) {
	mw := func(a string) func(http.Handler) http.Handler {
		return func(handler http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handler.ServeHTTP(w, r)
			})
		}
	}

	assert.False(t, VerifyCorrectOrdering(
		NewResolvableMiddleware("a", mw("a"), "b", "d"),
		NewResolvableMiddleware("b", mw("b"), "c", "e"),
		NewResolvableMiddleware("c", mw("c"), "d", "e"),
		NewResolvableMiddleware("e", mw("e")),
		NewResolvableMiddleware("d", mw("d"), "e"),
	), "should fail an invalid ordering")

	assert.True(t, VerifyCorrectOrdering(
		NewResolvableMiddleware("e", mw("e")),
		NewResolvableMiddleware("d", mw("d"), "e"),
		NewResolvableMiddleware("c", mw("c"), "d", "e"),
		NewResolvableMiddleware("b", mw("b"), "c", "e"),
		NewResolvableMiddleware("a", mw("a"), "b", "d"),
	), "should return true for a valid ordering")

	assert.False(t, VerifyCorrectOrdering(
		NewResolvableMiddleware("a", mw("a"), "b"),
		NewResolvableMiddleware("b", mw("b"), "c"),
		NewResolvableMiddleware("c", mw("c"), "a"),
	), "was not able to reject an ordering that has cycles")

	assert.False(t, VerifyCorrectOrdering(
		NewResolvableMiddleware("a", mw("a"), "b"),
		NewResolvableMiddleware("b", mw("b"), "c"),
	), "was not able to reject an ordering that has missing deps")
}

func TestNewResolver(t *testing.T) {
	values := []string{}
	mw := func(a string) func(http.Handler) http.Handler {
		return func(handler http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				values = append(values, a)
				handler.ServeHTTP(w, r)
			})
		}
	}

	resolver, err := NewResolver(
		NewResolvableMiddleware("a", mw("a"), "b", "d"),
		NewResolvableMiddleware("b", mw("b"), "c", "e"),
		NewResolvableMiddleware("c", mw("c"), "d", "e"),
		NewResolvableMiddleware("e", mw("e")),
		NewResolvableMiddleware("d", mw("d"), "e"),
		)
	require.NoError(t, err)
	assert.True(t, VerifyCorrectOrdering(resolver.OrderedMiddleware()...), "correctly orders middleware")

	h := resolver.Setup()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "http://localhost/test", nil))
	assert.Equal(t, []string{"e", "d", "c", "b", "a"}, values)

	resolver, err = NewResolver(
		NewResolvableMiddleware("a", mw("a"), "b"),
		NewResolvableMiddleware("b", mw("b"), "c"),
		NewResolvableMiddleware("c", mw("c"), "a"),
	)
	require.Error(t, err, "should detect and error on cycle")

	resolver, err = NewResolver(
		NewResolvableMiddleware("a", mw("a"), "b"),
		NewResolvableMiddleware("b", mw("b"), "c"),
	)
	require.Error(t, err, "should fail when not all deps passed")


}