/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package certs

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"
)

var (
	ellipticCurve = elliptic.P256()
	bigOne        = big.NewInt(1)
)

// CertPair is a private key and certificate for use for client auth, as a CA, or serving.
type CertPair struct {
	Key  crypto.Signer
	Cert *x509.Certificate
}

// CertBytes returns the PEM-encoded version of the certificate for this pair.
func (k CertPair) CertBytes() []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: k.Cert.Raw,
	})
}

// AsBytes encodes keypair in the appropriate formats for on-disk storage (PEM and
// PKCS8, respectively).
func (k CertPair) AsBytes() (cert []byte, key []byte, err error) {
	cert = k.CertBytes()

	rawKeyData, err := x509.MarshalPKCS8PrivateKey(k.Key)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to encode private key: %w", err)
	}

	key = pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: rawKeyData,
	})

	return cert, key, nil
}

// TinyCA supports signing client-certs and can be used as an auth mechanism with envtest.
type TinyCA struct {
	CA      CertPair
	orgName string

	nextSerial *big.Int
}

// LoadTinyCA loads an existing CA from cert and key files on disk.
func LoadTinyCA(certPath, keyPath string) (*TinyCA, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read CA cert: %w", err)
	}
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read CA key: %w", err)
	}

	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return nil, fmt.Errorf("unable to decode CA cert PEM")
	}
	caCert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("unable to parse CA cert: %w", err)
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, fmt.Errorf("unable to decode CA key PEM")
	}
	caKey, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		rsaKey, rsaErr := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
		if rsaErr == nil {
			caKey = rsaKey
		} else {
			ecKey, ecErr := x509.ParseECPrivateKey(keyBlock.Bytes)
			if ecErr != nil {
				return nil, fmt.Errorf("unable to parse CA key: %w", err)
			}
			caKey = ecKey
		}
	}

	signer, ok := caKey.(crypto.Signer)
	if !ok {
		return nil, fmt.Errorf("CA key does not implement crypto.Signer")
	}

	return &TinyCA{
		CA:         CertPair{Key: signer, Cert: caCert},
		orgName:    "envtest",
		nextSerial: big.NewInt(1),
	}, nil
}

// ClientInfo describes some Kubernetes user for the purposes of creating
// client certificates.
type ClientInfo struct {
	Name   string
	Groups []string
}

// NewClientCert produces a new CertPair suitable for use with Kubernetes
// client cert auth with an API server validating based on this CA.
func (c *TinyCA) NewClientCert(user ClientInfo) (CertPair, error) {
	now := time.Now()

	key, err := ecdsa.GenerateKey(ellipticCurve, crand.Reader)
	if err != nil {
		return CertPair{}, fmt.Errorf("unable to create private key: %w", err)
	}

	serial := new(big.Int).Set(c.nextSerial)
	c.nextSerial.Add(c.nextSerial, bigOne)

	template := x509.Certificate{
		Subject:      pkix.Name{CommonName: user.Name, Organization: user.Groups},
		SerialNumber: serial,
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		NotBefore:    now.UTC(),
		NotAfter:     now.Add(168 * time.Hour).UTC(),
	}

	certRaw, err := x509.CreateCertificate(crand.Reader, &template, c.CA.Cert, key.Public(), c.CA.Key)
	if err != nil {
		return CertPair{}, fmt.Errorf("unable to create certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(certRaw)
	if err != nil {
		return CertPair{}, fmt.Errorf("generated invalid certificate, could not parse: %w", err)
	}

	return CertPair{
		Key:  key,
		Cert: cert,
	}, nil
}
