package infrastructure

import (
	"context"
	"errors"

	"skykin-platform/internal/advertisers/model"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetRoleBySlug(ctx context.Context, slug string) (*model.Role, error) {
	var role model.Role
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *Repository) CreateAdvertiser(ctx context.Context, a *model.Advertiser) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *Repository) CreatePortalUser(ctx context.Context, u *model.PortalUser) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *Repository) GetPortalUserByEmail(ctx context.Context, email string) (*model.PortalUser, error) {
	var u model.PortalUser
	err := r.db.WithContext(ctx).
		Preload("Role").
		Preload("Advertiser").
		Where("email = ?", email).
		First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) GetPortalUserByID(ctx context.Context, id string) (*model.PortalUser, error) {
	var u model.PortalUser
	err := r.db.WithContext(ctx).
		Preload("Role").
		Preload("Advertiser").
		Where("id = ?", id).
		First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) SeedRoles(ctx context.Context) error {
	roles := []model.Role{
		{Slug: "operator_admin", DisplayName: "Operator Admin"},
		{Slug: "advertiser", DisplayName: "Advertiser"},
		{Slug: "read_only_analyst", DisplayName: "Read-Only Analyst"},
	}
	for _, role := range roles {
		if err := r.db.WithContext(ctx).Where("slug = ?", role.Slug).FirstOrCreate(&role).Error; err != nil {
			return err
		}
	}
	return nil
}

// CreateAdvertiserAndPortalUser creates or finds an advertiser by company name
// and creates the given portal user within a single DB transaction. This
// prevents orphaned advertisers when user creation fails and avoids creating
// duplicate advertiser rows for the same company name.
func (r *Repository) CreateAdvertiserAndPortalUser(ctx context.Context, adv *model.Advertiser, u *model.PortalUser) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing model.Advertiser
		if err := tx.Where("company_name = ?", adv.CompanyName).First(&existing).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// create new advertiser
				if err := tx.Create(adv).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			// reuse existing advertiser id
			adv.ID = existing.ID
		}

		// ensure portal user references advertiser id
		u.AdvertiserID = &adv.ID
		if err := tx.Create(u).Error; err != nil {
			return err
		}
		return nil
	})
}
