package grpccreds

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"google.golang.org/grpc/credentials"
	"os"
)

// ServerCredentials создаёт TLS-credentials для gRPC-сервера
// из файлов сертификата и приватного ключа.
func ServerCredentials(certFile, keyFile string) (credentials.TransportCredentials, error) {
	if certFile == "" || keyFile == "" {
		return nil, fmt.Errorf("grpc tls: cert and key paths are required")
	}

	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("grpc server tls: %w", err)
	}

	return creds, nil
}

// ClientCredentials создаёт TLS-credentials для gRPC-клиента.
// certFile — путь к сертификату сервера (для самоподписанного — тот же grpc.pem).
func ClientCredentials(certFile string) (credentials.TransportCredentials, error) {
	if certFile == "" {
		return nil, fmt.Errorf("grpc tls: cert path is required")
	}

	caCert, err := os.ReadFile(certFile)
	if err != nil {
		return nil, fmt.Errorf("read grpc cert: %w", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("grpc tls: failed to parse cert %s", certFile)
	}

	tlsConfig := &tls.Config{
		RootCAs:    certPool,
		MinVersion: tls.VersionTLS12,
	}

	return credentials.NewTLS(tlsConfig), nil
}
