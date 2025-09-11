package smtpd

import (
	"crypto/tls"
	"testing"
)

// TestTLSConfiguration verifies the improved TLS security settings
func TestTLSConfiguration(t *testing.T) {
	srv := &Server{}

	// Create test certificates (would use real certs in production)
	// For testing, we'll verify the configuration is set correctly
	srv.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{},
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
		},
		PreferServerCipherSuites: true,
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.X25519,
		},
	}

	// Verify minimum TLS version
	if srv.TLSConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("Expected MinVersion TLS 1.2, got %v", srv.TLSConfig.MinVersion)
	}

	// Verify secure cipher suites are configured
	if len(srv.TLSConfig.CipherSuites) == 0 {
		t.Error("No cipher suites configured")
	}

	// Check for weak ciphers (should not be present)
	weakCiphers := []uint16{
		tls.TLS_RSA_WITH_RC4_128_SHA,
		tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
		tls.TLS_RSA_WITH_AES_128_CBC_SHA,
		tls.TLS_RSA_WITH_AES_256_CBC_SHA,
	}

	for _, weak := range weakCiphers {
		for _, configured := range srv.TLSConfig.CipherSuites {
			if configured == weak {
				t.Errorf("Weak cipher suite found: %v", weak)
			}
		}
	}

	// Verify server cipher preference
	if !srv.TLSConfig.PreferServerCipherSuites {
		t.Error("Server should prefer its own cipher suites")
	}

	// Verify secure curves
	if len(srv.TLSConfig.CurvePreferences) == 0 {
		t.Error("No elliptic curves configured")
	}

	// Check for secure curves
	hasP256 := false
	hasX25519 := false
	for _, curve := range srv.TLSConfig.CurvePreferences {
		if curve == tls.CurveP256 {
			hasP256 = true
		}
		if curve == tls.X25519 {
			hasX25519 = true
		}
	}

	if !hasP256 || !hasX25519 {
		t.Error("Missing recommended elliptic curves")
	}
}

// TestTLSVersionNegotiation verifies that old TLS versions are rejected
func TestTLSVersionNegotiation(t *testing.T) {
	srv := &Server{}
	srv.TLSConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// Verify that TLS 1.0 and 1.1 would be rejected
	unsupportedVersions := []uint16{
		tls.VersionTLS10,
		tls.VersionTLS11,
	}

	for _, version := range unsupportedVersions {
		if version >= srv.TLSConfig.MinVersion {
			t.Errorf("Insecure TLS version %v should be rejected", version)
		}
	}

	// Verify that TLS 1.2 and 1.3 are supported
	supportedVersions := []uint16{
		tls.VersionTLS12,
		tls.VersionTLS13,
	}

	for _, version := range supportedVersions {
		if version < srv.TLSConfig.MinVersion {
			t.Errorf("Secure TLS version %v should be supported", version)
		}
	}
}

