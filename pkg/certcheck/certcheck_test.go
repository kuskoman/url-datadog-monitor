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
	cert, err := CheckCertificate("https://example.com", false)

	if err != nil {
		t.Skipf("Skipping real certificate test due to connection error: %v", err)
		return
	}

	if cert.Subject == "" {
		t.Errorf("Expected non-empty subject")
	}

	if cert.Issuer == "" {
		t.Errorf("Expected non-empty issuer")
	}

	if time.Now().After(cert.NotAfter) {
		t.Errorf("Certificate appears to be expired: NotAfter = %v", cert.NotAfter)
	}

	if time.Now().Before(cert.NotBefore) {
		t.Errorf("Certificate is not yet valid: NotBefore = %v", cert.NotBefore)
	}

	if !cert.IsValid {
		t.Errorf("Expected certificate to be valid, got invalid with error: %v", cert.Error)
	}
}
