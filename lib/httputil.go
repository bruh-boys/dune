package lib

import (
	"fmt"
	"net/http/httputil"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(HTTPUTIL, `

declare namespace httputil {
	export function newSingleHostReverseProxy(target: http.URL): ReverseProxy
	
	export interface ReverseProxy {
		serveHTTP(w: http.ResponseWriter, r: http.Request): void
	}
}

`)
}

var HTTPUTIL = []dune.NativeFunction{
	{
		Name:      "httputil.newSingleHostReverseProxy",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			a := args[0]
			u, ok := a.ToObjectOrNil().(*URL)
			if !ok {
				return dune.NullValue, fmt.Errorf("invalid argument 1: expected http.URL, got %v", a.TypeName())
			}

			p := &reverseProxy{
				proxy: httputil.NewSingleHostReverseProxy(u.url),
			}

			return dune.NewObject(p), nil
		},
	},
}

type reverseProxy struct {
	proxy *httputil.ReverseProxy
}

func (*reverseProxy) Type() string {
	return "httputil.ReverseProxy"
}

func (r *reverseProxy) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "serveHTTP":
		return r.serveHTTP
	}
	return nil
}

func (r *reverseProxy) serveHTTP(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 2 {
		return dune.NullValue, fmt.Errorf("expected 2 argument, got %d", len(args))
	}

	w, ok := args[0].ToObjectOrNil().(*responseWriter)
	if !ok {
		return dune.NullValue, fmt.Errorf("invalid argument 1: expected http.ResponseWriter, got %v", args[0].TypeName())
	}

	req, ok := args[1].ToObjectOrNil().(*request)
	if !ok {
		return dune.NullValue, fmt.Errorf("invalid argument 2: expected http.Request, got %v", args[1].TypeName())
	}

	r.proxy.ServeHTTP(w.writer, req.request)

	return dune.NullValue, nil
}
