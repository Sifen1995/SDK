package bootstrap

import (
	campaignInfra "skykin-platform/internal/campaigns/infrastructure"
	billingdomain "skykin-platform/internal/billing/domain"

	"gorm.io/gorm"
)

// NewCampaignQuotaReader exposes campaign quota counts to billing without letting
// billing import campaign infrastructure directly.
func NewCampaignQuotaReader(db *gorm.DB) billingdomain.CampaignQuotaReader {
	return campaignInfra.NewRepository(db)
}
