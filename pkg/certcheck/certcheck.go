package certcheck

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"strings"
	"time"
)

const (
	SchemeHTTPS    = "https"
	DefaultHTTPSPort = "443"
)

// CertificateDetails holds information about an SSL certificate
type CertificateDetails struct {
	Subject      string
	Issuer       string
	SerialNumber string
	NotBefore    time.Time
	NotAfter     time.Time
	DNSNames     []string
	IsValid      bool
	Error        error
}

// CheckCertificate retrieves and validates the SSL certificate for a given URL
func CheckCertificate(rawURL string, verifyChain bool) (*CertificateDetails, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	host := parsedURL.Host
	var port string

	if strings.Contains(host, ":") {
		var portErr error
		host, port, portErr = net.SplitHostPort(parsedURL.Host)
		if portErr != nil {
			return nil, fmt.Errorf("failed to split host and port: %w", portErr)
		}
	} else if parsedURL.Scheme == SchemeHTTPS {
		port = DefaultHTTPSPort
	} else {
		return nil, fmt.Errorf("not an HTTPS URL, cannot check certificate")
	}

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%s", host, port), &tls.Config{
		InsecureSkipVerify: !verifyChain,
	})
	if err != nil {
		return nil, fmt.Errorf("TLS connection failed: %w", err)
	}
	defer conn.Close()

	cert := conn.ConnectionState().PeerCertificates[0]
	details := &CertificateDetails{
		Subject:      cert.Subject.CommonName,
		Issuer:       cert.Issuer.CommonName,
		SerialNumber: cert.SerialNumber.String(),
		NotBefore:    cert.NotBefore,
		NotAfter:     cert.NotAfter,
		DNSNames:     cert.DNSNames,
		IsValid:      true,
	}

	err = cert.VerifyHostname(host)
	if err != nil {
		details.IsValid = false
		details.Error = fmt.Errorf("hostname verification failed: %w", err)
		return details, details.Error
	}

	if verifyChain {
		roots, err := x509.SystemCertPool()
		if err != nil {
			details.IsValid = false
			details.Error = fmt.Errorf("failed to load system cert pool: %w", err)
			return details, details.Error
		}

		opts := x509.VerifyOptions{
			DNSName: host,
			Roots:   roots,
		}
		_, err = cert.Verify(opts)
		if err != nil {
			details.IsValid = false
			details.Error = fmt.Errorf("certificate chain verification failed: %w", err)
			return details, details.Error
		}
	}

	now := time.Now()
	if now.After(cert.NotAfter) || now.Before(cert.NotBefore) {
		details.IsValid = false
		details.Error = fmt.Errorf("certificate is not valid at this time (valid from %s to %s)",
			cert.NotBefore.Format(time.RFC3339),
			cert.NotAfter.Format(time.RFC3339))
		return details, details.Error
	}

	return details, nil
}

// LogCertificateInfo logs information about an SSL certificate
func LogCertificateInfo(logger *slog.Logger, url string, cert *CertificateDetails) {
	expiryDays := time.Until(cert.NotAfter).Hours() / 24
	
	if cert.IsValid {
		logger.Info("Certificate is valid",
			slog.String("url", url),
			slog.String("subject", cert.Subject),
			slog.String("issuer", cert.Issuer),
			slog.Float64("days_until_expiry", expiryDays),
			slog.Time("expires", cert.NotAfter),
			slog.Time("valid_from", cert.NotBefore))
	} else {
		logger.Warn("Certificate has issues",
			slog.String("url", url),
			slog.String("subject", cert.Subject),
			slog.String("issuer", cert.Issuer),
			slog.Float64("days_until_expiry", expiryDays),
			slog.Time("expires", cert.NotAfter),
			slog.Time("valid_from", cert.NotBefore),
			slog.Any("error", cert.Error))
	}
}