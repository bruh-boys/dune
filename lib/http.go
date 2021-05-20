package lib

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dunelang/dune"
)

const genericError = "We are sorry, something went wrong"

func init() {
	// set a default timeout for the whole app
	http.DefaultClient.Timeout = time.Second * 60

	cacheBreaker = RandString(9)

	dune.RegisterLib(HTTP, `


declare namespace http {
	export const OK: number
	export const REDIRECT: number
	export const BAD_REQUEST: number
	export const UNAUTHORIZED: number
	export const NOT_FOUND: number
	export const INTERNAL_ERROR: number
	export const UNAVAILABLE: number

	export type SameSite = number
	export const SameSiteDefaultMode: SameSite
	export const SameSiteLaxMode: SameSite
	export const SameSiteStrictMode: SameSite
	export const SameSiteNoneMode: SameSite
	
    export function get(url: string, timeout?: time.Duration | number, config?: tls.Config): string
    export function post(url: string, data?: any): string

    export function getJSON(url: string): any

    export function cacheBreaker(): string
    export function resetCacheBreaker(): string

    export function encodeURIComponent(url: string): string
    export function decodeURIComponent(url: string): string

    export function parseURL(url?: string): URL

    export type Handler = (w: ResponseWriter, r: Request, routeData?: any) => void

    export interface Server {
        address: string
        addressTLS: string
        tlsConfig: tls.Config
        handler: Handler
		readHeaderTimeout: time.Duration | number
        writeTimeout: time.Duration | number
		readTimeout: time.Duration | number
        idleTimeout: time.Duration | number
        start(): void
        close(): void
        shutdown(duration?: time.Duration | number): void
    }

    export function newServer(): Server

    export type METHOD = "GET" | "POST" | "PUT" | "PATCH" | "DELETE" | "OPTIONS"

	export function newRequest(method: METHOD, url: string, data?: any): Request
	
	export function newResponseRecorder(r: Request): ResponseWriter

    export interface Request {
        /**
         * If the request is using a TLS connection
         */
        tls: boolean

        /**
         * The http method.
         */
        method: METHOD

        host: string

        url: URL

        referer: string

        userAgent: string

        body: io.ReaderCloser

        remoteAddr: string
        remoteIP: string

		/**
		 * The extension of the URL
		 */
        extension: string

        // value returns the first value for the named component of the query.
        // POST and PUT body parameters take precedence over URL query string values.
		value(key: string): string
		
		// json works as value but deserializes the value into an object.
        json(key: string): any

        // int works as value but converts the value to an int.
        int(key: string): number

        // float works as value but converts the value to a float.
        float(key: string): number

        // bool works as value but converts the value to an bool.
        bool(key: string): boolean

        // date works as value but converts the value to time.Time.
		date(key: string): time.Time
		
		routeInt(segment: number): number
		routeString(segment: number): string

        headers(): string[]
        header(key: string): string
        setHeader(key: string, value: string): void

        file(name: string): File

        values(): any

		formValues(): StringMap
		
        cookie(key: string): Cookie | null

        addCookie(c: Cookie): void

        setBasicAuth(user: string, password: string): void
        basicAuth(): { user: string, password: string }

        execute(timeout?: number | time.Duration, tlsconf?: tls.Config): Response
        executeString(timeout?: number | time.Duration, tlsconf?: tls.Config): string
        executeJSON(timeout?: number | time.Duration, tlsconf?: tls.Config): any 
    }


    export interface File {
        name: string
        contentType: string
        size: number
        read(b: byte[]): number
		ReadAt(p: byte[], off: number): number
        close(): void
    }

    export function newCookie(): Cookie

    export interface Cookie {
        domain: string
        path: string
        expires: time.Time
        name: string
        value: string
        secure: boolean
		httpOnly: boolean
		sameSite: SameSite
    }

    export interface URL {
        scheme: string
        host: string
        port: string

        /**
         * The host without the port number if present
         */
        hostName: string

        /**
         * returns the subdomain part *only if* the host has a format xxx.xxxx.xx.
         */
        subdomain: string

        path: string
        query: string
		pathAndQuery: string
		
		string(): string
    }

    // interface FormValues {
    //     [key: string]: any
    // }  


    export interface Response {
        status: number
        handled: boolean
        proto: string
		body(): string
		json(): any
		bytes(): byte[]
		cookies(): Cookie[]
		headers(): string[]
		header(name: string): string[]
    }


    export interface ResponseWriter {
        readonly status: number

        handled: boolean

		body(): string
		json(): any
		bytes(): byte[]
		
        cookie(name: string): Cookie

        cookies(): Cookie[]

		addCookie(c: Cookie): void
		
		headers(): string[]
		header(name: string): string[]

        /**
         * Writes v to the server response.
         */
        write(v: any): number

		writeGziped(v: any): number

        /**
         * Writes v to the server response setting json content type if
         * the header is not already set.
         */
        writeJSON(v: any, skipCacheHeader?: boolean): void

        /**
         * Writes v to the server response setting json content type if
         * the header is not already set.
         */
        writeJSONStatus(status: number, v: any, skipCacheHeader?: boolean): void

        /**
         * Serves a static file
         */
        writeFile(name: string, data: byte[] | string | io.File | io.FileSystem): void

        /**
         * Sets the http status header.
         */
        setStatus(status: number): void

        /**
         * Sets the content type header.
         */
        setContentType(type: string): void

        /**
         * Sets the content type header.
         */
        setHeader(name: string, value: string): void

        /**
         * Send a error to the client
         */
        writeError(status: number, msg?: string): void

        /**
         * Send a error with json content-type to the client
         */
        writeJSONError(status: number, msg?: string): void

        redirect(url: string, status?: number): void
    }


}

`)
}

const MAX_PARSE_FORM_MEMORY = 10000

var cacheBreaker string

var HTTP = []dune.NativeFunction{
	{
		Name: "->http.OK",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(200), nil
		},
	},
	{
		Name: "->http.REDIRECT",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(302), nil
		},
	},
	{
		Name: "->http.BAD_REQUEST",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(400), nil
		},
	},
	{
		Name: "->http.UNAUTHORIZED",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(401), nil
		},
	},
	{
		Name: "->http.NOT_FOUND",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(404), nil
		},
	},
	{
		Name: "->http.INTERNAL_ERROR",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(500), nil
		},
	},
	{
		Name: "->http.UNAVAILABLE",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(503), nil
		},
	},
	{
		Name: "->http.SameSiteDefaultMode",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(http.SameSiteDefaultMode)), nil
		},
	},
	{
		Name: "->http.SameSiteDefaultMode",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(http.SameSiteDefaultMode)), nil
		},
	},

	{
		Name: "->http.SameSiteLaxMode",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(http.SameSiteLaxMode)), nil
		},
	},
	{
		Name: "->http.SameSiteStrictMode",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(http.SameSiteStrictMode)), nil
		},
	},
	{
		Name: "->http.SameSiteNoneMode",
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewInt(int(http.SameSiteNoneMode)), nil
		},
	},
	{
		Name:      "http.cacheBreaker",
		Arguments: 0,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			return dune.NewString(cacheBreaker), nil
		},
	},
	{
		Name:        "http.resetCacheBreaker",
		Arguments:   0,
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			cacheBreaker = RandString(9)
			return dune.NullValue, nil
		},
	},
	{
		Name:        "http.newServer",
		Arguments:   0,
		Permissions: []string{"netListen"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			s := &server{
				vm:                vm,
				handler:           -1,
				readHeaderTimeout: 5 * time.Second,
				readTimeout:       5 * time.Second,
				writeTimeout:      5 * time.Second,
				idleTimeout:       5 * time.Second,
			}

			return dune.NewObject(s), nil
		},
	},
	{
		Name:      "http.newResponseRecorder",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Object); err != nil {
				return dune.NullValue, err
			}

			r, ok := args[0].ToObjectOrNil().(*request)
			if !ok {
				return dune.NullValue, fmt.Errorf("expected a request, got %v", args[0].TypeName())
			}

			w := &responseWriter{
				writer:  httptest.NewRecorder(),
				request: r.request,
			}

			return dune.NewObject(w), nil
		},
	},
	{
		Name:      "http.newCookie",
		Arguments: 0,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			c := &cookie{}
			return dune.NewObject(c), nil
		},
	},
	{
		Name:      "http.encodeURIComponent",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String); err != nil {
				return dune.NullValue, err
			}
			v := args[0].String()
			u := url.QueryEscape(v)
			return dune.NewString(u), nil
		},
	},
	{
		Name:      "http.decodeURIComponent",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String); err != nil {
				return dune.NullValue, err
			}
			v := args[0].String()
			u, err := url.QueryUnescape(v)
			if err != nil {
				return dune.NullValue, err
			}
			return dune.NewString(u), nil
		},
	},
	{
		Name:        "http.newRequest",
		Arguments:   -1,
		Permissions: []string{"networking"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgRange(args, 2, 3); err != nil {
				return dune.NullValue, err
			}

			if args[0].Type != dune.String {
				return dune.NullValue, fmt.Errorf("expected argument 1 to be string, got %v", args[0].Type)
			}

			if args[1].Type != dune.String {
				return dune.NullValue, fmt.Errorf("expected argument 2 to be string, got %v", args[1].Type)
			}

			var method string
			var urlStr string
			var queryMap map[dune.Value]dune.Value
			var reader io.Reader
			var contentType string

			switch len(args) {
			case 2:
				method = args[0].String()
				urlStr = args[1].String()
			case 3:
				method = args[0].String()
				urlStr = args[1].String()
				form := url.Values{}

				v := args[2]

				switch v.Type {
				case dune.Null, dune.Undefined:
				case dune.String:
					switch method {
					case "POST", "PUT", "PATCH":
					default:
						return dune.NullValue, fmt.Errorf("can only pass a data string with POST, PUT or PATCH")
					}
					reader = strings.NewReader(v.String())
					contentType = "application/json; charset=UTF-8"
				case dune.Map:
					m := v.ToMap()
					if method == "GET" {
						queryMap = m.Map
					} else {
						m.RLock()
						for k, v := range m.Map {
							if v.IsNilOrEmpty() {
								continue
							}
							vs, err := serialize(v)
							if err != nil {
								return dune.NullValue, fmt.Errorf("error serializing parameter: %v", v.Type)
							}
							form.Add(k.String(), vs)
						}
						m.RUnlock()
						reader = strings.NewReader(form.Encode())
						contentType = "application/x-www-form-urlencoded"
					}
				default:
					return dune.NullValue, fmt.Errorf("expected argument 3 to be object, got %v", v.Type)
				}
			}

			r, err := http.NewRequest(method, urlStr, reader)
			if err != nil {
				return dune.NullValue, err
			}
			switch method {
			case "POST", "PUT", "PATCH":
				r.Header.Add("Content-Type", contentType)
			case "GET":
				if queryMap != nil {
					q := r.URL.Query()
					for k, v := range queryMap {
						if v.IsNilOrEmpty() {
							continue
						}
						vs, err := serialize(v)
						if err != nil {
							return dune.NullValue, fmt.Errorf("error serializing parameter: %v", v.Type)
						}
						q.Add(k.String(), vs)
					}
					r.URL.RawQuery = q.Encode()
				}
			}

			return dune.NewObject(&request{request: r}), nil
		},
	},
	{
		Name:        "http.get",
		Arguments:   -1,
		Permissions: []string{"networking"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			client := &http.Client{}
			timeout := 20 * time.Second

			ln := len(args)

			if ln == 0 {
				return dune.NullValue, fmt.Errorf("expected 1 to 3 arguments, got %d", len(args))
			}

			a := args[0]
			if a.Type != dune.String {
				return dune.NullValue, fmt.Errorf("expected argument 0 to be string, got %s", a.TypeName())
			}
			url := a.String()

			if ln == 0 {
			} else if ln > 1 {
				a := args[1]
				switch a.Type {
				case dune.Undefined, dune.Null:
				case dune.Int, dune.Object:
					var err error
					timeout, err = ToDuration(args[1])
					if err != nil {
						return dune.NullValue, err
					}
				default:
					return dune.NullValue, fmt.Errorf("expected argument 1 to be duration")
				}
			}

			if ln > 2 {
				b := args[2]
				switch b.Type {
				case dune.Null, dune.Undefined:
				case dune.Object:
					t, ok := args[2].ToObjectOrNil().(*tlsConfig)
					if !ok {
						return dune.NullValue, fmt.Errorf("expected argument 2 to be tls.Config")
					}
					client.Transport = getTransport(t.conf)
				default:
					return dune.NullValue, fmt.Errorf("expected argument 2 to be string, got %s", b.TypeName())
				}
			}

			client.Timeout = timeout
			resp, err := client.Get(url)
			if err != nil {
				return dune.NullValue, err
			}

			b, err := ioutil.ReadAll(resp.Body)

			resp.Body.Close()

			if err != nil {
				return dune.NullValue, err
			}

			if !isHTTPSuccess(resp.StatusCode) {
				return dune.NullValue, fmt.Errorf("http Error %d: %v", resp.StatusCode, string(b))
			}

			return dune.NewString(string(b)), nil
		},
	},
	{
		Name:        "http.post",
		Arguments:   2,
		Permissions: []string{"networking"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String, dune.Map); err != nil {
				return dune.NullValue, err
			}
			u := args[0].String()

			data := url.Values{}

			m := args[1].ToMap()
			m.RLock()
			for k, v := range m.Map {
				data.Add(k.String(), v.String())
			}
			m.RUnlock()

			resp, err := http.PostForm(u, data)
			if err != nil {
				return dune.NullValue, err
			}

			b, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				return dune.NullValue, err
			}

			return dune.NewString(string(b)), nil
		},
	},
	{
		Name:        "http.getJSON",
		Arguments:   1,
		Permissions: []string{"networking"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String); err != nil {
				return dune.NullValue, err
			}
			url := args[0].String()

			resp, err := http.Get(url)
			if err != nil {
				return dune.NullValue, err
			}

			b, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				return dune.NullValue, err
			}

			v, err := unmarshal(b)
			if err != nil {
				return dune.NullValue, err
			}

			return v, nil
		},
	},
	{
		Name:      "http.parseURL",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateOptionalArgs(args, dune.String); err != nil {
				return dune.NullValue, err
			}

			if len(args) == 0 {
				u := &url.URL{}
				return dune.NewObject(&URL{url: u}), nil
			}

			rawURL := args[0].String()
			u, err := url.Parse(rawURL)
			if err != nil {
				return dune.NullValue, err
			}
			return dune.NewObject(&URL{u}), nil
		},
	},
}

func getTransport(tlsConf *tls.Config) *http.Transport {
	return &http.Transport{
		TLSClientConfig: tlsConf,
		Proxy:           http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

func isHTTPSuccess(httpCode int) bool {
	return httpCode >= 200 && httpCode < 300
}

func serialize(v dune.Value) (string, error) {
	switch v.Type {
	case dune.Int, dune.Float, dune.String, dune.Bool, dune.Rune:
		return v.String(), nil
	}

	b, err := json.Marshal(v.ExportMarshal(0))
	if err != nil {
		return "", err
	}

	// return strings without quotes, only the content of the string
	ln := len(b)
	if ln >= 2 && b[0] == '"' && b[ln-1] == '"' {
		b = b[1 : ln-1]
	}

	return string(b), nil
}

type server struct {
	sync.Mutex
	address           string
	addressTLS        string
	handler           int
	closureHandler    *dune.Closure
	methodHandler     *dune.Method
	tlsConfig         *tlsConfig
	server            *http.Server
	tlsServer         *http.Server
	readHeaderTimeout time.Duration
	writeTimeout      time.Duration
	readTimeout       time.Duration
	idleTimeout       time.Duration
	vm                *dune.VM
}

func (s *server) Type() string {
	return "http.Server"
}

func (s *server) GetField(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "address":
		return dune.NewString(s.address), nil
	case "addressTLS":
		return dune.NewString(s.addressTLS), nil
	case "tlsConfig":
		return dune.NewObject(s.tlsConfig), nil
	case "handler":
		if s.closureHandler != nil {
			return dune.NewObject(s.closureHandler), nil
		}
		if s.handler != -1 {
			return dune.NewFunction(s.handler), nil
		}
		return dune.NullValue, nil
	case "readHeaderTimeout":
		return dune.NewObject(Duration(s.readHeaderTimeout)), nil
	case "writeTimeout":
		return dune.NewObject(Duration(s.writeTimeout)), nil
	case "readTimeout":
		return dune.NewObject(Duration(s.readTimeout)), nil
	case "idleTimeout":
		return dune.NewObject(Duration(s.idleTimeout)), nil
	}
	return dune.UndefinedValue, nil
}

func (s *server) SetField(name string, v dune.Value, vm *dune.VM) error {
	switch name {
	case "address":
		if v.Type != dune.String {
			return fmt.Errorf("invalid type, expected string")
		}
		s.address = v.String()
		return nil

	case "addressTLS":
		if v.Type != dune.String {
			return fmt.Errorf("invalid type, expected string")
		}
		s.addressTLS = v.String()
		return nil

	case "handler":
		switch v.Type {
		case dune.Func:
			s.handler = v.ToFunction()
			return nil

		case dune.Object:
			switch t := v.ToObject().(type) {
			case *dune.Closure:
				s.closureHandler = t
				return nil
			case *dune.Method:
				s.methodHandler = t
				return nil
			default:
				return fmt.Errorf("invalid handler type %v", v.TypeName())
			}

		default:
			return fmt.Errorf("invalid type, %v is not a function", v.TypeName())
		}

	case "tlsConfig":
		if v.Type != dune.Object {
			return fmt.Errorf("invalid type, expected a tls object")
		}
		tls, ok := v.ToObjectOrNil().(*tlsConfig)
		if !ok {
			return fmt.Errorf("invalid type, expected a tls object")
		}
		s.tlsConfig = tls
		return nil

	case "readHeaderTimeout":
		d, err := ToDuration(v)
		if err != nil {
			return err
		}
		s.readHeaderTimeout = d
		return nil

	case "writeTimeout":
		d, err := ToDuration(v)
		if err != nil {
			return err
		}
		s.writeTimeout = d
		return nil

	case "readTimeout":
		d, err := ToDuration(v)
		if err != nil {
			return err
		}
		s.readTimeout = d
		return nil

	case "idleTimeout":
		d, err := ToDuration(v)
		if err != nil {
			return err
		}
		s.idleTimeout = d
		return nil
	}

	return ErrReadOnlyOrUndefined
}

func (s *server) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "start":
		return s.start
	case "close":
		return s.close
	case "shutdown":
		return s.shutdown
	}
	return nil
}

func (s *server) shutdown(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	l := len(args)

	var d time.Duration

	switch l {
	case 0:
		d = time.Second
	case 1:
		var a = args[0]
		switch a.Type {
		case dune.Int:
			dd, err := ToDuration(a)
			if err != nil {
				return dune.NullValue, err
			}
			d = dd
		case dune.Object:
			dur, ok := a.ToObject().(Duration)
			if !ok {
				return dune.NullValue, fmt.Errorf("expected duration, got %s", a.TypeName())
			}
			d = time.Duration(dur)
		}
	default:
		return dune.NullValue, fmt.Errorf("expected 0 or 1 argument, got %d", l)
	}

	ctx, cancel := context.WithTimeout(context.Background(), d)

	s.Lock()
	err := s.server.Shutdown(ctx)
	s.Unlock()

	var err2 error
	if s.tlsServer != nil {
		err2 = s.tlsServer.Shutdown(ctx)
	}

	if err == nil {
		err = err2
	}

	cancel()
	return dune.NullValue, err
}
func (s *server) close(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	s.Lock()
	err := s.server.Close()
	s.Unlock()

	var err2 error
	if s.tlsServer != nil {
		err2 = s.tlsServer.Close()
	}

	if err == nil {
		err = err2
	}

	return dune.NullValue, err
}

func (s *server) start(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if s.tlsConfig != nil {
		s.tlsServer = &http.Server{
			ReadHeaderTimeout: s.readHeaderTimeout,
			ReadTimeout:       s.readTimeout,
			WriteTimeout:      s.writeTimeout,
			IdleTimeout:       s.idleTimeout,
			TLSConfig:         s.tlsConfig.conf,
			Addr:              s.addressTLS,
			Handler:           s,
		}

		if s.tlsConfig.certManager != nil {
			go func() {
				// serve HTTP, which will redirect automatically to HTTPS
				h := s.tlsConfig.certManager.manager.HTTPHandler(nil)
				if err := http.ListenAndServe(":http", h); err != nil {
					vm.Error = err
					fmt.Fprintln(vm.GetStderr(), err)
				}
			}()
		} else if s.address != "" {
			go func() {
				s.server = &http.Server{
					ReadHeaderTimeout: s.readHeaderTimeout,
					ReadTimeout:       s.readTimeout,
					WriteTimeout:      s.writeTimeout,
					IdleTimeout:       s.idleTimeout,
					Addr:              s.address,
					Handler:           s,
				}

				if err := s.server.ListenAndServe(); err != nil {
					vm.Error = err
					fmt.Fprintln(vm.GetStderr(), err)
				}
			}()
		}

		if err := s.tlsServer.ListenAndServeTLS("", ""); err != nil {
			return dune.NullValue, fmt.Errorf("ListenAndServeTLS: %w", err)
		}

		return dune.NullValue, nil
	}

	s.server = &http.Server{
		ReadHeaderTimeout: s.readHeaderTimeout,
		ReadTimeout:       s.readTimeout,
		WriteTimeout:      s.writeTimeout,
		IdleTimeout:       s.idleTimeout,
		Addr:              s.address,
		Handler:           s,
	}

	return dune.NullValue, s.server.ListenAndServe()
}
func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vm := s.vm.CloneInitialized(s.vm.Program, s.vm.Globals())

	rr := &responseWriter{
		writer:  w,
		request: r,
	}

	req := &request{
		request: r,
		writer:  w,
	}

	if s.closureHandler != nil {
		if _, err := vm.RunClosure(s.closureHandler, dune.NewObject(rr), dune.NewObject(req)); err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
		}
	} else if s.methodHandler != nil {
		if _, err := vm.RunMethod(s.methodHandler, dune.NewObject(rr), dune.NewObject(req)); err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
		}
	} else if s.handler >= 0 {
		if _, err := vm.RunFuncIndex(s.handler, dune.NewObject(rr), dune.NewObject(req)); err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
		}
	} else {
		fmt.Println("no handler")
	}
}

type cookie struct {
	domain   string
	path     string
	expires  time.Time
	name     string
	value    string
	secure   bool
	httpOnly bool
	sameSite http.SameSite
}

func (c *cookie) Type() string {
	return "http.Cookie"
}

func (c *cookie) String() string {
	return fmt.Sprintf("{ name: '%s', path: '%s', value: '%s', expires: '%v' }", c.name, c.path, c.value, c.expires)
}

func (c *cookie) GetField(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "domain":
		return dune.NewString(c.name), nil
	case "path":
		return dune.NewString(c.path), nil
	case "expires":
		return dune.NewObject(TimeObj(c.expires)), nil
	case "name":
		return dune.NewString(c.name), nil
	case "value":
		return dune.NewString(c.value), nil
	case "secure":
		return dune.NewBool(c.secure), nil
	case "httpOnly":
		return dune.NewBool(c.httpOnly), nil
	case "sameSite":
		return dune.NewInt(int(c.sameSite)), nil
	}
	return dune.UndefinedValue, nil
}

func (c *cookie) SetField(name string, v dune.Value, vm *dune.VM) error {
	switch name {
	case "domain":
		if v.Type != dune.String {
			return ErrInvalidType
		}
		c.domain = v.String()
		return nil
	case "path":
		if v.Type != dune.String {
			return ErrInvalidType
		}
		c.path = v.String()
		return nil
	case "expires":
		if v.Type != dune.Object {
			return ErrInvalidType
		}
		t, ok := v.ToObject().(TimeObj)
		if !ok {
			return ErrInvalidType
		}
		c.expires = time.Time(t)
		return nil
	case "name":
		if v.Type != dune.String {
			return ErrInvalidType
		}
		c.name = v.String()
		return nil
	case "value":
		if v.Type != dune.String {
			return ErrInvalidType
		}
		c.value = v.String()
		return nil
	case "secure":
		if v.Type != dune.Bool {
			return ErrInvalidType
		}
		c.secure = v.ToBool()
		return nil
	case "httpOnly":
		if v.Type != dune.Bool {
			return ErrInvalidType
		}
		c.httpOnly = v.ToBool()
		return nil
	case "sameSite":
		if v.Type != dune.Int {
			return ErrInvalidType
		}

		vv := http.SameSite(v.ToInt())
		switch vv {
		case http.SameSiteDefaultMode:
			c.sameSite = http.SameSiteDefaultMode
		case http.SameSiteLaxMode:
			c.sameSite = http.SameSiteLaxMode
		case http.SameSiteStrictMode:
			c.sameSite = http.SameSiteStrictMode
		case http.SameSiteNoneMode:
			c.sameSite = http.SameSiteNoneMode
		default:
			return fmt.Errorf("invalid SameSite value: %v", vv)
		}
		return nil
	}
	return ErrReadOnlyOrUndefined
}

type response struct {
	r       *http.Response
	handled bool
}

func (r *response) Type() string {
	return "http.Response"
}

func (r *response) GetField(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "handled":
		return dune.NewBool(r.handled), nil
	case "status":
		return dune.NewInt(r.r.StatusCode), nil
	case "proto":
		return dune.NewString(r.r.Proto), nil
	}
	return dune.UndefinedValue, nil
}

func (r *response) SetField(name string, v dune.Value, vm *dune.VM) error {
	switch name {
	case "handled":
		if v.Type != dune.Bool {
			return fmt.Errorf("invalid type. Expected boolean")
		}
		r.handled = v.ToBool()
		return nil
	}
	return ErrReadOnlyOrUndefined
}

func (r *response) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "headers":
		return r.headers
	case "header":
		return r.header
	case "body":
		return r.body
	case "json":
		return r.json
	case "bytes":
		return r.bytes
	case "cookies":
		return r.cookies
	}
	return nil
}

func (r *response) cookies(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}

	cookies := r.r.Cookies()

	v := make([]dune.Value, len(cookies))

	for i, k := range cookies {
		c := &cookie{
			domain:   k.Domain,
			path:     k.Path,
			expires:  k.Expires,
			secure:   k.Secure,
			httpOnly: k.HttpOnly,
			name:     k.Name,
			value:    k.Value,
			sameSite: k.SameSite,
		}
		v[i] = dune.NewObject(c)
	}

	return dune.NewArrayValues(v), nil
}

func (r *response) headers(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgRange(args, 0, 0); err != nil {
		return dune.NullValue, err
	}

	resp := r.r

	values := make([]dune.Value, len(resp.Header))

	i := 0
	for k := range resp.Header {
		values[i] = dune.NewString(k)
		i++
	}

	return dune.NewArrayValues(values), nil
}

func (r *response) header(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	name := args[0].String()

	headers := r.r.Header[name]

	values := make([]dune.Value, len(headers))

	for i, v := range headers {
		values[i] = dune.NewString(v)
	}

	return dune.NewArrayValues(values), nil
}

func (r *response) body(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgRange(args, 0, 0); err != nil {
		return dune.NullValue, err
	}

	resp := r.r

	b, err := ioutil.ReadAll(resp.Body)

	resp.Body.Close()

	if err != nil {
		return dune.NullValue, err
	}

	return dune.NewString(string(b)), nil
}

func (r *response) json(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgRange(args, 0, 0); err != nil {
		return dune.NullValue, err
	}

	resp := r.r

	b, err := ioutil.ReadAll(resp.Body)

	resp.Body.Close()

	if err != nil {
		return dune.NullValue, err
	}

	return unmarshal(b)
}

func (r *response) bytes(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgRange(args, 0, 0); err != nil {
		return dune.NullValue, err
	}

	resp := r.r

	b, err := ioutil.ReadAll(resp.Body)

	resp.Body.Close()

	if err != nil {
		return dune.NullValue, err
	}

	return dune.NewBytes(b), nil
}

type request struct {
	request *http.Request
	writer  http.ResponseWriter

	// to allow get them multiple times
	requestValues dune.Value

	// for testing
	tls bool
}

func (r *request) Type() string {
	return "http.Request"
}

func (r *request) GetField(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "body":
		return dune.NewObject(&readerCloser{r.request.Body}), nil
	case "tls":
		if r.tls {
			return dune.TrueValue, nil
		}
		return dune.NewBool(r.request.TLS != nil), nil
	case "host":
		return dune.NewString(r.request.Host), nil
	case "method":
		return dune.NewString(r.request.Method), nil
	case "userAgent":
		return dune.NewString(r.request.UserAgent()), nil
	case "referer":
		return dune.NewString(r.request.Referer()), nil
	case "remoteAddr":
		return dune.NewString(r.request.RemoteAddr), nil
	case "remoteIP":
		a := r.request.RemoteAddr
		if !strings.ContainsRune(a, ':') {
			return dune.NewString(a), nil
		}
		ip, _, err := net.SplitHostPort(r.request.RemoteAddr)
		if err != nil {
			return dune.NullValue, err
		}
		return dune.NewString(ip), nil
	case "url":
		u := r.request.URL
		u.Host = r.request.Host
		url := &URL{url: u}
		return dune.NewObject(url), nil
	case "extension":
		return dune.NewString(filepath.Ext(r.request.URL.Path)), nil
	}

	return dune.UndefinedValue, nil
}

func (r *request) SetField(name string, v dune.Value, vm *dune.VM) error {
	switch name {
	case "method":
		if v.Type != dune.String {
			return fmt.Errorf("invalid type. Expected string")
		}
		r.request.Method = v.String()
		return nil
	case "host":
		if v.Type != dune.String {
			return fmt.Errorf("invalid type. Expected string")
		}
		r.request.Host = v.String()
		return nil
	case "tls":
		if v.Type != dune.Bool {
			return fmt.Errorf("invalid type. Expected bool")
		}
		r.tls = v.ToBool()
		return nil
	}
	return ErrReadOnlyOrUndefined
}

func (r *request) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "headers":
		return r.headers
	case "header":
		return r.header
	case "setBasicAuth":
		return r.setBasicAuth
	case "basicAuth":
		return r.basicAuth
	case "setHeader":
		return r.setHeader
	case "execute":
		return r.execute
	case "executeString":
		return r.executeString
	case "executeJSON":
		return r.executeJSON
	case "value":
		return r.value
	case "int":
		return r.formInt
	case "float":
		return r.formFloat
	case "bool":
		return r.formBool
	case "date":
		return r.formDate
	case "json":
		return r.formJSON
	case "routeString":
		return r.routeString
	case "routeInt":
		return r.routeInt
	case "file":
		return r.file
	case "cookie":
		return r.cookie
	case "addCookie":
		return r.addCookie
	case "values":
		return r.values
	case "formValues":
		return r.formValues
	}
	return nil
}

func (r *request) values(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if r.requestValues != dune.NullValue {
		return r.requestValues, nil
	}

	var form url.Values
	req := r.request

	contentType := req.Header.Get("Content-Type")

	if strings.Contains(contentType, "application/json") {
		buf, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return dune.NullValue, err
		}
		if len(buf) == 0 {
			return dune.NullValue, nil
		}
		v, err := unmarshal(buf)
		r.requestValues = v
		return v, err
	}

	if req.Method == "POST" {
		if err := req.ParseMultipartForm(MAX_PARSE_FORM_MEMORY); err != nil {
			if err := req.ParseForm(); err != nil {
				return dune.NullValue, err
			}
		}
		form = req.Form
	} else {
		form = req.URL.Query()
	}

	values := make(map[dune.Value]dune.Value, len(form))
	for k, v := range form {
		values[dune.NewString(k)] = dune.NewString(v[0])
	}

	v := dune.NewMapValues(values)
	r.requestValues = v
	return v, nil
}

func (r *request) formValues(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	req := r.request
	if req.Method != "POST" {
		return dune.NullValue, nil
	}
	if err := req.ParseMultipartForm(MAX_PARSE_FORM_MEMORY); err != nil {
		if err := req.ParseForm(); err != nil {
			return dune.NullValue, err
		}
	}
	form := req.Form
	values := make(map[dune.Value]dune.Value, len(form))
	for k, v := range form {
		values[dune.NewString(k)] = dune.NewString(v[0])
	}
	return dune.NewMapValues(values), nil
}

func (r *request) routeInt(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Int); err != nil {
		return dune.NullValue, err
	}

	index := int(args[0].ToInt())

	url := r.request.URL.Path
	ext := filepath.Ext(url)
	if ext != "" {
		url = strings.TrimSuffix(url, ext)
	}
	segments := Split(url, "/")

	if len(segments) > index {
		i, err := strconv.Atoi(segments[index])
		if err == nil {
			return dune.NewInt(i), nil
		}
	}

	return dune.NullValue, nil
}

func (r *request) routeString(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Int); err != nil {
		return dune.NullValue, err
	}

	index := int(args[0].ToInt())

	url := r.request.URL.Path
	ext := filepath.Ext(url)
	if ext != "" {
		url = strings.TrimSuffix(url, ext)
	}
	segments := Split(url, "/")

	if len(segments) > index {
		return dune.NewString(segments[index]), nil
	}

	return dune.NullValue, nil
}

func (r *request) basicAuth(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}

	user, pwd, ok := r.request.BasicAuth()
	if !ok {
		return dune.NullValue, nil
	}

	m := make(map[dune.Value]dune.Value)
	m[dune.NewString("user")] = dune.NewString(user)
	m[dune.NewString("password")] = dune.NewString(pwd)

	return dune.NewMapValues(m), nil
}

func (r *request) setBasicAuth(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String, dune.String); err != nil {
		return dune.NullValue, err
	}

	user := args[0].String()
	pwd := args[1].String()
	r.request.SetBasicAuth(user, pwd)

	return dune.NullValue, nil
}

func (r *request) execute(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	client := &http.Client{}

	ln := len(args)

	if ln > 0 {
		a := args[0]
		switch a.Type {
		case dune.Undefined, dune.Null:
		case dune.Int, dune.Object:
			d, err := ToDuration(a)
			if err != nil {
				return dune.NullValue, err
			}

			client.Timeout = d
		default:
			return dune.NullValue, fmt.Errorf("expected argument 1 to be duration")
		}
	}

	if ln > 1 {
		tlsc, ok := args[1].ToObjectOrNil().(*tlsConfig)
		if !ok {
			return dune.NullValue, fmt.Errorf("expected arg 2 to be TLSConfig")
		}
		client.Transport = getTransport(tlsc.conf)
	}

	resp, err := client.Do(r.request)
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NewObject(&response{r: resp}), nil
}

func (r *request) executeString(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	client := &http.Client{}

	ln := len(args)

	if ln > 0 {
		a := args[0]
		switch a.Type {
		case dune.Undefined, dune.Null:
		case dune.Int, dune.Object:
			d, err := ToDuration(a)
			if err != nil {
				return dune.NullValue, err
			}

			client.Timeout = d
		default:
			return dune.NullValue, fmt.Errorf("expected argument 1 to be duration")
		}
	}

	if ln > 1 {
		tlsc, ok := args[1].ToObjectOrNil().(*tlsConfig)
		if !ok {
			return dune.NullValue, fmt.Errorf("expected arg 2 to be TLSConfig")
		}
		client.Transport = getTransport(tlsc.conf)
	}

	resp, err := client.Do(r.request)
	if err != nil {
		return dune.NullValue, err
	}

	b, err := ioutil.ReadAll(resp.Body)

	resp.Body.Close()

	if err != nil {
		return dune.NullValue, err
	}

	if !isHTTPSuccess(resp.StatusCode) {
		return dune.NullValue, fmt.Errorf("http Error %d: %v", resp.StatusCode, string(b))
	}

	return dune.NewString(string(b)), nil
}

func (r *request) executeJSON(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	client := &http.Client{}

	ln := len(args)

	if ln > 0 {
		a := args[0]
		switch a.Type {
		case dune.Undefined, dune.Null:
		case dune.Int, dune.Object:
			d, err := ToDuration(a)
			if err != nil {
				return dune.NullValue, err
			}

			client.Timeout = d
		default:
			return dune.NullValue, fmt.Errorf("expected argument 1 to be duration")
		}
	}

	if ln > 1 {
		tlsc, ok := args[1].ToObjectOrNil().(*tlsConfig)
		if !ok {
			return dune.NullValue, fmt.Errorf("expected arg 2 to be TLSConfig")
		}
		client.Transport = getTransport(tlsc.conf)
	}

	resp, err := client.Do(r.request)
	if err != nil {
		return dune.NullValue, err
	}

	b, err := ioutil.ReadAll(resp.Body)

	resp.Body.Close()

	if err != nil {
		return dune.NullValue, err
	}

	if !isHTTPSuccess(resp.StatusCode) {
		return dune.NullValue, fmt.Errorf("http Error %d: %v", resp.StatusCode, string(b))
	}

	if len(b) == 0 {
		return dune.NullValue, nil
	}

	v, err := unmarshal(b)
	if err != nil {
		return dune.NullValue, err
	}

	return v, nil
}

func (r *request) headers(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}

	headers := r.request.Header
	result := make([]dune.Value, len(headers))
	var i int
	for key := range headers {
		result[i] = dune.NewString(key)
		i++
	}
	return dune.NewArrayValues(result), nil
}

func (r *request) header(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}
	key := args[0].String()
	v := r.request.Header.Get(key)
	return dune.NewString(v), nil
}

func (r *request) setHeader(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String, dune.String); err != nil {
		return dune.NullValue, err
	}

	key := args[0].String()
	value := args[1].String()
	r.request.Header.Set(key, value)

	return dune.NullValue, nil
}

func (r *request) file(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}
	key := args[0].String()

	req := r.request

	file, header, err := req.FormFile(key)
	if err != nil {
		if err == http.ErrMissingFile {
			return dune.NullValue, nil
		}
		return dune.NullValue, err
	}

	name := filepath.Base(header.Filename)
	ctype := header.Header.Get("Content-Type")
	return dune.NewObject(newFormFile(file, name, ctype, header.Size, vm)), nil
}

func newFormFile(file multipart.File, name string, contentType string, size int64, vm *dune.VM) formFile {
	f := formFile{
		file:        file,
		name:        name,
		contentType: contentType,
		size:        size,
	}

	vm.SetGlobalFinalizer(f)
	return f
}

type formFile struct {
	file        multipart.File
	size        int64
	name        string
	contentType string
}

func (f formFile) Type() string {
	return "multipart.File"
}

func (f formFile) GetField(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "name":
		return dune.NewString(f.name), nil
	case "size":
		return dune.NewInt64(f.size), nil
	case "contentType":
		return dune.NewString(f.contentType), nil
	}
	return dune.UndefinedValue, nil
}

func (f formFile) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "close":
		return f.close
	}
	return nil
}

func (f formFile) Read(p []byte) (n int, err error) {
	return f.file.Read(p)
}

func (f formFile) ReadAt(p []byte, off int64) (n int, err error) {
	return f.file.ReadAt(p, off)
}

func (f formFile) Close() error {
	c, ok := f.file.(io.Closer)
	if ok {
		if err := c.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (f formFile) close(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if c, ok := f.file.(io.Closer); ok {
		if err := c.Close(); err != nil {
			return dune.NullValue, err
		}
	}

	return dune.NullValue, nil
}

func (r *request) addCookie(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Object); err != nil {
		return dune.NullValue, err
	}

	c, ok := args[0].ToObject().(*cookie)
	if !ok {
		return dune.NullValue, ErrInvalidType
	}

	k := &http.Cookie{
		Domain:   c.domain,
		Path:     c.path,
		Expires:  c.expires,
		Name:     c.name,
		Value:    c.value,
		Secure:   c.secure,
		HttpOnly: c.httpOnly,
		SameSite: c.sameSite,
	}

	r.request.AddCookie(k)
	return dune.NullValue, nil
}

func (r *request) cookie(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	name := args[0].String()

	k, err := r.request.Cookie(name)
	if err != nil {
		if err == http.ErrNoCookie {
			return dune.NullValue, nil
		}
		return dune.NullValue, err
	}

	c := &cookie{
		domain:   k.Domain,
		path:     k.Path,
		expires:  k.Expires,
		secure:   k.Secure,
		httpOnly: k.HttpOnly,
		name:     k.Name,
		value:    k.Value,
		sameSite: k.SameSite,
	}

	return dune.NewObject(c), nil
}

func (r *request) value(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}
	name := args[0].String()

	req := r.request
	if req.Method == "GET" {
		s := req.URL.Query().Get(name)
		if s == "" {
			return dune.NullValue, nil
		}
		return dune.NewString(s), nil
	}

	return dune.NewString(req.FormValue(name)), nil
}

func (r *request) formInt(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}
	name := args[0].String()

	var s string
	req := r.request
	if req.Method == "GET" {
		s = req.URL.Query().Get(name)
	} else {
		s = req.FormValue(name)
	}

	if s == "" || s == "undefined" {
		return dune.NullValue, nil
	}

	if s == "NaN" {
		return dune.NullValue, dune.NewTypeError("parse", "invalid format: NaN")
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return dune.NullValue, dune.NewTypeError("parse", "Invalid format: %s", s)
	}

	return dune.NewInt(i), nil
}

func (r *request) formFloat(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}
	name := args[0].String()

	var s string
	req := r.request
	if req.Method == "GET" {
		s = req.URL.Query().Get(name)
	} else {
		s = req.FormValue(name)
	}

	if s == "" || s == "undefined" {
		return dune.NullValue, nil
	}

	loc := vm.Localizer
	if loc == nil {
		loc = defaultLocalizer
	}

	f, err := loc.ParseNumber(s)
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NewFloat(f), nil
}

func (r *request) formDate(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}
	name := args[0].String()

	var value string
	req := r.request
	if req.Method == "GET" {
		value = req.URL.Query().Get(name)
	} else {
		value = req.FormValue(name)
	}

	if value == "" || value == "undefined" {
		return dune.NullValue, nil
	}

	loc := vm.Localizer
	if loc == nil {
		loc = defaultLocalizer
	}

	t, err := loc.ParseDate(value, "", GetLocation(vm))
	if err != nil {
		return dune.NullValue, dune.NewTypeError("parse", "Invalid date: %v", err)
	}

	return dune.NewObject(TimeObj(t)), nil
}

func (r *request) formBool(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}
	name := args[0].String()

	var s string
	req := r.request
	if req.Method == "GET" {
		s = req.URL.Query().Get(name)
	} else {
		s = req.FormValue(name)
	}

	switch s {
	case "true", "1", "on":
		return dune.TrueValue, nil
	default:
		return dune.FalseValue, nil
	}
}

func (r *request) formJSON(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}
	name := args[0].String()

	var s string
	req := r.request
	if req.Method == "GET" {
		s = req.URL.Query().Get(name)
	} else {
		s = req.FormValue(name)
	}

	if s == "" || s == "undefined" {
		return dune.NullValue, nil
	}

	v, err := unmarshal([]byte(s))
	if err != nil {
		return dune.NullValue, dune.NewTypeError("parse", "Invalid JSON: %v", err)
	}

	return v, nil
}

type URL struct {
	url *url.URL
}

func (*URL) Type() string {
	return "http.URL"
}

func (u *URL) SetField(name string, v dune.Value, vm *dune.VM) error {
	switch name {
	case "path":
		if !vm.HasPermission("trusted") {
			return ErrUnauthorized
		}
		if v.Type != dune.String {
			return fmt.Errorf("invalid type. Expected string")
		}
		u.url.Path = v.String()
		return nil
	case "scheme":
		if v.Type != dune.String {
			return ErrInvalidType
		}
		u.url.Scheme = v.String()
		return nil
	case "host":
		if v.Type != dune.String {
			return ErrInvalidType
		}
		u.url.Host = v.String()
		return nil
	}
	return ErrReadOnlyOrUndefined
}

func (u *URL) GetField(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "scheme":
		return dune.NewString(u.url.Scheme), nil
	case "host":
		return dune.NewString(u.url.Host), nil
	case "hostName":
		// returns the host without the port number if present
		host := u.url.Host
		i := strings.IndexRune(host, ':')
		if i != -1 {
			host = host[:i]
		}
		return dune.NewString(host), nil
	case "port":
		// returns the host without the port number if present
		host := u.url.Host
		i := strings.IndexRune(host, ':')
		if i != -1 {
			return dune.NewString(host[i+1:]), nil
		}
		return dune.NullValue, nil
	case "subdomain":
		return dune.NewString(getSubdomain(u.url.Host)), nil
	case "path":
		return dune.NewString(u.url.Path), nil
	case "query":
		return dune.NewString(u.url.RawQuery), nil
	case "pathAndQuery":
		p := u.url.Path
		q := u.url.RawQuery
		if q != "" {
			p += "?" + q
		}
		return dune.NewString(p), nil
	}
	return dune.UndefinedValue, nil
}
func (u *URL) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "string":
		return u.string
	}
	return nil
}

func (u *URL) string(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	s := u.String()
	return dune.NewString(s), nil
}

func (u *URL) String() string {
	return u.url.Scheme + u.url.String()
}

// return the subdomain if the host has a format subdomain.xxxx.xx
func getSubdomain(host string) string {
	parts := Split(host, ".")
	if len(parts) != 3 {
		return ""
	}
	return parts[0]
}

type responseWriter struct {
	writer  http.ResponseWriter
	request *http.Request
	status  int
	handled bool
}

func (*responseWriter) Type() string {
	return "http.ResponseWriter"
}

func (r *responseWriter) GetField(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "status":
		rc, ok := r.writer.(*httptest.ResponseRecorder)
		if ok {
			return dune.NewInt(rc.Code), nil
		}
		return dune.NewInt(r.status), nil
	case "handled":
		return dune.NewBool(r.handled), nil
	}
	return dune.UndefinedValue, nil
}

func (r *responseWriter) SetField(name string, v dune.Value, vm *dune.VM) error {
	switch name {
	case "handled":
		if v.Type != dune.Bool {
			return fmt.Errorf("invalid type. Expected boolean")
		}
		r.handled = v.ToBool()
		return nil
	}
	return ErrReadOnlyOrUndefined
}

func (r *responseWriter) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "addCookie":
		r.handled = true
		return r.addCookie
	case "cookie":
		return r.cookie
	case "cookies":
		return r.cookies
	case "headers":
		return r.headers
	case "header":
		return r.header
	case "write":
		r.handled = true
		return r.write
	case "writeGziped":
		r.handled = true
		return r.writeGziped
	case "writeJSON":
		r.handled = true
		return r.writeJSON
	case "writeJSONStatus":
		r.handled = true
		return r.writeJSONStatus
	case "writeFile":
		r.handled = true
		return r.writeFile
	case "redirect":
		r.handled = true
		return r.redirect
	case "setContentType":
		r.handled = true
		return r.setContentType
	case "setHeader":
		r.handled = true
		return r.setHeader
	case "setStatus":
		r.handled = true
		return r.setStatus
	case "writeError":
		r.handled = true
		return r.writeError
	case "writeJSONError":
		r.handled = true
		return r.writeJSONError
	case "body":
		return r.body
	case "json":
		return r.json
	case "bytes":
		return r.bytes
	}
	return nil
}

func (r *responseWriter) Write(p []byte) (n int, err error) {

	// set 200 by default when anything is written
	if r.status == 0 {
		r.status = 200
	}

	return r.writer.Write(p)
}

func (r *responseWriter) cookies(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}

	rc, ok := r.writer.(*httptest.ResponseRecorder)
	if !ok {
		return dune.NullValue, fmt.Errorf("the response is not a recorder")
	}

	cookies := rc.Result().Cookies()

	v := make([]dune.Value, len(cookies))

	for i, k := range cookies {
		c := &cookie{
			domain:   k.Domain,
			path:     k.Path,
			expires:  k.Expires,
			secure:   k.Secure,
			httpOnly: k.HttpOnly,
			name:     k.Name,
			value:    k.Value,
			sameSite: k.SameSite,
		}

		v[i] = dune.NewObject(c)
	}

	return dune.NewArrayValues(v), nil
}

func (r *responseWriter) cookie(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	rc, ok := r.writer.(*httptest.ResponseRecorder)
	if !ok {
		return dune.NullValue, fmt.Errorf("the response is not a recorder")
	}

	request := &http.Request{Header: http.Header{"Cookie": rc.Header()["Set-Cookie"]}}

	name := args[0].String()

	// Extract the dropped cookie from the request.
	k, err := request.Cookie(name)
	if err != nil {
		return dune.NullValue, err
	}

	c := &cookie{
		domain:   k.Domain,
		path:     k.Path,
		expires:  k.Expires,
		secure:   k.Secure,
		httpOnly: k.HttpOnly,
		name:     k.Name,
		value:    k.Value,
		sameSite: k.SameSite,
	}

	return dune.NewObject(c), nil
}

func (r *responseWriter) addCookie(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Object); err != nil {
		return dune.NullValue, err
	}

	c, ok := args[0].ToObject().(*cookie)
	if !ok {
		return dune.NullValue, ErrInvalidType
	}

	path := c.path
	if path == "" {
		path = "/"
	}

	k := &http.Cookie{
		Domain:   c.domain,
		Path:     path,
		Expires:  c.expires,
		Name:     c.name,
		Value:    c.value,
		Secure:   c.secure,
		HttpOnly: c.httpOnly,
		SameSite: c.sameSite,
	}

	http.SetCookie(r.writer, k)
	return dune.NullValue, nil
}

func (r *responseWriter) headers(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgRange(args, 0, 0); err != nil {
		return dune.NullValue, err
	}

	rc := r.writer.(*httptest.ResponseRecorder)

	headers := rc.Header()

	values := make([]dune.Value, len(headers))

	i := 0
	for k := range headers {
		values[i] = dune.NewString(k)
		i++
	}

	return dune.NewArrayValues(values), nil
}

func (r *responseWriter) header(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String); err != nil {
		return dune.NullValue, err
	}

	name := args[0].String()

	rc := r.writer.(*httptest.ResponseRecorder)

	headers := rc.Header().Values(name)

	values := make([]dune.Value, len(headers))

	for i, v := range headers {
		values[i] = dune.NewString(v)
	}

	return dune.NewArrayValues(values), nil
}

func (r *responseWriter) body(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgRange(args, 0, 0); err != nil {
		return dune.NullValue, err
	}

	rc, ok := r.writer.(*httptest.ResponseRecorder)
	if !ok {
		return dune.NullValue, fmt.Errorf("the response is not a recorder")
	}

	return dune.NewString(rc.Body.String()), nil
}

func (r *responseWriter) json(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgRange(args, 0, 0); err != nil {
		return dune.NullValue, err
	}

	rc, ok := r.writer.(*httptest.ResponseRecorder)
	if !ok {
		return dune.NullValue, fmt.Errorf("the response is not a recorder")
	}

	b, err := ioutil.ReadAll(rc.Body)

	if err != nil {
		return dune.NullValue, err
	}

	return unmarshal(b)
}

func (r *responseWriter) bytes(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgRange(args, 0, 0); err != nil {
		return dune.NullValue, err
	}

	rc, ok := r.writer.(*httptest.ResponseRecorder)
	if !ok {
		return dune.NullValue, fmt.Errorf("the response is not a recorder")
	}

	return dune.NewBytes(rc.Body.Bytes()), nil
}

func (r *responseWriter) redirect(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateOptionalArgs(args, dune.String, dune.Int); err != nil {
		return dune.NullValue, err
	}

	var url string
	var status int

	switch len(args) {
	case 0:
		return dune.NullValue, fmt.Errorf("expected at least 1 argument, got 0")
	case 1:
		url = args[0].String()
		status = http.StatusFound
	case 2:
		url = args[0].String()
		status = int(args[1].ToInt())
	default:
		return dune.NullValue, fmt.Errorf("expected at least 1 argument, got 0")
	}

	r.status = status
	http.Redirect(r.writer, r.request, url, status)

	return dune.NullValue, nil
}

func (r *responseWriter) writeGziped(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if !strings.Contains(r.request.Header.Get("Accept-Encoding"), "gzip") {
		return r.write(args, vm)
	}

	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}
	var a = args[0]
	var b []byte

	switch a.Type {
	case dune.Null, dune.Undefined:
		return dune.NullValue, nil
	case dune.String, dune.Bytes:
		b = a.ToBytes()
	default:
		b = []byte(a.String())
	}

	if err := vm.AddAllocations(len(b)); err != nil {
		return dune.NullValue, err
	}
	r.writer.Header().Set("Content-Encoding", "gzip")
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(b); err != nil {
		return dune.NullValue, err
	}
	gz.Close()
	b = buf.Bytes()
	r.writer.Write(b)
	return dune.NullValue, nil
}

func (r *responseWriter) write(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}
	var a = args[0]
	var b []byte

	switch a.Type {
	case dune.Null, dune.Undefined:
		return dune.NullValue, nil
	case dune.String, dune.Bytes:
		b = a.ToBytes()
	default:
		b = []byte(a.String())
	}

	if err := vm.AddAllocations(len(b)); err != nil {
		return dune.NullValue, err
	}

	r.writer.Write(b)

	// set 200 by default when anything is written
	if r.status == 0 {
		r.status = 200
	}

	return dune.NullValue, nil
}

func (r *responseWriter) writeError(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateOptionalArgs(args, dune.Int, dune.String); err != nil {
		return dune.NullValue, err
	}
	l := len(args)
	if l == 0 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	code := int(args[0].ToInt())

	r.status = code

	r.writer.WriteHeader(code)

	if l == 2 {
		r.writer.Write(args[1].ToBytes())
	} else {
		switch code {
		case 400:
			r.writer.Write([]byte("Bad Request"))
		case 401:
			r.writer.Write([]byte("Unauthorized"))
		case 403:
			r.writer.Write([]byte("Forbidden"))
		case 404:
			r.writer.Write([]byte("Not Found"))
		default:
			r.writer.Write([]byte(Translate(genericError, vm)))
		}
	}

	r.writer.Write([]byte("\n"))
	return dune.NullValue, nil
}

func (r *responseWriter) writeJSONError(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateOptionalArgs(args, dune.Int, dune.String); err != nil {
		return dune.NullValue, err
	}
	l := len(args)
	if l == 0 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	code := int(args[0].ToInt())

	r.status = code

	var err []byte

	if l == 2 {
		err = args[1].ToBytes()
	} else {
		switch code {
		case 400:
			err = []byte("Bad Request")
		case 401:
			err = []byte("Unauthorized")
		case 404:
			err = []byte("Not Found")
		default:
			err = []byte(Translate(genericError, vm))
		}
	}

	r.writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	r.writer.WriteHeader(code)
	r.writer.Write([]byte(`{"error":"`))
	r.writer.Write(err)
	r.writer.Write([]byte("\"}\n"))
	return dune.NullValue, nil
}

func (r *responseWriter) writeJSONStatus(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	l := len(args)
	if l < 2 {
		return dune.NullValue, fmt.Errorf("expected 2 arguments, got %d", l)
	}

	a := args[0]
	if a.Type != dune.Int {
		return dune.NullValue, fmt.Errorf("expected argument 1 status of type int, got %s", a.TypeName())
	}

	return r.doWriteJSON(int(a.ToInt()), args[1:], vm)
}

func (r *responseWriter) writeJSON(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	return r.doWriteJSON(200, args, vm)
}

func (r *responseWriter) doWriteJSON(status int, args []dune.Value, vm *dune.VM) (dune.Value, error) {
	l := len(args)
	if l == 0 || l > 2 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", l)
	}

	v := args[0].ExportMarshal(0)

	var skipCacheHeader bool
	if l > 1 {
		i := args[1]
		if i.Type != dune.Bool {
			return dune.NullValue, fmt.Errorf("expected argument 2 of type boolean, got %s", i.TypeName())
		}
		skipCacheHeader = i.ToBool()
	}

	b, err := json.Marshal(v)

	if err != nil {
		return dune.NullValue, err
	}

	if err := vm.AddAllocations(len(b)); err != nil {
		return dune.NullValue, err
	}

	b = append(b, '\n')

	w := r.writer

	var compress bool

	// don't write again headers if there have already been written
	if r.status == 0 {
		compress = strings.Contains(r.request.Header.Get("Accept-Encoding"), "gzip")
		if compress {
			w.Header().Set("Content-Encoding", "gzip")
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		if !skipCacheHeader {
			w.Header().Set("Cache-Breaker", cacheBreaker)
		}

		r.status = status
		r.writer.WriteHeader(status)
	}

	if compress {
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		if _, err := gz.Write(b); err != nil {
			return dune.NullValue, err
		}
		gz.Close()
		b = buf.Bytes()
	}

	w.Write(b)
	return dune.NullValue, nil
}

var contentDate = time.Date(2016, 3, 4, 0, 0, 0, 0, time.UTC)

func (r *responseWriter) writeFile(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	ln := len(args)
	if ln != 2 {
		return dune.NullValue, fmt.Errorf("expected 2 or 3 arguments, got %d", ln)
	}

	if args[0].Type != dune.String {
		return dune.NullValue, fmt.Errorf("expected argument to be a string, got %v", args[0].TypeName())
	}

	// set 200 by default when anything is written
	if r.status == 0 {
		r.status = 200
	}

	name := args[0].String()

	var reader io.ReadSeeker

	var lastModified time.Time

	switch args[1].Type {
	case dune.String:
	case dune.Bytes:
		data := []byte(args[1].ToBytes())
		reader = bytes.NewReader(data)
		lastModified = contentDate
	case dune.Object:
		f, ok := args[1].ToObject().(io.ReadSeeker)
		if ok {
			reader = f
			lastModified = contentDate
		} else {
			fs, ok := args[1].ToObject().(*FileSystemObj)
			if !ok {
				return dune.NullValue, ErrInvalidType
			}
			fi, err := fs.FS.Stat(name)
			if err != nil {
				return dune.NullValue, err
			}
			lastModified = fi.ModTime()
			f, err := fs.FS.Open(name)
			if err != nil {
				return dune.NullValue, err
			}
			reader = f
			defer f.Close()
		}
	default:
		return dune.NullValue, ErrInvalidType
	}

	if strings.Contains(r.request.Header.Get("Accept-Encoding"), "gzip") {
		serveGziped(r.writer, r.request, filepath.Base(name), lastModified, reader)
	} else {
		http.ServeContent(r.writer, r.request, filepath.Base(name), contentDate, reader)
	}
	return dune.NullValue, nil
}

func serveGziped(w http.ResponseWriter, r *http.Request, name string, lastModified time.Time, f io.ReadSeeker) {
	w.Header().Set("Content-Encoding", "gzip")
	gz := gzip.NewWriter(w)
	defer gz.Close()

	gw := &gzipResponseWriter{Writer: gz, ResponseWriter: w}
	http.ServeContent(gw, r, name, lastModified, f)
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
	sniffDone bool
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if !w.sniffDone {
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", http.DetectContentType(b))
		}
		w.sniffDone = true
	}

	return w.Writer.Write(b)
}

func (r *responseWriter) setContentType(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	var a = args[0]
	if a.Type != dune.String {
		return dune.NullValue, fmt.Errorf("expected a string, %s", a.TypeName())
	}

	r.writer.Header().Set("Content-Type", a.String())
	return dune.NullValue, nil
}

func (r *responseWriter) setHeader(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 2 {
		return dune.NullValue, fmt.Errorf("expected 2 arguments, got %d", len(args))
	}

	var a = args[0]
	if a.Type != dune.String {
		return dune.NullValue, fmt.Errorf("expected a string, %s", a.TypeName())
	}

	var b = args[1]
	if b.Type != dune.String {
		return dune.NullValue, fmt.Errorf("expected a string, %s", a.TypeName())
	}

	r.writer.Header().Set(a.String(), b.String())
	return dune.NullValue, nil
}

func (r *responseWriter) setStatus(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}

	var a = args[0]
	if a.Type != dune.Int {
		return dune.NullValue, fmt.Errorf("expected a int, %s", a.TypeName())
	}

	s := int(a.ToInt())
	r.writer.WriteHeader(s)
	r.status = s
	return dune.NullValue, nil
}
