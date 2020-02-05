package middleware

type orderedMap struct {
	names map[Name]struct{}
	order []Name
}

func (o *orderedMap) Add(middleware ResolvableMiddleware) {
	if _, ok := o.names[middleware.Name()]; !ok {
		o.names[middleware.Name()] = struct{}{}
		o.order = append(o.order, middleware.Name())
	}
}

func (o *orderedMap) Has(name Name) bool {
	_, ok := o.names[name]
	return ok
}

func (o *orderedMap) Keys() []Name {
	return o.order
}

