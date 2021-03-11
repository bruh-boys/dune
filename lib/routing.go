package lib

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(Router, `
	
	declare namespace routing {
		export function newRouter(): Router
	
		export interface Router {
			reset(): void
			add(r: Route): void
			match(url: string): RouteMatch | null
			print(): void
		}
	
		export interface RouteMatch {
			route: Route
			data: string[]
			int(name: string): number
			string(name: string): string
		}
	
		interface Any {
			[prop: string]: any
		}

		export interface Route extends Any {
			url: string
			params?: string[]
			filter?: Function
			handler: Function
		}
	}

	`)
}

var Router = []dune.NativeFunction{
	{
		Name: "routing.newRouter",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			r := newRouter()
			return dune.NewObject(r), nil
		},
	},
}

type httpRouter struct {
	node *routeNode
}

func (r httpRouter) Type() string {
	return "routing.Router"
}

func (r httpRouter) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "reset":
		return r.reset
	case "add":
		return r.add
	case "match":
		return r.match
	case "print":
		return r.print
	}
	return nil
}

func (r httpRouter) print(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	r.node.print()
	return dune.NullValue, nil
}

func (r httpRouter) match(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	url := args[0].ToString()
	m, ok := r.Match(url)
	if ok {
		return dune.NewObject(m), nil
	}

	return dune.NullValue, nil
}

func (r httpRouter) reset(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	r.Reset()
	return dune.NullValue, nil
}

func (r httpRouter) add(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	var route *httpRoute

	v := args[0]

	switch v.Type {

	case dune.Object:
		o, ok := v.ToObject().(*httpRoute)
		if !ok {
			return dune.NullValue, fmt.Errorf("invalid type for a route: %v", v.TypeName())
		}
		route = o

	case dune.Map:
		values := make(map[string]dune.Value)
		var url string

		mo := v.ToMap()
		mo.RLock()
		defer mo.RUnlock()

		for k, mv := range mo.Map {
			s := k.ToString()
			if s == "url" {
				url = mv.ToString()
			} else {
				values[s] = mv
			}
		}

		route = &httpRoute{
			URL:   fixRouteURL(url),
			Value: values,
		}

	default:
		return dune.NullValue, fmt.Errorf("invalid type for route")
	}

	r.Add(route)

	return dune.NullValue, nil
}

func newRouter() *httpRouter {
	return &httpRouter{node: newNode()}
}

// remove all routes
func (r *httpRouter) Reset() {
	r.node = newNode()
}

func extensionAsSegment(url string) string {
	ext := filepath.Ext(url)
	if ext != "" {
		url = url[:len(url)-len(ext)] + "/" + ext[1:]
	}
	return url
}

func wildcardAsSegment(url string) string {
	ln := len(url)
	if ln > 2 {
		if url[ln-1] == '*' && url[ln-2] != '/' {
			url = url[:ln-1] + "/*"
		}
	}
	return url
}

func fixRouteURL(url string) string {
	url = extensionAsSegment(url)
	url = wildcardAsSegment(url)
	return url
}

func (r httpRouter) Add(t *httpRoute) {
	segments := strings.Split(t.URL, "/")

	node := r.node
	t.Params = nil

	for _, s := range segments {
		if s == "" {
			continue
		}

		if s[0] == ':' {
			t.Params = append(t.Params, s[1:])
			s = ":"
		}

		n, ok := node.child[s]
		if ok {
			node = n
			continue
		}

		n = newNode()
		node.child[s] = n
		node = n
	}

	node.route = t
}

func (r httpRouter) Match(url string) (routeMatch, bool) {
	url = extensionAsSegment(url)
	segments := Split(url, "/")

	var lastNotMatched bool
	var lastWildcardNode *routeNode

	var lastParamIndex int
	var lastParamNode *routeNode

	node := r.node

	for i, l := 0, len(segments); i < l; i++ {
		s := segments[i]

		var hasWildcard bool
		if n, ok := node.child["*"]; ok {
			lastWildcardNode = n
			hasWildcard = true
		}

		n, ok := node.child[s]
		if ok {
			// it is the last segment
			if l-i == 1 {
				node = n
				break
			}

			// only match if the next node has children
			if len(n.child) > 0 {
				// if there is a param also save it in case this
				// fixed path doesn't match in the next segment
				if n, ok := node.child[":"]; ok {
					lastParamNode = n
					lastParamIndex = i
				}
				node = n
				continue
			}
		}

		n, ok = node.child[":"]
		if ok {
			node = n
			lastParamNode = nil
			continue
		}

		if len(node.child) == 0 || hasWildcard {
			if l > i {
				lastNotMatched = true
			}
			break
		}

		// if this fixed path can't continue but the previous segment was a param
		// get back and take the param instead of the fixed segment
		if lastParamNode != nil {
			node = lastParamNode
			i = lastParamIndex
			lastParamNode = nil
			continue
		}

		lastNotMatched = true
		break
	}

	if node.route == nil {
		if n, ok := node.child["*"]; ok {
			return routeMatch{Route: n.route, Params: getParams(segments, n.route.URL)}, true
		}

		if lastWildcardNode != nil {
			return routeMatch{Route: lastWildcardNode.route, Params: getParams(segments, lastWildcardNode.route.URL)}, true
		}
		return routeMatch{}, false
	}

	if lastNotMatched {
		if lastWildcardNode != nil {
			return routeMatch{Route: lastWildcardNode.route, Params: getParams(segments, lastWildcardNode.route.URL)}, true
		}

		if node.route.URL == "/" {
			return routeMatch{Route: node.route}, true
		}

		return routeMatch{}, false
	}

	return routeMatch{Route: node.route, Params: getParams(segments, node.route.URL)}, true
}

func getParams(segments []string, route string) []string {
	var result []string

	routeParts := Split(route, "/")

	for i, part := range routeParts {
		if part[0] == ':' {
			result = append(result, segments[i])
		}
	}

	// fmt.Println(route, segments, result)
	// fmt.Println("---------------------------")
	return result
}

func newNode() *routeNode {
	return &routeNode{child: make(map[string]*routeNode)}
}

type routeNode struct {
	child map[string]*routeNode
	route *httpRoute
}

func (r *routeNode) print() {
	if r.route != nil {
		fmt.Println(r.route.URL)
	}

	for _, v := range r.child {
		v.print()
	}
}

// func (r *routeNode) printChildKeys() {
// 	for k := range r.child {
// 		fmt.Println(k)
// 	}
// }

type routeMatch struct {
	Route  *httpRoute
	Params []string
}

func (r routeMatch) Type() string {
	return "routing.RouteMatch"
}

func (r routeMatch) GetProperty(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "route":
		if r.Route == nil {
			return dune.NullValue, nil
		}
		return dune.NewObject(r.Route), nil
	case "data":
		p := make([]dune.Value, len(r.Params))
		for i, v := range r.Params {
			p[i] = dune.NewString(v)
		}
		return dune.NewArrayValues(p), nil
	}

	return dune.UndefinedValue, nil
}

func (r routeMatch) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "string":
		return r.string
	case "int":
		return r.int
	}
	return nil
}

func (r routeMatch) int(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}
	name := args[0].ToString()
	for i, k := range r.Route.Params {
		if k == name {
			s := r.Params[i]
			if s == "" {
				return dune.NullValue, nil
			}

			i, err := strconv.Atoi(s)
			if err != nil {
				return dune.NullValue, err
			}

			return dune.NewInt(i), nil
		}
	}
	return dune.NullValue, nil
}

func (r routeMatch) string(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}
	name := args[0].ToString()
	for i, k := range r.Route.Params {
		if k == name {
			return dune.NewString(r.Params[i]), nil
		}
	}
	return dune.NullValue, nil
}

func (m routeMatch) GetParam(name string) string {
	for i, k := range m.Route.Params {
		if k == name {
			return m.Params[i]
		}
	}
	return ""
}

type httpRoute struct {
	sync.RWMutex
	URL    string
	Params []string
	Value  map[string]dune.Value
}

func NewRoute(url string) *httpRoute {
	return &httpRoute{URL: fixRouteURL(url)}
}

func (r *httpRoute) Type() string {
	return "routing.Route"
}

func (r *httpRoute) GetProperty(name string, vm *dune.VM) (dune.Value, error) {
	switch name {

	case "url":
		return dune.NewString(r.URL), nil

	case "params":
		params := make([]dune.Value, len(r.Params))
		for i, p := range r.Params {
			params[i] = dune.NewString(p)
		}
		return dune.NewArrayValues(params), nil

	default:
		r.RLock()
		v, ok := r.Value[name]
		r.RUnlock()
		if !ok {
			return dune.UndefinedValue, nil
		}
		return v, nil
	}
}

func (r *httpRoute) SetProperty(name string, v dune.Value) error {
	switch name {

	case "url":
		r.URL = v.ToString()
		return nil

	case "params":
		return ErrReadOnly

	default:
		r.Lock()
		r.Value[name] = v
		r.Unlock()
		return nil
	}
}
