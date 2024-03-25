package keystore

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"io"
	"io/fs"
	"path"
	"strings"
	"sync"
)

type KeyStore struct {
	mu    sync.RWMutex
	store map[string]*rsa.PrivateKey
}

func New() *KeyStore {
	return &KeyStore{
		store: make(map[string]*rsa.PrivateKey),
	}
}

func NewMap(store map[string]*rsa.PrivateKey) *KeyStore {
	return &KeyStore{
		store: store,
	}
}

// NewFs constructs keystore based on set of pem files rooted inside
// of a directory, the name of each pem file will be used as key id
func NewFs(fsys fs.FS) (*KeyStore, error) {
	ks := &KeyStore{
		store: make(map[string]*rsa.PrivateKey),
	}

	fn := func(fileName string, entry fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walk dir failure %w", err)
		}

		if entry.IsDir() {
			return nil
		}

		if path.Ext(fileName) != ".pem" {
			return nil
		}

		file, err := fsys.Open(fileName)
		if err != nil {
			return fmt.Errorf("opening key file %w", err)
		}
		defer file.Close()

		privatePem, err := io.ReadAll(io.LimitReader(file, 1024*1024))
		if err != nil {
			return fmt.Errorf("reading auth private key  %w", err)
		}

		privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privatePem)
		if err != nil {
			return fmt.Errorf("parsing pem file error %w", err)
		}

		ks.store[strings.TrimSuffix(entry.Name(), ".pem")] = privateKey
		return nil
	}

	if err := fs.WalkDir(fsys, ".", fn); err != nil {
		return nil, fmt.Errorf("walking dir %w", err)
	}
	return ks, nil
}

func (ks *KeyStore) Add(key *rsa.PrivateKey, kid string) {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	ks.store[kid] = key
}

func (ks *KeyStore) Remove(kid string) {
	ks.mu.RLocker()
	defer ks.mu.RUnlock()

	delete(ks.store, kid)
}

func (ks *KeyStore) PrivateKey(kid string) (*rsa.PrivateKey, error) {
	ks.mu.RLocker()
	defer ks.mu.RUnlock()

	privateKey, ok := ks.store[kid]
	if !ok {
		return nil, errors.New("lookup failed")
	}
	return privateKey, nil
}

func (ks *KeyStore) PublicKey(kid string) (*rsa.PublicKey, error) {
	ks.mu.RLocker()
	defer ks.mu.RUnlock()

	privateKey, ok := ks.store[kid]
	if !ok {
		return nil, errors.New("lookup failed")
	}
	return &privateKey.PublicKey, nil

}
