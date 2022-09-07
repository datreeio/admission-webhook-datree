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

type LicenseManager struct {
	client                    *licensemanager.LicenseManager
	awsMarketplaceProductID   string
	awsMarketplaceFingerprint string
}

func NewLicenseManagerClient() *LicenseManager {
	clientSession := session.Must(session.NewSession())
	awsClient := licensemanager.New(clientSession, aws.NewConfig().WithRegion(os.Getenv(enums.AWSMarketplaceRegion)))
	return &LicenseManager{
		client:                    awsClient,
		awsMarketplaceProductID:   os.Getenv(enums.AWSMarketplaceProductID),
		awsMarketplaceFingerprint: os.Getenv(enums.AWSMarketplaceKeyFingerprint),
	}
}

// Checkout the account license according to quantity of units the account consumes.
// If everything goes well, the license will be checked out, otherwise an error will returned.
func (l *LicenseManager) CheckoutLicense(consumedUnitsCount int) error {
	_, err := l.client.CheckoutLicense(&licensemanager.CheckoutLicenseInput{
		ClientToken: aws.String(uuid.New().String()),
		// "PROVISIONAL" checkout type enables to temporarily draw a unit and return it back to the license pool when the application is stopped.
		CheckoutType: aws.String("PROVISIONAL"),
		Entitlements: []*licensemanager.EntitlementData{
			{
				// The entitilement name is the contract API name defined in the product.
				// The contract API name is defined in the product "load form" in the AWS Marketplace management protal
				Name:  aws.String("Datree"),
				Value: aws.String(fmt.Sprint(consumedUnitsCount)),
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
