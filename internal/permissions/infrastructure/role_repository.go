package infrastructure

import (
	"context"
	"time"

	"skykin-platform/internal/permissions/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleModel struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string    `gorm:"type:varchar(100)"`
	Description string    `gorm:"type:text"`
	IsSystem    bool      `gorm:"default:false"`
	CreatedAt   time.Time
}

func (RoleModel) TableName() string { return "rbac_roles" }

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

var _ domain.RoleRepository = (*RoleRepository)(nil)

func (r *RoleRepository) FindAll(ctx context.Context) ([]*domain.Role, error) {
	var rows []RoleModel
	if err := r.db.WithContext(ctx).Order("name asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.Role, len(rows))
	for i := range rows {
		role, err := r.roleWithPermissions(ctx, &rows[i])
		if err != nil {
			return nil, err
		}
		out[i] = role
	}
	return out, nil
}

func (r *RoleRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	var row RoleModel
	if err := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&row).Error; err != nil {
		return nil, err
	}
	return r.roleWithPermissions(ctx, &row)
}

func (r *RoleRepository) FindByName(ctx context.Context, name string) (*domain.Role, error) {
	var row RoleModel
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&row).Error; err != nil {
		return nil, err
	}
	return r.roleWithPermissions(ctx, &row)
}

func (r *RoleRepository) Create(ctx context.Context, role *domain.Role) error {
	row := &RoleModel{
		Name: role.Name, Description: role.Description, IsSystem: false,
	}
	if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
		return err
	}
	role.ID, _ = uuid.Parse(row.ID)
	role.IsSystem = false
	role.CreatedAt = row.CreatedAt
	return nil
}

func (r *RoleRepository) GetPermissionNames(ctx context.Context, roleName string) ([]string, error) {
	var names []string
	err := r.db.WithContext(ctx).Raw(`
		SELECT p.name
		FROM rbac_permissions p
		INNER JOIN rbac_role_permissions rp ON rp.permission_id = p.id
		INNER JOIN rbac_roles r ON r.id = rp.role_id
		WHERE r.name = ?
	`, roleName).Scan(&names).Error
	return names, err
}

func (r *RoleRepository) AssignPermission(
	ctx context.Context,
	roleID uuid.UUID,
	permissionID uuid.UUID,
	grantedBy uuid.UUID,
) error {
	return r.db.WithContext(ctx).Exec(`
		INSERT INTO rbac_role_permissions (role_id, permission_id, granted_by)
		VALUES (?, ?, ?)
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`, roleID, permissionID, grantedBy).Error
}

func (r *RoleRepository) RevokePermission(ctx context.Context, roleID, permissionID uuid.UUID) error {
	return r.db.WithContext(ctx).Exec(`
		DELETE FROM rbac_role_permissions
		WHERE role_id = ? AND permission_id = ?
	`, roleID, permissionID).Error
}

func (r *RoleRepository) roleWithPermissions(ctx context.Context, row *RoleModel) (*domain.Role, error) {
	perms, err := r.loadPermissions(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	id, _ := uuid.Parse(row.ID)
	return &domain.Role{
		ID: id, Name: row.Name, Description: row.Description,
		IsSystem: row.IsSystem, Permissions: perms, CreatedAt: row.CreatedAt,
	}, nil
}

func (r *RoleRepository) loadPermissions(ctx context.Context, roleID string) ([]*domain.Permission, error) {
	type permRow struct {
		ID          string
		Name        string
		Resource    string
		Action      string
		Description string
		CreatedAt   time.Time
	}
	var rows []permRow
	err := r.db.WithContext(ctx).Raw(`
		SELECT p.id, p.name, p.resource, p.action, p.description, p.created_at
		FROM rbac_permissions p
		INNER JOIN rbac_role_permissions rp ON rp.permission_id = p.id
		WHERE rp.role_id = ?
		ORDER BY p.resource, p.action
	`, roleID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Permission, len(rows))
	for i := range rows {
		pid, _ := uuid.Parse(rows[i].ID)
		out[i] = &domain.Permission{
			ID: pid, Name: rows[i].Name, Resource: rows[i].Resource,
			Action: rows[i].Action, Description: rows[i].Description,
			CreatedAt: rows[i].CreatedAt,
		}
	}
	return out, nil
}
