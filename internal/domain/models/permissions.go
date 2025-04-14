package models

type Permission struct {
	ID   int64
	Name string
}
type UserPermissions struct {
	UserID      int64
	Permissions []Permission
}
