// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package gqlschema

type CertificateSigningRequestInfo struct {
	Subject      string `json:"subject"`
	KeyAlgorithm string `json:"keyAlgorithm"`
}

type CertificationResult struct {
	CertificateChain  string `json:"certificateChain"`
	CaCertificate     string `json:"caCertificate"`
	ClientCertificate string `json:"clientCertificate"`
}

type Configuration struct {
	Token                         *Token                         `json:"token"`
	CertificateSigningRequestInfo *CertificateSigningRequestInfo `json:"certificateSigningRequestInfo"`
	ManagementPlaneInfo           *ManagementPlaneInfo           `json:"managementPlaneInfo"`
}

type ManagementPlaneInfo struct {
	DirectorURL string `json:"directorURL"`
}

type Token struct {
	Token string `json:"token"`
}
