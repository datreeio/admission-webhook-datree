package licensemanagerclient

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// LicenseCredentialsProvider is the custom credential provider that can retrieve valid temporary aws credentials
type LicenseCredentialsProvider struct {
	fallBackProvider   aws.CredentialsProvider
	mux                sync.RWMutex
	licenseCredentials aws.Credentials
	err                error
}

// NewLicenseCredentialsProvider method will create a LicenseCredentialProvider Object which contains valid temporary aws credentials
func NewLicenseCredentialsProvider() (*LicenseCredentialsProvider, error) {
	licenseCredentialProvider := &LicenseCredentialsProvider{}
	fallBackProvider, err := createCredentialProvider()
	if err != nil {
		return licenseCredentialProvider, fmt.Errorf("failed to create LicenseCredentialsProvider, %w", err)
	}
	licenseCredentialProvider.fallBackProvider = fallBackProvider
	return licenseCredentialProvider, nil
}

// Retrieve method will retrieve temporary aws credentials from the credential provider
func (l *LicenseCredentialsProvider) Retrieve(ctx context.Context) (aws.Credentials, error) {
	l.mux.RLock()
	defer l.mux.RUnlock()
	l.licenseCredentials, l.err = l.fallBackProvider.Retrieve(ctx)
	return l.licenseCredentials, l.err
}

func createCredentialProvider() (aws.CredentialsProvider, error) {
	// LoadDefaultConfig will examine all "default" credential providers
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create FallBackProvider, %w", err)
	}

	if cfg.Credentials != nil {
		if _, err := cfg.Credentials.Retrieve(ctx); err != nil {
			// If the "default" credentials provider cannot retrieve credentials, enable fallback to customCredentialsProvider.
			return nil, err
		}
	}

	return cfg.Credentials, nil
}
