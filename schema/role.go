package schema

const (
	PermissionCreateUser = 1 << iota
	PermissionModifySelfTasks
	PermissionModifyAllUsers
	PermissionModifyAllUsersRestricted
	PermissionViewAllTasks
	PermissionModifyAllTasks
)

var (
	RoleAnon    = PermissionCreateUser
	RoleUser    = PermissionModifySelfTasks
	RoleManager = RoleUser | PermissionModifyAllUsersRestricted | PermissionViewAllTasks
	RoleAdmin   = RoleManager | PermissionModifyAllUsers | PermissionModifyAllTasks
)

func RoleHasPermission(role, perm int) bool {
	return (role & perm) > 0
}
