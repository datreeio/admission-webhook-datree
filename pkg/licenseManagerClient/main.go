package licensemanagerclient

import (
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/licensemanager"
)

func PollCheckoutLicense(token string) {
	l := NewLicenseManagerClient(nil)

	isPolling := false
	go func() {
		for {
			if !isPolling {
				isPolling = true
				_, err := l.checkoutLicense(token)
				if err != nil {
					_, err := l.checkoutLicense(token)
					if err != nil {
						log.Printf("failed to checkout license, %v", err)
						// wait 30 seconds before polling again
						time.Sleep(30 * time.Second)
						isPolling = false
					}
				}
				isPolling = false
			} else {
				time.Sleep(30 * time.Minute)
			}
		}
	}()
}

type LicenseManager struct {
	client *licensemanager.LicenseManager
}

func NewLicenseManagerClient(p client.ConfigProvider, cfgs ...*aws.Config) *LicenseManager {
	awsClient := licensemanager.New(p, cfgs...)
	return &LicenseManager{
		client: awsClient,
	}
}

func (l *LicenseManager) checkoutLicense(token string) (*licensemanager.CheckoutLicenseOutput, error) {
	awsMarketplaceProductID := "ad0ee0c8-f50f-464a-9bc4-d6270592dd36"
	awsMarketplaceFingerprint := "aws:294406891311:AWS/Marketplace:issuer-fingerprint"

	res, err := l.client.CheckoutLicense(&licensemanager.CheckoutLicenseInput{
		ClientToken:    aws.String(token),
		CheckoutType:   aws.String("PROVISIONAL"),
		Entitlements:   []*licensemanager.EntitlementData{},
		ProductSKU:     aws.String(awsMarketplaceProductID),
		KeyFingerprint: aws.String(awsMarketplaceFingerprint),
	})

	if err != nil {
		return nil, err
	}

	return res, nil
}
