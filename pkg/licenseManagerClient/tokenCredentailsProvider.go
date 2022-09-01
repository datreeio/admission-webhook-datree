package licensemanagerclient

// import (
// 	"context"
// 	"fmt"
// 	"io/ioutil"
// 	"os"
// 	"sync"
// 	"time"

// 	"github.com/aws/aws-sdk-go-v2/aws"
// 	"github.com/aws/aws-sdk-go-v2/config"
// 	"github.com/aws/aws-sdk-go-v2/service/sts"
// )

// const awsRefreshTokenFilePathEnvVar = "AWS_LICENSE_ACCESS_FILE"

// // licenseManagerTokenCredentialsProvider defines and contains StsAssumeRoleWithWebIdentityProvider
// type licenseManagerTokenCredentialsProvider struct {
// 	stsCredentialProvider *stsAssumeRoleWithWebIdentityProvider
// 	mux                   sync.RWMutex
// 	licenseCredentials    aws.Credentials
// 	err                   error
// }

// // Retrieve method will retrieve credentials from credential provider.
// // Make this method public to make this provider satisfies CredentialProvider interface
// func (a *licenseManagerTokenCredentialsProvider) Retrieve(ctx context.Context) (aws.Credentials, error) {
// 	a.mux.RLock()
// 	defer a.mux.RUnlock()
// 	a.licenseCredentials, a.err = a.stsCredentialProvider.Retrieve(ctx)
// 	return a.licenseCredentials, a.err
// }

// // newLicenseManagerTokenCredentialsProvider will create and return a LicenseManagerTokenCredentialsProvider Object which wraps up stsAssumeRoleWithWebIdentityProvider
// func newLicenseManagerTokenCredentialsProvider() (*licenseManagerTokenCredentialsProvider, error) {
// 	// 1. Retrieve variables From yaml environment
// 	envConfig, err := config.NewEnvConfig()
// 	if err != nil {
// 		return &licenseManagerTokenCredentialsProvider{}, fmt.Errorf("failed to create LicenseManagerTokenCredentialsProvider, %w", err)
// 	}
// 	roleArn := envConfig.RoleARN
// 	var roleSessionName string
// 	if envConfig.RoleSessionName == "" {
// 		roleSessionName = fmt.Sprintf("aws-sdk-go-v2-%v", time.Now().UnixNano())
// 	} else {
// 		roleSessionName = envConfig.RoleSessionName
// 	}
// 	tokenFilePath := os.Getenv(awsRefreshTokenFilePathEnvVar)
// 	b, err := ioutil.ReadFile(tokenFilePath)
// 	if err != nil {
// 		return &licenseManagerTokenCredentialsProvider{}, fmt.Errorf("failed to create LicenseManagerTokenCredentialsProvider, %w", err)
// 	}
// 	refreshToken := aws.String(string(b))

// 	// 2. Create stsClient
// 	cfg, err := config.LoadDefaultConfig(context.TODO())
// 	if err != nil {
// 		return &licenseManagerTokenCredentialsProvider{}, fmt.Errorf("failed to create LicenseManagerTokenCredentialsProvider, %w", err)
// 	}
// 	stsClient := sts.NewFromConfig(cfg, func(o *sts.Options) {
// 		o.Region = configureStsClientRegion(cfg.Region)
// 		o.Credentials = aws.AnonymousCredentials{}
// 	})

// 	// 3. Configure StsAssumeRoleWithWebIdentityProvider
// 	stsCredentialProvider := newStsAssumeRoleWithWebIdentityProvider(stsClient, roleArn, roleSessionName, refreshToken)

// 	// 4. Build and return
// 	return &licenseManagerTokenCredentialsProvider{
// 		stsCredentialProvider: stsCredentialProvider,
// 	}, nil
// }

// func configureStsClientRegion(configRegion string) string {
// 	defaultRegion := "us-east-1"
// 	if configRegion == "" {
// 		return defaultRegion
// 	} else {
// 		return configRegion
// 	}
// }
