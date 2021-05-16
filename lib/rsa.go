package lib

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(RSA, `

declare namespace rsa {
    export function generateKey(size?: number): PrivateKey
    export function decodePEMKey(key: string | byte[]): PrivateKey
    export function decodePublicPEMKey(key: string | byte[]): PublicKey
    export function signPKCS1v15(key: PrivateKey, mesage: string | byte[]): byte[]
    export function verifyPKCS1v15(key: PublicKey, mesage: string | byte[], signature: string | byte[]): boolean

    interface PrivateKey {
        publicKey: PublicKey
        encodePEMKey(): byte[]
        encodePublicPEMKey(): byte[]
    }

    interface PublicKey {

    }
}

`)
}

var RSA = []dune.NativeFunction{
	{
		Name:      "rsa.generateKey",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateOptionalArgs(args, dune.Int); err != nil {
				return dune.NullValue, err
			}

			reader := rand.Reader

			var bitSize int
			if len(args) == 0 {
				bitSize = 2048
			} else {
				bitSize = int(args[0].ToInt())
			}

			key, err := rsa.GenerateKey(reader, bitSize)
			if err != nil {
				return dune.NullValue, err
			}

			return dune.NewObject(&rsaPrivateKey{key}), nil
		},
	},
	{
		Name:      "rsa.decodePEMKey",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			v := args[0].ToBytes()

			block, _ := pem.Decode(v)

			if block == nil {
				return dune.NullValue, fmt.Errorf("error decoding private key")
			}

			enc := x509.IsEncryptedPEMBlock(block)

			b := block.Bytes

			var err error
			if enc {
				b, err = x509.DecryptPEMBlock(block, nil)
				if err != nil {
					return dune.NullValue, fmt.Errorf("error decrypting private key")
				}
			}

			key, err := x509.ParsePKCS1PrivateKey(b)
			if err != nil {
				return dune.NullValue, fmt.Errorf("error parsing private key: %w", err)
			}

			return dune.NewObject(&rsaPrivateKey{key}), nil
		},
	},
	{
		Name:      "rsa.decodePublicPEMKey",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			v := args[0].ToBytes()

			block, _ := pem.Decode(v)

			if block == nil {
				return dune.NullValue, fmt.Errorf("error decoding public key")
			}

			enc := x509.IsEncryptedPEMBlock(block)

			b := block.Bytes

			var err error
			if enc {
				b, err = x509.DecryptPEMBlock(block, nil)
				if err != nil {
					return dune.NullValue, fmt.Errorf("error decrypting public key")
				}
			}
			ifc, err := x509.ParsePKIXPublicKey(b)
			if err != nil {
				return dune.NullValue, fmt.Errorf("error parsing public key: %v", err)
			}

			key, ok := ifc.(*rsa.PublicKey)
			if !ok {
				return dune.NullValue, fmt.Errorf("not an RSA public key")
			}

			return dune.NewObject(&rsaPublicKey{key}), nil
		},
	},
	{
		Name:      "rsa.signPKCS1v15",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			key, ok := args[0].ToObjectOrNil().(*rsaPrivateKey)
			if !ok {
				return dune.NullValue, fmt.Errorf("expected a rsa key, got %v", args[0].TypeName())
			}

			message := args[1].ToBytes()

			// Only small messages can be signed directly; thus the hash of a
			// message, rather than the message itself, is signed. This requires
			// that the hash function be collision resistant. SHA-256 is the
			// least-strong hash function that should be used for this at the time
			// of writing (2016).
			hashed := sha256.Sum256(message)

			rng := rand.Reader

			signature, err := rsa.SignPKCS1v15(rng, key.key, crypto.SHA256, hashed[:])
			if err != nil {
				return dune.NullValue, err
			}

			return dune.NewBytes(signature), nil
		},
	},
	{
		Name:      "rsa.verifyPKCS1v15",
		Arguments: 3,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			key, ok := args[0].ToObjectOrNil().(*rsaPublicKey)
			if !ok {
				return dune.NullValue, fmt.Errorf("expected a rsa key, got %v", args[0].TypeName())
			}

			message := args[1].ToBytes()
			signature := args[2].ToBytes()

			// Only small messages can be signed directly; thus the hash of a
			// message, rather than the message itself, is signed. This requires
			// that the hash function be collision resistant. SHA-256 is the
			// least-strong hash function that should be used for this at the time
			// of writing (2016).
			hashed := sha256.Sum256(message)

			err := rsa.VerifyPKCS1v15(key.key, crypto.SHA256, hashed[:], signature)
			if err != nil {
				return dune.FalseValue, err
			}

			return dune.TrueValue, err
		},
	},
}

type rsaPrivateKey struct {
	key *rsa.PrivateKey
}

func (k *rsaPrivateKey) Type() string {
	return "RSA_Private_Key"
}

func (k *rsaPrivateKey) GetField(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "publicKey":
		return dune.NewObject(&rsaPublicKey{&k.key.PublicKey}), nil
	}

	return dune.UndefinedValue, nil
}

func (k *rsaPrivateKey) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "encodePEMKey":
		return k.encodePEMKey
	case "encodePublicPEMKey":
		return k.encodePublicPEMKey
	}
	return nil
}

func (k *rsaPrivateKey) encodePEMKey(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	b := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(k.key),
		},
	)
	return dune.NewBytes(b), nil
}

func (k *rsaPrivateKey) encodePublicPEMKey(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	pubASN1, err := x509.MarshalPKIXPublicKey(k.key.Public())
	if err != nil {
		return dune.NullValue, err
	}

	b := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubASN1,
	})

	return dune.NewBytes(b), nil
}

type rsaPublicKey struct {
	key *rsa.PublicKey
}

func (k *rsaPublicKey) Type() string {
	return "RSA_Private_Key"
}
