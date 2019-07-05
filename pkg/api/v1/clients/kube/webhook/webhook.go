package webhook

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/webhook/certwatcher"
	"go.uber.org/zap"
)

const (
	certName = "tls.crt"
	keyName  = "tls.key"
)

// DefaultPort is the default port that the webhook server serves.
var DefaultPort = 443

// Server is an admission webhook server that can serve traffic and
// generates related k8s resources for deploying.
type Server struct {
	// Host is the address that the server will listen on.
	// Defaults to "" - all addresses.
	Host string

	// Port is the port number that the server will serve.
	// It will be defaulted to 443 if unspecified.
	Port int

	// CertDir is the directory that contains the server key and certificate.
	// If using FSCertWriter in Provisioner, the server itself will provision the certificate and
	// store it in this directory.
	// If using SecretCertWriter in Provisioner, the server will provision the certificate in a secret,
	// the user is responsible to mount the secret to the this location for the server to consume.
	CertDir string

	// WebhookMux is the multiplexer that handles different webhooks.
	WebhookMux *http.ServeMux

	// webhooks keep track of all registered webhooks for dependency injection,
	// and to provide better panic messages on duplicate webhook registration.
	webhooks map[string]http.Handler

	// defaultingOnce ensures that the default fields are only ever set once.
	defaultingOnce sync.Once
}

// setDefaults does defaulting for the Server.
func (s *Server) setDefaults() {
	s.webhooks = map[string]http.Handler{}
	if s.WebhookMux == nil {
		s.WebhookMux = http.NewServeMux()
	}

	if s.Port <= 0 {
		s.Port = DefaultPort
	}

	if len(s.CertDir) == 0 {
		s.CertDir = path.Join("/tmp", "k8s-webhook-server", "serving-certs")
	}
}

// Register marks the given webhook as being served at the given path.
// It panics if two hooks are registered on the same path.
func (s *Server) Register(path string, hook http.Handler) {
	s.defaultingOnce.Do(s.setDefaults)
	_, found := s.webhooks[path]
	if found {
		panic(fmt.Errorf("can't register duplicate path: %v", path))
	}
	// TODO(directxman12): call setfields if we've already started the server
	s.webhooks[path] = hook
	s.WebhookMux.Handle(path, hook)
}

// Start runs the server.
// It will install the webhook related resources depend on the server configuration.
func (s *Server) Start(ctx context.Context) error {
	s.defaultingOnce.Do(s.setDefaults)

	baseHookLog := contextutils.LoggerFrom(ctx).With(zap.String("webhook", "server"))

	certPath := filepath.Join(s.CertDir, certName)
	keyPath := filepath.Join(s.CertDir, keyName)

	certWatcher, err := certwatcher.New(ctx, certPath, keyPath)
	if err != nil {
		return err
	}

	go func() {
		if err := certWatcher.Start(ctx); err != nil {
			baseHookLog.Error(err, "certificate watcher error")
		}
	}()

	cfg := &tls.Config{
		NextProtos:     []string{"h2"},
		GetCertificate: certWatcher.GetCertificate,
	}

	listener, err := tls.Listen("tcp", net.JoinHostPort(s.Host, strconv.Itoa(int(s.Port))), cfg)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Handler: s.WebhookMux,
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		<-ctx.Done()

		// TODO: use a context with reasonable timeout
		if err := srv.Shutdown(ctx); err != nil {
			// Error from closing listeners, or context timeout
			baseHookLog.Error(err, "error shutting down the HTTP server")
		}
		close(idleConnsClosed)
	}()

	err = srv.Serve(listener)
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	<-idleConnsClosed
	return nil
}
