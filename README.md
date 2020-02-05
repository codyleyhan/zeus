# Zeus

Zeus is a simple middleware ordering package that uses a topological sort
to correctly order middleware that have defined their dependencies.  This
ensures that if you have multiple middleware that rely on each other that
they will always be added in the correct order without having to verify the
ordering yourself.

## Example Usage
```go
mw := func(a string) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			values = append(values, a)
			handler.ServeHTTP(w, r)
		})
	}
}

resolver, err := zeus.NewResolver(
	zeus.NewResolvableMiddleware("a", mw("a"), "b", "d"),
	zeus.NewResolvableMiddleware("b", mw("b"), "c", "e"),
	zeus.NewResolvableMiddleware("c", mw("c"), "d", "e"),
	zeus.NewResolvableMiddleware("e", mw("e")),
	zeus.NewResolvableMiddleware("d", mw("d"), "e"),
)

resolver.Setup() // returns a chained http middlware func in correct order
```