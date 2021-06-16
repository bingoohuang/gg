package rbac

import (
	"errors"

	"gorm.io/gorm"
)

var (
	ErrRoleNotFound       = errors.New("role not found")
	ErrPermNotFound       = errors.New("permission not found")
	ErrDeleteAssignedPerm = errors.New("cannot delete assigned permission")
)

// UserRole represents the relationship between users and roles
type UserRole struct {
	ID     uint
	UserID uint
	RoleID uint
}

// TableName sets the table name
func (u UserRole) TableName() string { return tablePrefix + "user_roles" }

// Role represents the database model of roles
type Role struct {
	ID   uint
	Name string
}

// TableName sets the table name
func (r Role) TableName() string { return tablePrefix + "roles" }

// RolePerm stores the relationship between roles and permissions
type RolePerm struct {
	ID     uint
	RoleID uint
	PermID uint
}

// TableName sets the table name
func (r RolePerm) TableName() string { return tablePrefix + "role_perms" }

// Perm represents the database model of permissions
type Perm struct {
	ID   uint
	Name string
}

// TableName sets the table name
func (p Perm) TableName() string { return tablePrefix + "perms" }

// Rbac helps deal with permissions
type Rbac struct {
	DB *gorm.DB
}

// Options has the options for initiating the package.
type Options struct {
	DB           *gorm.DB
	TablesPrefix string
}

var tablePrefix string
var rbac *Rbac

// New initiates authority.
func New(opts Options) *Rbac {
	tablePrefix = opts.TablesPrefix
	rbac = &Rbac{DB: opts.DB}
	migrateTables(opts.DB)
	return rbac
}

// Instance returns the initiated instance.
func Instance() *Rbac { return rbac }

// NewRole stores a role in the database it accepts the role name.
func (a *Rbac) NewRole(roleName string) error {
	var dbRole Role
	r := a.DB.Where("name=?", roleName).First(&dbRole)
	if r.Error != nil && errors.Is(r.Error, gorm.ErrRecordNotFound) {
		return a.DB.Create(&Role{Name: roleName}).Error
	}

	return r.Error
}

// NewPerm stores a permission in the database it accepts the permission name.
func (a *Rbac) NewPerm(permName string) error {
	perm := Perm{}
	r := a.DB.Where("name=?", permName).First(&perm)
	if r.Error != nil && errors.Is(r.Error, gorm.ErrRecordNotFound) {
		return a.DB.Create(&Perm{Name: permName}).Error
	}

	return r.Error
}

// AssignPerms assigns a group of permissions to a given role
// it accepts in the first parameter the role name, it returns an error if there is not matching record
// of the role name in the database.
// the second parameter is a slice of strings which represents a group of permissions to be assigned to the role
// if any of these permissions doesn't have a matching record in the database the operations stops, changes reverted
// and error is returned
// in case of success nothing is returned
func (a *Rbac) AssignPerms(roleName string, permNames ...string) error {
	var role Role
	if r := a.DB.Where("name=?", roleName).First(&role); r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return r.Error
	}

	var perms []Perm
	for _, permName := range permNames {
		var perm Perm
		if r := a.DB.Where("name=?", permName).First(&perm); r.Error != nil {
			if errors.Is(r.Error, gorm.ErrRecordNotFound) {
				return ErrPermNotFound
			}
			return r.Error
		}

		perms = append(perms, perm)
	}

	// insert data into RolePermissions table
	for _, perm := range perms {
		// ignore any assigned permission
		var rolePerm RolePerm
		if r := a.DB.Where("role_id=?", role.ID).Where("perm_id =?", perm.ID).First(&rolePerm); r.Error != nil { // assign the record
			if cRes := a.DB.Create(&RolePerm{RoleID: role.ID, PermID: perm.ID}); cRes.Error != nil {
				return cRes.Error
			}
		}
	}

	return nil
}

// AssignRole assigns a given role to a user
// the first parameter is the user id, the second parameter is the role name
// if the role name doesn't have a matching record in the data base an error is returned
// if the user have already a role assigned to him an error is returned
func (a *Rbac) AssignRole(userID uint, roleName string) error {
	// make sure the role exist
	var role Role
	if r := a.DB.Where("name=?", roleName).First(&role); r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return r.Error
	}

	// check if the role is already assigned
	userRole := UserRole{}
	if r := a.DB.Where("user_id=?", userID).Where("role_id=?", role.ID).First(&userRole); r.Error == nil {
		return nil
	}

	// assign the role
	a.DB.Create(&UserRole{UserID: userID, RoleID: role.ID})

	return nil
}

// CheckRole checks if a role is assigned to a user it accepts the user id as the first parameter
// the role as the second parameter
// it returns an error if the role is not present in database.
func (a *Rbac) CheckRole(userID uint, roleName string) (bool, error) {
	// find the role
	var role Role
	if r := a.DB.Where("name=?", roleName).First(&role); r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return false, ErrRoleNotFound
		}
		return false, r.Error
	}

	// check if the role is assigned
	userRole := UserRole{}
	if r := a.DB.Where("user_id=?", userID).Where("role_id=?", role.ID).First(&userRole); r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, r.Error
	}

	return true, nil
}

// CheckPerm checks if a permission is assigned to the role that's assigned to the user.
// it accepts the user id as the first parameter
// the permission as the second parameter
// it returns an error if the permission is not present in the database
func (a *Rbac) CheckPerm(userID uint, permName string) (bool, error) {
	var userRoles []UserRole
	if r := a.DB.Where("user_id=?", userID).Find(&userRoles); r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, r.Error
	}

	//prepare an array of role ids
	var roleIDs []uint
	for _, r := range userRoles {
		roleIDs = append(roleIDs, r.RoleID)
	}

	// find the permission
	var perm Perm
	if r := a.DB.Where("name=?", permName).First(&perm); r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return false, ErrPermNotFound
		}
		return false, r.Error
	}

	// find the role permission
	var rolePerm RolePerm
	r := a.DB.Where("role_id IN (?)", roleIDs).Where("perm_id=?", perm.ID).First(&rolePerm)
	return r.Error == nil, nil
}

// CheckRolePerm checks if a role has the permission assigned
// it accepts the role as the first parameter
// it accepts the permission as the second parameter
// it returns an error if the role is not present in database
// it returns an error if the permission is not present in database
func (a *Rbac) CheckRolePerm(roleName string, permName string) (bool, error) {
	// find the role
	var role Role
	if r := a.DB.Where("name=?", roleName).First(&role); r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return false, errors.New("role not found")
		}
		return false, r.Error
	}

	// find the permission
	var perm Perm
	if r := a.DB.Where("name=?", permName).First(&perm); r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return false, errors.New("permission not found")
		}
		return false, r.Error
	}

	// find the rolePerm
	var rolePerm RolePerm
	if r := a.DB.Where("role_id=?", role.ID).Where("perm_id=?", perm.ID).First(&rolePerm); r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, r.Error
	}

	return true, nil
}

// RevokeRole revokes a user's role
// it returns a error in case of any
func (a *Rbac) RevokeRole(userID uint, roleName string) error {
	// find the role
	var role Role
	if r := a.DB.Where("name=?", roleName).First(&role); r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return r.Error
	}

	// revoke the role
	return a.DB.Where("user_id=?", userID).Where("role_id=?", role.ID).Delete(UserRole{}).Error
}

// RevokePerm revokes a permission from the user's assigned role
// it returns an error in case of any
func (a *Rbac) RevokePerm(userID uint, permName string) error {
	// revoke the permission from all roles of the user
	// find the user roles
	var userRoles []UserRole
	if r := a.DB.Where("user_id=?", userID).Find(&userRoles); r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return nil
		}
		return r.Error
	}

	// find the permission
	var perm Perm
	if r := a.DB.Where("name=?", permName).First(&perm); r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return ErrPermNotFound
		}
		return r.Error
	}

	for _, r := range userRoles { // revoke the permission
		a.DB.Where("role_id=?", r.RoleID).Where("perm_id=?", perm.ID).Delete(RolePerm{})
	}

	return nil
}

// RevokeRolePerm revokes a permission from a given role
// it returns an error in case of any
func (a *Rbac) RevokeRolePerm(roleName string, permName string) error {
	// find the role
	var role Role
	if r := a.DB.Where("name=?", roleName).First(&role); r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return r.Error
	}

	// find the permission
	var perm Perm
	if r := a.DB.Where("name=?", permName).First(&perm); r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return ErrPermNotFound
		}
		return r.Error
	}

	// revoke the permission
	return a.DB.Where("role_id=?", role.ID).Where("perm_id=?", perm.ID).Delete(RolePerm{}).Error
}

// GetRoles returns all stored roles
func (a *Rbac) GetRoles() ([]string, error) {
	var result []string
	var roles []Role
	a.DB.Find(&roles)

	for _, role := range roles {
		result = append(result, role.Name)
	}

	return result, nil
}

// GetUserRoles returns all user assigned roles
func (a *Rbac) GetUserRoles(userID uint) ([]string, error) {
	var result []string
	var userRoles []UserRole
	a.DB.Where("user_id=?", userID).Find(&userRoles)

	for _, r := range userRoles {
		var role Role
		// for every user role get the role name
		if r := a.DB.Where("id=?", r.RoleID).Find(&role); r.Error == nil {
			result = append(result, role.Name)
		}
	}

	return result, nil
}

// GetPerms returns all stored permissions
func (a *Rbac) GetPerms() ([]string, error) {
	var result []string
	var perms []Perm
	a.DB.Find(&perms)

	for _, perm := range perms {
		result = append(result, perm.Name)
	}

	return result, nil
}

// DeleteRole deletes a given role
// if the role is assigned to a user it returns an error
func (a *Rbac) DeleteRole(roleName string) error {
	// find the role
	var role Role
	if r := a.DB.Where("name=?", roleName).First(&role); r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return r.Error
	}

	// check if the role is assigned to a user
	var userRole UserRole
	if r := a.DB.Where("role_id=?", role.ID).First(&userRole); r.Error == nil {
		return ErrDeleteAssignedPerm
	}

	// revoke the assignment of permissions before deleting the role
	a.DB.Where("role_id=?", role.ID).Delete(RolePerm{})
	// delete the role
	a.DB.Where("name=?", roleName).Delete(Role{})

	return nil
}

// DeletePerm deletes a given permission if the permission is assigned to a role it returns an error.
func (a *Rbac) DeletePerm(permName string) error {
	// find the permission
	var perm Perm
	if r := a.DB.Where("name=?", permName).First(&perm); r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return ErrPermNotFound
		}
		return r.Error
	}

	// check if the permission is assigned to a role
	var rolePerm RolePerm
	if r := a.DB.Where("perm_id=?", perm.ID).First(&rolePerm); r.Error == nil {
		return ErrDeleteAssignedPerm
	}

	// delete the permission
	a.DB.Where("name=?", permName).Delete(Perm{})

	return nil
}

func migrateTables(db *gorm.DB) {
	db.AutoMigrate(&Role{})
	db.AutoMigrate(&Perm{})
	db.AutoMigrate(&RolePerm{})
	db.AutoMigrate(&UserRole{})
}
