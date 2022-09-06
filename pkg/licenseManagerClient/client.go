package licensemanagerclient

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/licensemanager"
	"github.com/datreeio/admission-webhook-datree/pkg/enums"
	"github.com/google/uuid"
)

const awsMarketplaceIssuer = "aws:294406891311:AWS/Marketplace:issuer-fingerprint"

type LicenseManager struct {
	client                    *licensemanager.LicenseManager
	awsMarketplaceProductID   string
	awsMarketplaceFingerprint string
}

func NewLicenseManagerClient() *LicenseManager {
	clientSession := session.Must(session.NewSession())
	awsClient := licensemanager.New(clientSession, aws.NewConfig().WithRegion("us-east-1"))
	return &LicenseManager{
		client:                    awsClient,
		awsMarketplaceProductID:   os.Getenv(enums.AWSMarketplaceProductSKU),
		awsMarketplaceFingerprint: awsMarketplaceIssuer,
	}
}

// Checkout the account license according to number of nodes, if everything goes well, the license will be checked out,
// otherwise an error will returned.
func (l *LicenseManager) CheckoutLicense(entititlementValue int) error {
	_, err := l.client.CheckoutLicense(&licensemanager.CheckoutLicenseInput{
		ClientToken:  aws.String(uuid.New().String()),
		CheckoutType: aws.String("PROVISIONAL"),
		Entitlements: []*licensemanager.EntitlementData{
			{
				Name:  aws.String("Datree"),
				Value: aws.String(fmt.Sprint(entititlementValue)),
				Unit:  aws.String("Count"),
			},
		},
		ProductSKU:     aws.String(l.awsMarketplaceProductID),
		KeyFingerprint: aws.String(l.awsMarketplaceFingerprint),
	})
	if err != nil {
		return err
	}

	return nil
}
