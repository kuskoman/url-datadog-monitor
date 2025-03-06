package certcheck

import (
	"testing"
	"time"
)

func TestParseNonHTTPSURL(t *testing.T) {
	_, err := CheckCertificate("http://example.com", false)
	if err == nil {
		t.Errorf("Expected error for non-HTTPS URL, got nil")
	}
}

func TestCheckRealCertificate(t *testing.T) {
	// This is an integration test that tries to check a real certificate
	cert, err := CheckCertificate("https://example.com", false)
	
	// If we can't connect, skip the test
	if err != nil {
		t.Skipf("Skipping real certificate test due to connection error: %v", err)
		return
	}
	
	// Basic checks on the returned certificate
	if cert.Subject == "" {
		t.Errorf("Expected non-empty subject")
	}
	
	if cert.Issuer == "" {
		t.Errorf("Expected non-empty issuer")
	}
	
	// Check that NotAfter is in the future
	if time.Now().After(cert.NotAfter) {
		t.Errorf("Certificate appears to be expired: NotAfter = %v", cert.NotAfter)
	}
	
	// Check that NotBefore is in the past
	if time.Now().Before(cert.NotBefore) {
		t.Errorf("Certificate is not yet valid: NotBefore = %v", cert.NotBefore)
	}
	
	// The certificate should be valid
	if !cert.IsValid {
		t.Errorf("Expected certificate to be valid, got invalid with error: %v", cert.Error)
	}
}