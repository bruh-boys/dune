package lib

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/acme/autocert"

	"github.com/dunelang/dune"
	"github.com/dunelang/dune/filesystem"
)

func init() {
	dune.RegisterLib(TLS, `

declare namespace tls {
    export function newConfig(insecureSkipVerify?: boolean): Config

    export interface Config {
		insecureSkipVerify: boolean
		certManager: autocert.CertManager
        loadCertificate(certPath: string, keyPath: string): void
        loadCertificateData(cert: byte[] | string, key: byte[] | string): void
	}

	export interface Certificate {
		cert: byte[]
		key: byte[]
	}
	
	export function generateCert(): Certificate 
}

declare namespace autocert {
	export interface CertManager {

	}

	export function newCertManager(dirCache: string, domains: string[], cache?: Cache): CertManager
	export function newCertManager(dirCache: string, hostPolicy: (host: string) => void, cache?: Cache): CertManager

	export interface Cache {
	}
	export function newFileSystemCache(fs: io.FileSystem): Cache
}

`)
}

var TLS = []dune.NativeFunction{
	{
		Name:      "autocert.newCertManager",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgRange(args, 2, 3); err != nil {
				return dune.NullValue, err
			}

			var cache *autocertCache

			ln := len(args)
			if ln < 2 || ln > 3 {
				return dune.NullValue, fmt.Errorf("expected 2 or 3 arguments, got %d", ln)
			}

			if args[0].Type != dune.String {
				return dune.NullValue, fmt.Errorf("invalid cache directory: %s", args[0].TypeName())
			}

			cacheDir := args[0].String()

			var hostPolicy autocert.HostPolicy
			switch args[1].Type {
			case dune.Array:
				domainValues := args[1].ToArray()
				domains := make([]string, len(domainValues))
				for i, v := range domainValues {
					domains[i] = v.String()
				}
				hostPolicy = autocert.HostWhitelist(domains...)

			case dune.Func:
				hostPolicy = func(ctx context.Context, host string) error {
					vm = vm.CloneInitialized(vm.Program, vm.Globals())
					_, err := vm.RunFuncIndex(args[1].ToFunction(), dune.NewString(host))
					return err
				}

			case dune.Object:
				c, ok := args[1].ToObjectOrNil().(*dune.Closure)
				if !ok {
					return dune.NullValue, fmt.Errorf("%v is not a function", args[1].TypeName())
				}
				hostPolicy = func(ctx context.Context, host string) error {
					vm = vm.CloneInitialized(vm.Program, vm.Globals())
					_, err := vm.RunClosure(c, dune.NewString(host))
					return err
				}

			default:
				return dune.NullValue, fmt.Errorf("invalid domains or hostPolicy: %s", args[1].TypeName())
			}

			if ln == 3 {
				var ok bool
				cache, ok = args[2].ToObjectOrNil().(*autocertCache)
				if !ok {
					return dune.NullValue, fmt.Errorf("invalid cache: %s", args[2].TypeName())
				}
			}

			if cache == nil {
				cache = &autocertCache{fs: vm.FileSystem}
			}

			cache.dir = cacheDir

			m := autocert.Manager{
				Cache:      cache,
				Prompt:     autocert.AcceptTOS,
				HostPolicy: hostPolicy,
			}

			cm := &certManager{&m}
			return dune.NewObject(cm), nil
		},
	},
	{
		Name:      "autocert.newFileSystemCache",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Object); err != nil {
				return dune.NullValue, err
			}

			fs, ok := args[0].ToObject().(*FileSystemObj)
			if !ok {
				return dune.NullValue, fmt.Errorf("invalid filesystem argument, got %v", args[0])
			}

			c := &autocertCache{fs: fs.FS}
			return dune.NewObject(c), nil
		},
	},
	{
		Name:      "tls.newConfig",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgRange(args, 0, 1); err != nil {
				return dune.NullValue, err
			}
			if err := ValidateOptionalArgs(args, dune.Bool); err != nil {
				return dune.NullValue, err
			}

			tc := &tlsConfig{
				conf: &tls.Config{
					MinVersion:               tls.VersionTLS12,
					CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
					PreferServerCipherSuites: true,
				},
			}

			if len(args) == 1 {
				tc.conf.InsecureSkipVerify = args[0].ToBool()
			}

			return dune.NewObject(tc), nil
		},
	},
	{
		Name:      "tls.generateCert",
		Arguments: 0,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			priv, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
			if err != nil {
				log.Fatal(err)
			}

			// https://godoc.org/github.com/rocketlaunchr/https-go

			// type GenerateOptions struct {
			// 	// Comma-separated hostnames and IPs to generate a certificate for.
			// 	Host string

			// 	// Creation date formatted as "Jan 1 15:04:05 2011".
			// 	// Default is time.Now().
			// 	ValidFrom string

			// 	// Duration that certificate is valid for.
			// 	// Default is 365*24*time.Hour.
			// 	ValidFor time.Duration

			// 	// Whether this cert should be its own Certificate Authority
			// 	// Default is false.
			// 	IsCA bool

			// 	// Size of RSA key to generate. Ignored if ECDSACurve is set
			// 	// Default is 2048.
			// 	RSABits int

			// 	// ECDSA curve to use to generate a key. Valid values are P224, P256 (recommended), P384, P521
			// 	// Default is "".
			// 	ECDSACurve string

			// 	// Generate an Ed25519 key
			// 	// Default is false.
			// 	ED25519Key bool
			// }

			template := x509.Certificate{
				SerialNumber: big.NewInt(1),
				Subject: pkix.Name{
					Organization: []string{"Acme Co"},
				},
				NotBefore: time.Now(),
				NotAfter:  time.Now().Add(time.Hour * 24 * 180),

				KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
				ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
				BasicConstraintsValid: true,
			}

			/*
				   hosts := strings.Split(*host, ",")
				   for _, h := range hosts {
					   if ip := net.ParseIP(h); ip != nil {
						   template.IPAddresses = append(template.IPAddresses, ip)
					   } else {
						   template.DNSNames = append(template.DNSNames, h)
					   }
				   }
				   if *isCA {
					   template.IsCA = true
					   template.KeyUsage |= x509.KeyUsageCertSign
				   }
			*/

			derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
			if err != nil {
				return dune.Value{}, fmt.Errorf("failed to create certificate: %s", err)
			}

			// Create public key
			pubBuf := new(bytes.Buffer)
			err = pem.Encode(pubBuf, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
			if err != nil {
				return dune.Value{}, fmt.Errorf("failed to write data to cert.pem: %w", err)
			}

			// Create private key
			privBuf := new(bytes.Buffer)
			privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
			if err != nil {
				return dune.Value{}, fmt.Errorf("unable to marshal private key: %w", err)
			}

			err = pem.Encode(privBuf, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
			if err != nil {
				return dune.Value{}, fmt.Errorf("failed to write data to key.pem: %w", err)
			}

			c := &certificate{
				cert: pubBuf.Bytes(),
				key:  privBuf.Bytes(),
			}

			return dune.NewObject(c), nil
		},
	},
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

type certificate struct {
	cert []byte
	key  []byte
}

func (c *certificate) Type() string {
	return "tls.Certificate"
}

func (c *certificate) GetField(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "cert":
		return dune.NewBytes(c.cert), nil
	case "key":
		return dune.NewBytes(c.key), nil
	}
	return dune.UndefinedValue, nil
}

type certManager struct {
	manager *autocert.Manager
}

func (c *certManager) Type() string {
	return "autocert.CertManager"
}

type tlsConfig struct {
	conf        *tls.Config
	certManager *certManager
}

func (t *tlsConfig) Type() string {
	return "tls.Config"
}

func (t *tlsConfig) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "loadCertificate":
		return t.loadCertificate
	case "loadCertificateData":
		return t.loadCertificateData
	}
	return nil
}

func (t *tlsConfig) GetField(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "insecureSkipVerify":
		return dune.NewBool(t.conf.InsecureSkipVerify), nil
	case "certManager":
		if t.certManager == nil {
			return dune.NullValue, nil
		}
		return dune.NewObject(t.certManager), nil
	}
	return dune.UndefinedValue, nil
}

func (t *tlsConfig) SetField(name string, v dune.Value, vm *dune.VM) error {
	switch name {
	case "insecureSkipVerify":
		if v.Type != dune.Bool {
			return fmt.Errorf("invalid type, expected bool")
		}
		t.conf.InsecureSkipVerify = v.ToBool()
		return nil
	case "certManager":
		c, ok := v.ToObjectOrNil().(*certManager)
		if !ok {
			return fmt.Errorf("invalid type, expected CertManager")
		}
		t.certManager = c
		t.conf.GetCertificate = c.manager.GetCertificate
		return nil
	}

	return ErrReadOnlyOrUndefined
}

func (t *tlsConfig) loadCertificate(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.String, dune.String); err != nil {
		return dune.NullValue, err
	}

	fs := vm.FileSystem
	if fs == nil {
		return dune.NullValue, fmt.Errorf("there is no filesystem set")
	}

	certPath := args[0].String()
	keyPath := args[1].String()

	certPEMBlock, err := filesystem.ReadAll(fs, certPath)
	if err != nil {
		return dune.NullValue, fmt.Errorf("error reading cert %s: %w", certPath, err)
	}

	keyPEMBlock, err := filesystem.ReadAll(fs, keyPath)
	if err != nil {
		return dune.NullValue, fmt.Errorf("error reading key %s: %w", keyPath, err)
	}

	cert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		return dune.NullValue, fmt.Errorf("error creating X509KeyPair: %w", err)
	}

	t.conf.Certificates = append(t.conf.Certificates, cert)

	return dune.NullValue, nil
}

func (t *tlsConfig) loadCertificateData(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgRange(args, 2, 2); err != nil {
		return dune.NullValue, err
	}

	certPEMBlock := args[0]
	keyPEMBlock := args[1]

	switch certPEMBlock.Type {
	case dune.String, dune.Bytes:
	default:
		return dune.NullValue, fmt.Errorf("expected cert of type string or bytes, got: %s", certPEMBlock.TypeName())
	}

	switch keyPEMBlock.Type {
	case dune.String, dune.Bytes:
	default:
		return dune.NullValue, fmt.Errorf("expected key of type string or bytes, got: %s", keyPEMBlock.TypeName())
	}

	cert, err := tls.X509KeyPair(certPEMBlock.ToBytes(), keyPEMBlock.ToBytes())
	if err != nil {
		return dune.NullValue, fmt.Errorf("error creating X509KeyPair: %w", err)
	}

	t.conf.Certificates = append(t.conf.Certificates, cert)

	return dune.NullValue, nil
}

type autocertCache struct {
	fs  filesystem.FS
	dir string
}

func (c *autocertCache) Type() string {
	return "autocert.Cache"
}

func (c *autocertCache) Get(ctx context.Context, name string) ([]byte, error) {
	name = filepath.Join(c.dir, name)
	var (
		data []byte
		err  error
		done = make(chan struct{})
	)
	go func() {
		data, err = filesystem.ReadAll(c.fs, name)
		close(done)
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-done:
	}
	if os.IsNotExist(err) {
		return nil, autocert.ErrCacheMiss
	}
	return data, err
}

func (c *autocertCache) Put(ctx context.Context, name string, data []byte) error {
	if err := c.fs.MkdirAll(c.dir); err != nil {
		return err
	}

	done := make(chan struct{})
	var err error
	go func() {
		defer close(done)
		select {
		case <-ctx.Done():
			// Don't overwrite the file if the context was canceled.
		default:
			name = filepath.Join(c.dir, name)
			err = c.fs.Write(name, data)
		}
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
	}
	return err
}

func (c *autocertCache) Delete(ctx context.Context, name string) error {
	name = filepath.Join(c.dir, name)
	var (
		err  error
		done = make(chan struct{})
	)
	go func() {
		err = c.fs.RemoveAll(name)
		close(done)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
	}
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
