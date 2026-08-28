// Copyright 2022 Douyu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xgin

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func Test_Server(t *testing.T) {
	c := DefaultConfig()
	c.Port = 0
	s := c.MustBuild()
	stoped := make(chan struct{}, 1)
	go func() {
		time.AfterFunc(time.Second, func() {
			stoped <- struct{}{}
		})
		assert.True(t, s.Healthz())
		assert.NotNil(t, s.Info())
		s.Serve()
	}()
	<-stoped
}

// generateSelfSignedCert 在临时目录生成自签证书，返回 certFile/keyFile 路径
func generateSelfSignedCert(t *testing.T) (string, string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)

	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &key.PublicKey, key)
	assert.NoError(t, err)

	certFile := filepath.Join(t.TempDir(), "cert.pem")
	keyFile := filepath.Join(t.TempDir(), "key.pem")

	certOut, err := os.Create(certFile)
	assert.NoError(t, err)
	assert.NoError(t, pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: der}))
	assert.NoError(t, certOut.Close())

	keyDER, err := x509.MarshalECPrivateKey(key)
	assert.NoError(t, err)
	keyOut, err := os.Create(keyFile)
	assert.NoError(t, err)
	assert.NoError(t, pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}))
	assert.NoError(t, keyOut.Close())

	return certFile, keyFile
}

func Test_Server_TLS(t *testing.T) {
	certFile, keyFile := generateSelfSignedCert(t)

	c := DefaultConfig()
	c.Port = 0
	c.CertFile = certFile
	c.KeyFile = keyFile
	c.ShutdownTimeout = 3 * time.Second

	s := c.MustBuild()
	assert.Equal(t, "https", s.Info().Scheme)

	s.GET("/health", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})

	errCh := make(chan error, 1)
	go func() { errCh <- s.Serve() }()

	assert.Eventually(t, func() bool {
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: 3 * time.Second,
		}
		resp, err := client.Get("https://" + s.listener.Addr().String() + "/health")
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, 5*time.Second, 100*time.Millisecond, "TLS server should serve request")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	assert.NoError(t, s.GracefulStop(ctx))

	select {
	case err := <-errCh:
		assert.NoError(t, err)
	case <-time.After(3 * time.Second):
		t.Fatal("Serve() did not return after GracefulStop")
	}
}
