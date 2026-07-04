package domain

import "github.com/google/uuid"

const (
	EventPermissionAssigned = "PermissionAssigned"
	EventPermissionRevoked  = "PermissionRevoked"
	EventRoleCreated        = "RoleCreated"
)

type PermissionAssignedPayload struct {
	RoleID       uuid.UUID
	RoleName     string
	PermissionID uuid.UUID
	Permission   string
	GrantedBy    uuid.UUID
}

type PermissionRevokedPayload struct {
	RoleID       uuid.UUID
	RoleName     string
	PermissionID uuid.UUID
	Permission   string
}

type RoleCreatedPayload struct {
	RoleID    uuid.UUID
	RoleName  string
	CreatedBy uuid.UUID
}
