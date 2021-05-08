package lib

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"sync"

	"github.com/dunelang/dune"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
)

func init() {
	dune.RegisterLib(Crypt, `

declare namespace crypto {
    export function signSHA1_RSA_PCKS1(privateKey: string, value: string): byte[]

    export function signTempSHA1(value: string): string
    export function checkTempSignSHA1(value: string, hash: string): boolean

    export function signSHA1(value: string): string
    export function checkSignSHA1(value: string, hash: string): boolean

    export function setGlobalPassword(pwd: string): void
    export function encrypt(value: byte[], pwd?: byte[]): byte[]
    export function decrypt(value: byte[], pwd?: byte[]): byte[]
    export function encryptTripleDES(value: byte[] | string, pwd?: byte[] | string): byte[]
    export function decryptTripleDES(value: byte[] | string, pwd?: byte[] | string): byte[]
    export function encryptString(value: string, pwd?: string): string
    export function decryptString(value: string, pwd?: string): string
    export function hashSHA(value: string): string
    export function hashSHA256(value: string): string
    export function hashSHA512(value: string): string
    export function hmacSHA256(value: byte[] | string, pwd?: byte[] | string): byte[]
    export function hmacSHA512(value: byte[] | string, pwd?: byte[] | string): byte[]
    export function hashPassword(pwd: string): string
    export function compareHashAndPassword(hash: string, pwd: string): boolean
	export function rand(n: number): number
    export function random(len: number): byte[]
    export function randomAlphanumeric(len: number): string
}


`)
}

var gPassMut sync.RWMutex
var globalPassword string
var tempSignKey = RandomAlphanumeric(30)

func setGlobalPassword(v string) {
	gPassMut.Lock()
	globalPassword = v
	gPassMut.Unlock()
}

func getGlobalPassword() string {
	var v string
	gPassMut.RLock()
	v = globalPassword
	gPassMut.RUnlock()
	return v
}

var Crypt = []dune.NativeFunction{
	{
		Name:      "crypto.signSHA1_RSA_PCKS1",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String, dune.String); err != nil {
				return dune.NullValue, err
			}

			key := args[0].ToBytes()
			text := args[1].String()

			h := sha1.New()
			h.Write([]byte(text))
			sum := h.Sum(nil)

			block, _ := pem.Decode(key)
			if block == nil {
				return dune.NullValue, fmt.Errorf("error parsing private key")
			}

			privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				return dune.NullValue, fmt.Errorf("error parsing private key: %w", err)
			}

			sig, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA1, sum)
			if err != nil {
				return dune.NullValue, fmt.Errorf("error signing: %w", err)
			}

			return dune.NewBytes(sig), nil
		},
	},
	{
		Name:      "crypto.signSHA1",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String); err != nil {
				return dune.NullValue, err
			}

			s := args[0].String() + getGlobalPassword()
			h := sha1.New()
			h.Write([]byte(s))
			hash := hex.EncodeToString(h.Sum(nil))
			return dune.NewString(hash), nil
		},
	},
	{
		Name:      "crypto.checkSignSHA1",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String, dune.String); err != nil {
				return dune.NullValue, err
			}

			s := args[0].String() + getGlobalPassword()
			h := sha1.New()
			h.Write([]byte(s))
			hash := hex.EncodeToString(h.Sum(nil))
			ok := hash == args[1].String()

			return dune.NewBool(ok), nil
		},
	},
	{
		Name:        "crypto.signTempSHA1",
		Arguments:   1,
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			// untrusted users can check but not sign

			if err := ValidateArgs(args, dune.String); err != nil {
				return dune.NullValue, err
			}

			s := args[0].String() + tempSignKey
			h := sha1.New()
			h.Write([]byte(s))
			hash := hex.EncodeToString(h.Sum(nil))
			return dune.NewString(hash), nil
		},
	},
	{
		Name:      "crypto.checkTempSignSHA1",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String, dune.String); err != nil {
				return dune.NullValue, err
			}

			// untrusted users can check but not sign

			s := args[0].String() + tempSignKey
			h := sha1.New()
			h.Write([]byte(s))
			hash := hex.EncodeToString(h.Sum(nil))

			ok := hash == args[1].String()
			return dune.NewBool(ok), nil
		},
	},
	{
		Name:        "crypto.setGlobalPassword",
		Arguments:   1,
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {

			if err := ValidateArgs(args, dune.String); err != nil {
				return dune.NullValue, err
			}

			setGlobalPassword(args[0].String())
			return dune.NullValue, nil
		},
	},
	{
		Name:      "crypto.hmacSHA256",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Bytes, dune.Bytes); err != nil {
				return dune.NullValue, err
			}

			msg := args[0].ToBytes()
			key := args[1].ToBytes()

			sig := hmac.New(sha256.New, key)
			sig.Write(msg)
			hash := sig.Sum(nil)

			return dune.NewBytes(hash), nil
		},
	},
	{
		Name:      "crypto.hmacSHA512",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Bytes, dune.Bytes); err != nil {
				return dune.NullValue, err
			}

			msg := args[0].ToBytes()
			key := args[1].ToBytes()

			sig := hmac.New(sha512.New, key)
			sig.Write(msg)
			hash := sig.Sum(nil)

			return dune.NewBytes(hash), nil
		},
	},
	{
		Name:      "crypto.hashSHA",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateOptionalArgs(args, dune.String); err != nil {
				return dune.NullValue, err
			}

			h := sha1.New()
			h.Write([]byte(args[0].String()))
			hash := hex.EncodeToString(h.Sum(nil))
			return dune.NewString(hash), nil
		},
	},
	{
		Name:      "crypto.hashSHA256",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateOptionalArgs(args, dune.String); err != nil {
				return dune.NullValue, err
			}

			h := sha256.New()
			h.Write([]byte(args[0].String()))
			hash := hex.EncodeToString(h.Sum(nil))
			return dune.NewString(hash), nil
		},
	},
	{
		Name:      "crypto.hashSHA512",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateOptionalArgs(args, dune.String); err != nil {
				return dune.NullValue, err
			}

			h := sha512.New()
			h.Write([]byte(args[0].String()))
			hash := hex.EncodeToString(h.Sum(nil))
			return dune.NewString(hash), nil
		},
	},
	{
		Name:      "crypto.encryptString",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateOptionalArgs(args, dune.String, dune.String); err != nil {
				return dune.NullValue, err
			}

			var pwd string
			switch len(args) {
			case 0:
				return dune.NullValue, fmt.Errorf("expected 1 argument, got 0")
			case 1:
				pwd = getGlobalPassword()
				if pwd == "" {
					return dune.NullValue, fmt.Errorf("no password configured")
				}
			case 2:
				pwd = args[1].String()
			}

			s, err := Encrypts(args[0].String(), pwd)
			if err != nil {
				return dune.NullValue, err
			}
			return dune.NewString(s), nil
		},
	},
	{
		Name:      "crypto.decryptString",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateOptionalArgs(args, dune.String, dune.String); err != nil {
				return dune.NullValue, err
			}

			var pwd string
			switch len(args) {
			case 0:
				return dune.NullValue, fmt.Errorf("expected 1 argument, got 0")
			case 1:
				pwd = getGlobalPassword()
				if pwd == "" {
					return dune.NullValue, fmt.Errorf("no password configured")
				}
			case 2:
				pwd = args[1].String()
			}

			s, err := Decrypts(args[0].String(), pwd)
			if err != nil {
				return dune.NullValue, err
			}
			return dune.NewString(s), nil
		},
	},
	{
		Name:      "crypto.encrypt",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateOptionalArgs(args, dune.String, dune.String); err != nil {
				return dune.NullValue, err
			}

			var pwd string
			switch len(args) {
			case 0:
				return dune.NullValue, fmt.Errorf("expected 1 argument, got 0")
			case 1:
				pwd = getGlobalPassword()
				if pwd == "" {
					return dune.NullValue, fmt.Errorf("no password configured")
				}

			case 2:
				pwd = args[1].String()
			}

			b, err := Encrypt(args[0].ToBytes(), []byte(pwd))
			if err != nil {
				return dune.NullValue, err
			}
			return dune.NewBytes(b), nil
		},
	},
	{
		Name:      "crypto.decrypt",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateOptionalArgs(args, dune.String, dune.String); err != nil {
				return dune.NullValue, err
			}

			var pwd string
			switch len(args) {
			case 0:
				return dune.NullValue, fmt.Errorf("expected 1 argument, got 0")
			case 1:
				pwd = getGlobalPassword()
				if pwd == "" {
					return dune.NullValue, fmt.Errorf("no password configured")
				}

			case 2:
				pwd = args[1].String()
			}

			b, err := Decrypt(args[0].ToBytes(), []byte(pwd))
			if err != nil {
				return dune.NullValue, err
			}
			return dune.NewBytes(b), nil
		},
	},
	{
		Name:      "crypto.encryptTripleDES",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Bytes, dune.Bytes); err != nil {
				return dune.NullValue, err
			}
			b, err := EncryptTripleDESCBC(args[0].ToBytes(), args[1].ToBytes())
			if err != nil {
				return dune.NullValue, err
			}
			return dune.NewBytes(b), nil
		},
	},
	{
		Name:      "crypto.decryptTripleDES",
		Arguments: -1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Bytes, dune.Bytes); err != nil {
				return dune.NullValue, err
			}
			b, err := DecryptTripleDESCBC(args[0].ToBytes(), args[1].ToBytes())
			if err != nil {
				return dune.NullValue, err
			}
			return dune.NewBytes(b), nil
		},
	},
	{
		Name:      "crypto.hashPassword",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String); err != nil {
				return dune.NullValue, err
			}

			s := HashPassword(args[0].String())
			return dune.NewString(s), nil
		},
	},
	{
		Name:      "crypto.compareHashAndPassword",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String, dune.String); err != nil {
				return dune.NullValue, err
			}

			ok := CheckHashPasword(args[0].String(), args[1].String())
			return dune.NewBool(ok), nil
		},
	},
	{
		Name:      "crypto.rand",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Int); err != nil {
				return dune.NullValue, err
			}
			ln := int(args[0].ToInt())
			v, err := rand.Int(rand.Reader, big.NewInt(int64(ln)))
			if err != nil {
				return dune.NullValue, err
			}
			return dune.NewInt64(v.Int64()), nil
		},
	},
	{
		Name:      "crypto.random",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Int); err != nil {
				return dune.NullValue, err
			}

			b := Random(int(args[0].ToInt()))
			return dune.NewBytes(b), nil
		},
	},
	{
		Name:      "crypto.randomAlphanumeric",
		Arguments: 1,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.Int); err != nil {
				return dune.NullValue, err
			}
			ln := int(args[0].ToInt())
			if ln < 1 {
				return dune.NullValue, fmt.Errorf("invalid len: %d", ln)
			}
			s := RandomAlphanumeric(ln)
			return dune.NewString(s), nil
		},
	},
}

const saltLen = 18

func HashPassword(pwd string) string {
	h, err := bcrypt.GenerateFromPassword([]byte(pwd), 12)
	if err != nil {
		// this should only happen if the factor is invalid, but we know it is ok
		panic(err)
	}
	return string(h)
}

func CheckHashPasword(hash, pwd string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pwd)) == nil
}

// Encrypts encrypts the text.
func Encrypts(text, password string) (string, error) {
	e, err := Encrypt([]byte(text), []byte(password))
	if err != nil {
		return "", err
	}

	encoder := base64.StdEncoding.WithPadding(base64.NoPadding)
	return encoder.EncodeToString(e), nil
}

// Decrypts decrypts the text.
func Decrypts(text, password string) (string, error) {
	encoder := base64.StdEncoding.WithPadding(base64.NoPadding)
	e, err := encoder.DecodeString(text)
	if err != nil {
		return "", err
	}

	d, err := Decrypt(e, []byte(password))
	if err != nil {
		return "", err
	}

	return string(d), err
}

func EncryptTripleDESCBC(decrypted, key []byte) ([]byte, error) {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}

	iv := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	blockMode := cipher.NewCBCEncrypter(block, iv)

	decrypted = ZeroPadding(decrypted, block.BlockSize())
	encrypted := make([]byte, len(decrypted))
	blockMode.CryptBlocks(encrypted, decrypted)
	return encrypted, nil
}

func DecryptTripleDESCBC(encrypted, key []byte) ([]byte, error) {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}

	iv := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	blockMode := cipher.NewCBCDecrypter(block, iv)

	decrypted := make([]byte, len(encrypted))
	blockMode.CryptBlocks(decrypted, encrypted)
	decrypted = ZeroUnPadding(decrypted)
	return decrypted, nil
}

func ZeroPadding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{0}, padding)
	return append(ciphertext, padtext...)
}

func ZeroUnPadding(origData []byte) []byte {
	return bytes.TrimFunc(origData,
		func(r rune) bool {
			return r == rune(0)
		})
}

// Encrypts encrypts the text.
func Encrypt(plaintext, password []byte) ([]byte, error) {
	key, salt := generateFromPassword(password)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	return append(salt, gcm.Seal(nonce, nonce, plaintext, nil)...), nil
}

// Decrypts decrypts the text.
func Decrypt(ciphertext, password []byte) ([]byte, error) {
	salt, c, err := decode(ciphertext)
	if err != nil {
		return nil, err
	}

	key := generateFromPasswordAndSalt(password, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return gcm.Open(nil, c[:gcm.NonceSize()], c[gcm.NonceSize():], nil)
}

// decode returns the salt and cipertext
func decode(ciphertext []byte) ([]byte, []byte, error) {
	if len(ciphertext) < saltLen {
		return nil, nil, fmt.Errorf("invalid ciphertext")
	}
	return ciphertext[:saltLen], ciphertext[saltLen:], nil
}

func generateFromPasswordAndSalt(password, salt []byte) []byte {
	return pbkdf2.Key(password, salt, 4096, 32, sha1.New)
}

// generateFromPassword returns the key and the salt.
//
// https://github.com/golang/crypto/blob/master/pbkdf2/pbkdf2.go
//
// dk := pbkdf2.Key([]byte("some password"), salt, 4096, 32, sha1.New)
//
func generateFromPassword(password []byte) ([]byte, []byte) {
	salt := Random(saltLen)
	dk := pbkdf2.Key(password, salt, 4096, 32, sha1.New)
	return dk, salt
}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func Random(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return b
}

func RandomAlphanumeric(size int) string {
	dictionary := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	l := byte(len(dictionary))
	b := make([]byte, size)
	rand.Read(b)
	for k, v := range b {
		b[k] = dictionary[v%l]
	}
	return string(b)
}
