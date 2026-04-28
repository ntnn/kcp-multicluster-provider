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

package certs_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/kcp-dev/multicluster-provider/envtest/internal/certs"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func writeCA(dir string, key any, cert *x509.Certificate) (certPath, keyPath string) {
	certPath = filepath.Join(dir, "ca.crt")
	keyPath = filepath.Join(dir, "ca.key")

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	Expect(os.WriteFile(certPath, certPEM, 0o600)).To(Succeed())

	var keyDER []byte
	var blockType string
	switch k := key.(type) {
	case *rsa.PrivateKey:
		keyDER = x509.MarshalPKCS1PrivateKey(k)
		blockType = "RSA PRIVATE KEY"
	case *ecdsa.PrivateKey:
		var err error
		keyDER, err = x509.MarshalECPrivateKey(k)
		Expect(err).NotTo(HaveOccurred())
		blockType = "EC PRIVATE KEY"
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: keyDER})
	Expect(os.WriteFile(keyPath, keyPEM, 0o600)).To(Succeed())

	return certPath, keyPath
}

func generateSelfSignedCA(key any) *x509.Certificate {
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test-ca"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	var pub any
	switch k := key.(type) {
	case *rsa.PrivateKey:
		pub = &k.PublicKey
	case *ecdsa.PrivateKey:
		pub = &k.PublicKey
	}

	certDER, err := x509.CreateCertificate(crand.Reader, template, template, pub, key)
	Expect(err).NotTo(HaveOccurred())
	cert, err := x509.ParseCertificate(certDER)
	Expect(err).NotTo(HaveOccurred())
	return cert
}

var _ = Describe("TinyCA", func() {
	var tmpDir string

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "tinyca-test-*")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(tmpDir)
	})

	Describe("LoadTinyCA", func() {
		It("should load an RSA (PKCS1) CA key", func() {
			rsaKey, err := rsa.GenerateKey(crand.Reader, 2048)
			Expect(err).NotTo(HaveOccurred())

			cert := generateSelfSignedCA(rsaKey)
			certPath, keyPath := writeCA(tmpDir, rsaKey, cert)

			ca, err := certs.LoadTinyCA(certPath, keyPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(ca).NotTo(BeNil())
		})

		It("should load an EC CA key", func() {
			ecKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
			Expect(err).NotTo(HaveOccurred())

			cert := generateSelfSignedCA(ecKey)
			certPath, keyPath := writeCA(tmpDir, ecKey, cert)

			ca, err := certs.LoadTinyCA(certPath, keyPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(ca).NotTo(BeNil())
		})
	})

	Describe("NewClientCert", func() {
		var ca *certs.TinyCA

		BeforeEach(func() {
			rsaKey, err := rsa.GenerateKey(crand.Reader, 2048)
			Expect(err).NotTo(HaveOccurred())

			cert := generateSelfSignedCA(rsaKey)
			certPath, keyPath := writeCA(tmpDir, rsaKey, cert)

			ca, err = certs.LoadTinyCA(certPath, keyPath)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should produce a valid client certificate", func() {
			pair, err := ca.NewClientCert(certs.ClientInfo{
				Name:   "user",
				Groups: []string{"group1", "group2"},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(pair.Cert.Subject.CommonName).To(Equal("user"))
			Expect(pair.Cert.Subject.Organization).To(ConsistOf("group1", "group2"))
			Expect(pair.Cert.ExtKeyUsage).To(ContainElement(x509.ExtKeyUsageClientAuth))
		})

		It("should be signed by the CA", func() {
			pair, err := ca.NewClientCert(certs.ClientInfo{Name: "user"})
			Expect(err).NotTo(HaveOccurred())
			Expect(pair.Cert.CheckSignatureFrom(ca.CA.Cert)).To(Succeed())
		})

		It("should produce unique serial numbers", func() {
			first, err := ca.NewClientCert(certs.ClientInfo{Name: "a"})
			Expect(err).NotTo(HaveOccurred())
			second, err := ca.NewClientCert(certs.ClientInfo{Name: "b"})
			Expect(err).NotTo(HaveOccurred())
			Expect(first.Cert.SerialNumber.Cmp(second.Cert.SerialNumber)).NotTo(Equal(0))
		})

		It("should serialize via AsBytes", func() {
			pair, err := ca.NewClientCert(certs.ClientInfo{Name: "user"})
			Expect(err).NotTo(HaveOccurred())

			certBytes, keyBytes, err := pair.AsBytes()
			Expect(err).NotTo(HaveOccurred())
			Expect(certBytes).NotTo(BeEmpty())
			Expect(keyBytes).NotTo(BeEmpty())

			block, _ := pem.Decode(keyBytes)
			Expect(block).NotTo(BeNil())
			Expect(block.Type).To(Equal("PRIVATE KEY"))
		})
	})
})
