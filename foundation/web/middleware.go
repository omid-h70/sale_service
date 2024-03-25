package web

type MiddlewareFunc func(h HandlerFunc) HandlerFunc

func wrapMiddlewares(mw []MiddlewareFunc, h HandlerFunc) HandlerFunc {

	for i := 0; i < len(mw); i++ {

		if mw[i] != nil {
			mw[i](h)
		}
	}
	return h
}
