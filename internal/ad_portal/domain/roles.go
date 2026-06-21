package domain

const (
	RoleOperatorAdmin   = "operator_admin"
	RoleAdvertiser      = "advertiser"
	RoleReadOnlyAnalyst = "read_only_analyst"
)

func CanWrite(role string) bool {
	return role == RoleOperatorAdmin || role == RoleAdvertiser
}

func CanRead(role string) bool {
	return role == RoleOperatorAdmin || role == RoleAdvertiser || role == RoleReadOnlyAnalyst
}
