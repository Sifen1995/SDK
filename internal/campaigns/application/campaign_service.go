package application

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"skykin-platform/internal/advertisers/domain"
	"skykin-platform/internal/campaigns/infrastructure"
	"skykin-platform/internal/campaigns/model"

	"gorm.io/datatypes"
)

type CampaignService struct {
	repo *infrastructure.Repository
}

func NewCampaignService(repo *infrastructure.Repository) *CampaignService {
	return &CampaignService{repo: repo}
}

type CreateCampaignInput struct {
	Name           string
	TargetIntent   string
	CreativeFormat string
	Title          string
	BodyText       string
	ImageURL       string
	CanvasJSON     map[string]any
	DailyBudgetCap float64
	TotalBudgetCap float64
}

func (s *CampaignService) Create(ctx context.Context, advertiserID, role string, in CreateCampaignInput) (*model.Campaign, error) {
	if !domain.CanWrite(role) {
		return nil, errors.New("forbidden")
	}
	if strings.TrimSpace(advertiserID) == "" {
		return nil, errors.New("account has no advertiser company; operator admins cannot create campaigns directly")
	}
	format, err := NormalizeCreativeFormat(in.CreativeFormat)
	if err != nil {
		return nil, err
	}
	canvas := datatypes.JSON([]byte("{}"))
	if in.CanvasJSON != nil {
		raw, _ := json.Marshal(in.CanvasJSON)
		canvas = datatypes.JSON(raw)
	}
	c := &model.Campaign{
		AdvertiserID:   advertiserID,
		Name:           in.Name,
		TargetIntent:   in.TargetIntent,
		CreativeFormat: format,
		Title:          in.Title,
		BodyText:       in.BodyText,
		ImageURL:       in.ImageURL,
		CanvasJSON:     canvas,
		DailyBudgetCap: in.DailyBudgetCap,
		TotalBudgetCap: in.TotalBudgetCap,
		IsActive:       false,
	}
	vr := ValidateCampaign(c)
	c.ValidationStatus = vr.Status
	c.ValidationNotes = vr.Notes
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CampaignService) List(ctx context.Context, advertiserID, role string) ([]model.Campaign, error) {
	if !domain.CanRead(role) {
		return nil, errors.New("forbidden")
	}
	if role == domain.RoleOperatorAdmin {
		return s.repo.ListAll(ctx)
	}
	return s.repo.ListByAdvertiser(ctx, advertiserID)
}

func (s *CampaignService) Get(ctx context.Context, advertiserID, role, id string) (*model.Campaign, error) {
	if !domain.CanRead(role) {
		return nil, errors.New("forbidden")
	}
	c, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if role != domain.RoleOperatorAdmin && c.AdvertiserID != advertiserID {
		return nil, errors.New("forbidden")
	}
	return c, nil
}

func (s *CampaignService) Activate(ctx context.Context, advertiserID, role, id string) (*model.Campaign, error) {
	if !domain.CanWrite(role) {
		return nil, errors.New("forbidden")
	}
	c, err := s.Get(ctx, advertiserID, role, id)
	if err != nil {
		return nil, err
	}
	if c.ValidationStatus != "passed" {
		return nil, errors.New("campaign cannot be activated until validation_status is passed")
	}
	c.IsActive = true
	c.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CampaignService) Preview(ctx context.Context, advertiserID, role, id string) (map[string]any, error) {
	c, err := s.Get(ctx, advertiserID, role, id)
	if err != nil {
		return nil, err
	}
	return PreviewCampaign(c), nil
}
