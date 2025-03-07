package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="URL",type=string,JSONPath=`.spec.url`
// +kubebuilder:printcolumn:name="Interval",type=integer,JSONPath=`.spec.interval`
// +kubebuilder:printcolumn:name="Last Check",type=string,JSONPath=`.status.lastCheckTime`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
// +kubebuilder:validation:XValidation:rule="self.spec.timeout < self.spec.interval",message="Timeout must be less than interval"
// +kubebuilder:validation:XValidation:rule="!has(self.spec.checkCert) || !has(self.spec.verifyCert) || self.spec.url.startsWith('https://')",message="Certificate validation only applies to HTTPS URLs"

// URLMonitor is the Schema for the urlmonitors API
type URLMonitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   URLMonitorSpec   `json:"spec,omitempty"`
	Status URLMonitorStatus `json:"status,omitempty"`
}

// URLMonitorSpec defines the desired state of URLMonitor
type URLMonitorSpec struct {
	// URL to monitor
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^https?://.*`
	URL string `json:"url"`

	// HTTP method to use for the request
	// +kubebuilder:default=GET
	// +kubebuilder:validation:Enum=GET;POST;PUT;DELETE;HEAD;OPTIONS
	Method string `json:"method,omitempty"`

	// Interval between checks in seconds
	// +kubebuilder:default=60
	// +kubebuilder:validation:Minimum=5
	// +kubebuilder:validation:Maximum=3600
	Interval int `json:"interval,omitempty"`

	// Timeout for the HTTP request in seconds
	// +kubebuilder:default=10
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=120
	Timeout int `json:"timeout,omitempty"`

	// Headers to include in the request
	// +optional
	Headers map[string]string `json:"headers,omitempty"`

	// Labels to attach to metrics
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Whether to check SSL certificate (for HTTPS URLs)
	// +optional
	// +kubebuilder:default=true
	CheckCert *bool `json:"checkCert,omitempty"`

	// Whether to verify SSL certificate chain
	// +optional
	// +kubebuilder:default=false
	VerifyCert *bool `json:"verifyCert,omitempty"`
}

// URLMonitorStatus defines the observed state of URLMonitor
type URLMonitorStatus struct {
	// Last time the URL was checked
	LastCheckTime metav1.Time `json:"lastCheckTime,omitempty"`

	// Status of the URL (up or down)
	Status string `json:"status,omitempty"`

	// HTTP status code from the last check
	StatusCode int `json:"statusCode,omitempty"`

	// Response time in milliseconds
	ResponseTime int64 `json:"responseTime,omitempty"`

	// Certificate information (if HTTPS and certificate checking is enabled)
	Certificate *CertificateStatus `json:"certificate,omitempty"`
}

// CertificateStatus contains information about the SSL certificate
type CertificateStatus struct {
	// Whether the certificate is valid
	// +kubebuilder:validation:Required
	Valid bool `json:"valid"`

	// Subject of the certificate
	Subject string `json:"subject,omitempty"`

	// Issuer of the certificate
	Issuer string `json:"issuer,omitempty"`

	// Expiration date of the certificate
	NotAfter metav1.Time `json:"notAfter,omitempty"`

	// Days until the certificate expires (as string to avoid float compatibility issues)
	// Format is a string representation of a float for cross-language compatibility
	DaysUntilExpiry string `json:"daysUntilExpiry,omitempty"`
}

// +kubebuilder:object:root=true

// URLMonitorList contains a list of URLMonitor
type URLMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []URLMonitor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&URLMonitor{}, &URLMonitorList{})
}