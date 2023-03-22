package rbac_test

import (
	"os"
	"testing"

	"github.com/bingoohuang/gg/pkg/rbac"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func TestMain(m *testing.M) {
	dsn := "root:root@tcp(127.0.0.1:3306)/db_test?charset=utf8mb4&parseTime=True&loc=Local"

	db, _ = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())
}

func TestNewRole(t *testing.T) {
	auth := rbac.New(rbac.Options{TablesPrefix: "rbac_", DB: db})

	// test create role
	err := auth.NewRole("role-a")
	if err != nil {
		t.Error("an error was not expected while creating role ", err)
	}

	var c int64
	res := db.Model(rbac.Role{}).Where("name=?", "role-a").Count(&c)
	if res.Error != nil {
		t.Error("unexpected error while storing role: ", err)
	}
	if c == 0 {
		t.Error("role has not been stored")
	}

	// test duplicated entries
	auth.NewRole("role-a")
	auth.NewRole("role-a")
	auth.NewRole("role-a")
	db.Model(rbac.Role{}).Where("name=?", "role-a").Count(&c)
	if c > 1 {
		t.Error("unexpected duplicated entries for role")
	}

	// clean up
	db.Where("name=?", "role-a").Delete(rbac.Role{})
}

func TestNewPerm(t *testing.T) {
	auth := rbac.New(rbac.Options{TablesPrefix: "rbac_", DB: db})

	// test create perm
	err := auth.NewPerm("perm-a")
	if err != nil {
		t.Error("an error was not expected while creating permision ", err)
	}

	var c int64
	res := db.Model(rbac.Perm{}).Where("name=?", "perm-a").Count(&c)
	if res.Error != nil {
		t.Error("unexpected error while storing perm: ", err)
	}
	if c == 0 {
		t.Error("perm has not been stored")
	}

	// test duplicated entries
	auth.NewPerm("perm-a")
	auth.NewPerm("perm-a")
	auth.NewPerm("perm-a")
	db.Model(rbac.Role{}).Where("name=?", "perm-a").Count(&c)
	if c > 1 {
		t.Error("unexpected duplicated entries for perm")
	}

	// clean up
	db.Where("name=?", "perm-a").Delete(rbac.Perm{})
}

func TestAssignPerm(t *testing.T) {
	auth := rbac.New(rbac.Options{TablesPrefix: "rbac_", DB: db})
	// first create a role
	err := auth.NewRole("role-a")
	if err != nil {
		t.Error("unexpected error while creating role.", err)
	}

	// second test create perms
	err = auth.NewPerm("perm-a")
	if err != nil {
		t.Error("unexpected error while creating perm to be assigned.", err)
	}
	err = auth.NewPerm("perm-b")
	if err != nil {
		t.Error("unexpected error while creating perm to be assigned.", err)
	}

	// assign the perms
	err = auth.AssignPerms("role-a", "perm-a", "perm-b")
	if err != nil {
		t.Error("unexpected error while assigning perms.", err)
	}

	// assign to missing role
	err = auth.AssignPerms("role-aa", "perm-a", "perm-b")
	if err == nil {
		t.Error("expecting error when assigning to missing role")
	}

	// assign to missing perm
	err = auth.AssignPerms("role-a", "perm-aa")
	if err == nil {
		t.Error("expecting error when assigning missing perm")
	}

	var r rbac.Role
	db.Where("name=?", "role-a").First(&r)
	var rolePermsCount int64
	db.Model(rbac.RolePerm{}).Where("role_id=?", r.ID).Count(&rolePermsCount)
	if rolePermsCount != 2 {
		t.Error("failed assigning roles to perm")
	}

	// clean up
	db.Where("role_id=?", r.ID).Delete(rbac.RolePerm{})
	db.Where("name=?", "role-a").Delete(rbac.Role{})
	db.Where("name=?", "perm-a").Delete(rbac.Perm{})
	db.Where("name=?", "perm-b").Delete(rbac.Perm{})
}

func TestAssignRole(t *testing.T) {
	auth := rbac.New(rbac.Options{TablesPrefix: "rbac_", DB: db})

	// first create a role
	err := auth.NewRole("role-a")
	if err != nil {
		t.Error("unexpected error while creating role to be assigned.", err)
	}

	// assign the role
	err = auth.AssignRole(1, "role-a")
	if err != nil {
		t.Error("unexpected error while assigning role.", err)
	}

	// double assign the role
	err = auth.AssignRole(1, "role-a")
	if err != nil {
		t.Error("unexpected error when assign a role to user more than one time", err)
	}

	// assign a second role
	auth.NewRole("role-b")
	err = auth.AssignRole(1, "role-b")
	if err != nil {
		t.Error("un expected error when assigning a second role. ", err)
	}

	// assign missing role
	err = auth.AssignRole(1, "role-aa")
	if err == nil {
		t.Error("expecting an error when assigning role to a user")
	}

	var r rbac.Role
	db.Where("name=?", "role-a").First(&r)
	var userRoles int64
	db.Model(rbac.UserRole{}).Where("role_id=?", r.ID).Count(&userRoles)
	if userRoles != 1 {
		t.Error("failed assigning roles to perm")
	}

	// clean up
	db.Where("user_id=?", 1).Delete(rbac.UserRole{})
	db.Where("name=?", "role-a").Delete(rbac.Role{})
	db.Where("name=?", "role-b").Delete(rbac.Role{})
}

func TestCheckRole(t *testing.T) {
	auth := rbac.New(rbac.Options{TablesPrefix: "rbac_", DB: db})

	// first create a role and assign it to a user
	err := auth.NewRole("role-a")
	if err != nil {
		t.Error("unexpected error while creating role to be assigned.", err)
	}
	// assign the role
	err = auth.AssignRole(1, "role-a")
	if err != nil {
		t.Error("unexpected error while assigning role.", err)
	}

	// assert
	ok, err := auth.CheckRole(1, "role-a")
	if err != nil {
		t.Error("unexpected error while checking user for assigned role.", err)
	}
	if !ok {
		t.Error("failed to check assinged role")
	}

	// check aa missing role
	_, err = auth.CheckRole(1, "role-aa")
	if err == nil {
		t.Error("expecting an error when checking a missing role")
	}

	// check a missing user
	ok, _ = auth.CheckRole(11, "role-a")
	if ok {
		t.Error("expecting false when checking missing role")
	}

	// clean up
	var r rbac.Role
	db.Where("name=?", "role-a").First(&r)
	db.Where("role_id=?", r.ID).Delete(rbac.UserRole{})
	db.Where("name=?", "role-a").Delete(rbac.Role{})
}

func TestCheckPerm(t *testing.T) {
	auth := rbac.New(rbac.Options{TablesPrefix: "rbac_", DB: db})
	// first create a role
	err := auth.NewRole("role-a")
	if err != nil {
		t.Error("unexpected error while creating role.", err)
	}

	// create perms
	err = auth.NewPerm("perm-a")
	if err != nil {
		t.Error("unexpected error while creating perm to be assigned.", err)
	}
	err = auth.NewPerm("perm-b")
	if err != nil {
		t.Error("unexpected error while creating perm to be assigned.", err)
	}

	// assign the perms
	err = auth.AssignPerms("role-a", "perm-a", "perm-b")
	if err != nil {
		t.Error("unexpected error while assigning perms.", err)
	}

	// test when no role is assigned
	ok, err := auth.CheckPerm(1, "perm-a")
	if err != nil {
		t.Error("expecting error to be nil when no role is assigned")
	}
	if ok {
		t.Error("expecting false to be returned when no role is assigned")
	}

	// assign the role
	err = auth.AssignRole(1, "role-a")
	if err != nil {
		t.Error("unexpected error while assigning role.", err)
	}

	// test a perm of an assigned role
	ok, err = auth.CheckPerm(1, "perm-a")
	if err != nil {
		t.Error("unexpected error while checking perm of a user.", err)
	}
	if !ok {
		t.Error("expecting true to be returned")
	}

	// check when user does not have roles
	ok, _ = auth.CheckPerm(111, "perm-a")
	if ok {
		t.Error("expecting an false when checking perm of not assigned  user")
	}

	// test assigning missing perm
	_, err = auth.CheckPerm(1, "perm-aa")
	if err == nil {
		t.Error("expecting an error when checking a missing perm")
	}

	// check for an exist but not assigned perm
	auth.NewPerm("perm-c")
	ok, _ = auth.CheckPerm(1, "perm-c")
	if ok {
		t.Error("expecting false when checking for not assigned perms")
	}

	// clean up
	var r rbac.Role
	db.Where("name=?", "role-a").First(&r)
	db.Where("role_id=?", r.ID).Delete(rbac.UserRole{})
	db.Where("role_id=?", r.ID).Delete(rbac.RolePerm{})
	db.Where("name=?", "perm-a").Delete(rbac.Perm{})
	db.Where("name=?", "perm-b").Delete(rbac.Perm{})
	db.Where("name=?", "perm-c").Delete(rbac.Perm{})
	db.Where("name=?", "role-a").Delete(rbac.Role{})
}

func TestCheckRolePerm(t *testing.T) {
	auth := rbac.New(rbac.Options{TablesPrefix: "rbac_", DB: db})

	// first create a role
	err := auth.NewRole("role-a")
	if err != nil {
		t.Error("unexpected error while creating role.", err)
	}

	// second test create perms
	err = auth.NewPerm("perm-a")
	if err != nil {
		t.Error("unexpected error while creating perm to be assigned.", err)
	}
	err = auth.NewPerm("perm-b")
	if err != nil {
		t.Error("unexpected error while creating perm to be assigned.", err)
	}

	// third assign the perms
	err = auth.AssignPerms("role-a", "perm-a", "perm-b")
	if err != nil {
		t.Error("unexpected error while assigning perms.", err)
	}

	// check the role perm
	ok, err := auth.CheckRolePerm("role-a", "perm-a")
	if err != nil {
		t.Error("unexpected error while checking role perm.", err)
	}
	if !ok {
		t.Error("failed assigning roles to perm check")
	}

	// check a missing role
	_, err = auth.CheckRolePerm("role-aa", "perm-a")
	if err == nil {
		t.Error("expecting an error when checking permisson of missing role")
	}

	// check with missing perm
	_, err = auth.CheckRolePerm("role-a", "perm-aa")
	if err == nil {
		t.Error("expecting an error when checking missing perm")
	}

	// check with not assigned perm
	auth.NewPerm("perm-c")
	ok, _ = auth.CheckRolePerm("role-a", "perm-c")
	if ok {
		t.Error("expecting false when checking a missing perm")
	}

	// clean up
	var r rbac.Role
	db.Where("name=?", "role-a").First(&r)
	db.Where("role_id=?", r.ID).Delete(rbac.RolePerm{})
	db.Where("name=?", "perm-a").Delete(rbac.Perm{})
	db.Where("name=?", "perm-b").Delete(rbac.Perm{})
	db.Where("name=?", "perm-c").Delete(rbac.Perm{})
	db.Where("name=?", "role-a").Delete(rbac.Role{})
}

func TestRevokeRole(t *testing.T) {
	auth := rbac.New(rbac.Options{TablesPrefix: "rbac_", DB: db})

	// first create a role
	err := auth.NewRole("role-a")
	if err != nil {
		t.Error("unexpected error while creating role.", err)
	}

	// assign the role
	err = auth.AssignRole(1, "role-a")
	if err != nil {
		t.Error("unexpected error while assigning role.", err)
	}

	// test
	err = auth.RevokeRole(1, "role-a")
	if err != nil {
		t.Error("unexpected error revoking user role.", err)
	}
	// revoke missing role
	err = auth.RevokeRole(1, "role-aa")
	if err == nil {
		t.Error("expecting error when revoking a missing role")
	}

	var c int64
	db.Model(rbac.UserRole{}).Where("user_id=?", 1).Count(&c)
	if c != 0 {
		t.Error("failed assert revoking user role")
	}

	var r rbac.Role
	db.Where("name=?", "role-a").First(&r)
	db.Where("role_id=?", r.ID).Delete(rbac.UserRole{})
	db.Where("name=?", "role-a").Delete(rbac.Role{})
}

func TestRevokePerm(t *testing.T) {
	auth := rbac.New(rbac.Options{TablesPrefix: "rbac_", DB: db})

	// first create a role
	err := auth.NewRole("role-a")
	if err != nil {
		t.Error("unexpected error while creating role.", err)
	}
	// second test create perms
	err = auth.NewPerm("perm-a")
	if err != nil {
		t.Error("unexpected error while creating perm to be assigned.", err)
	}
	err = auth.NewPerm("perm-b")
	if err != nil {
		t.Error("unexpected error while creating perm to be assigned.", err)
	}

	// third assign the perms
	err = auth.AssignPerms("role-a", "perm-a", "perm-b")
	if err != nil {
		t.Error("unexpected error while assigning perms.", err)
	}

	// assign the role
	err = auth.AssignRole(1, "role-a")
	if err != nil {
		t.Error("unexpected error while assigning role.", err)
	}

	// case: user not assigned role
	err = auth.RevokePerm(11, "perm-a")
	if err != nil {
		t.Error("expecting error to be nil", err)
	}

	// test
	err = auth.RevokePerm(1, "perm-a")
	if err != nil {
		t.Error("unexpected error while revoking role perms.", err)
	}

	// revoke missing permissin
	err = auth.RevokePerm(1, "perm-aa")
	if err == nil {
		t.Error("expecting error when revoking a missing perm")
	}

	// assert, count assigned perm, should be one
	var r rbac.Role
	db.Where("name=?", "role-a").First(&r)
	var c int64
	db.Model(rbac.RolePerm{}).Where("role_id=?", r.ID).Count(&c)
	if c != 1 {
		t.Error("failed assert revoking perm role")
	}

	// clean up
	db.Where("role_id=?", r.ID).Delete(rbac.UserRole{})
	db.Where("role_id=?", r.ID).Delete(rbac.RolePerm{})
	db.Where("name=?", "perm-a").Delete(rbac.Perm{})
	db.Where("name=?", "perm-b").Delete(rbac.Perm{})
	db.Where("name=?", "role-a").Delete(rbac.Role{})
}

func TestRevokeRolePerm(t *testing.T) {
	auth := rbac.New(rbac.Options{TablesPrefix: "rbac_", DB: db})

	// first create a role
	err := auth.NewRole("role-a")
	if err != nil {
		t.Error("unexpected error while creating role.", err)
	}
	// second test create perms
	err = auth.NewPerm("perm-a")
	if err != nil {
		t.Error("unexpected error while creating perm to be assigned.", err)
	}
	err = auth.NewPerm("perm-b")
	if err != nil {
		t.Error("unexpected error while creating perm to be assigned.", err)
	}

	// third assign the perms
	err = auth.AssignPerms("role-a", "perm-a", "perm-b")
	if err != nil {
		t.Error("unexpected error while assigning perms.", err)
	}

	// test revoke missing role
	err = auth.RevokeRolePerm("role-aa", "perm-a")
	if err == nil {
		t.Error("expecting an error when revoking a missing role")
	}

	// test revoke missing perm
	err = auth.RevokeRolePerm("role-a", "perm-aa")
	if err == nil {
		t.Error("expecting an error when revoking a missing perm")
	}

	err = auth.RevokeRolePerm("role-a", "perm-a")
	if err != nil {
		t.Error("unexpected error while revoking role perms.", err)
	}
	// assert, count assigned perm, should be one
	var r rbac.Role
	db.Where("name=?", "role-a").First(&r)
	var c int64
	db.Model(rbac.RolePerm{}).Where("role_id=?", r.ID).Count(&c)
	if c != 1 {
		t.Error("failed assert revoking perm role")
	}

	// clean up
	db.Where("role_id=?", r.ID).Delete(rbac.RolePerm{})
	db.Where("name=?", "perm-a").Delete(rbac.Perm{})
	db.Where("name=?", "perm-b").Delete(rbac.Perm{})
	db.Where("name=?", "role-a").Delete(rbac.Role{})
}

func TestGetRoles(t *testing.T) {
	auth := rbac.New(rbac.Options{TablesPrefix: "rbac_", DB: db})

	// first create roles
	err := auth.NewRole("role-a")
	if err != nil {
		t.Error("unexpected error while creating role.", err)
	}
	err = auth.NewRole("role-b")
	if err != nil {
		t.Error("unexpected error while creating role.", err)
	}

	// test
	roles, err := auth.GetRoles()
	// check
	if len(roles) != 2 {
		t.Error("failed assert getting roles")
	}
	db.Where("name=?", "role-a").Delete(rbac.Role{})
	db.Where("name=?", "role-b").Delete(rbac.Role{})
}

func TestGetPerms(t *testing.T) {
	auth := rbac.New(rbac.Options{TablesPrefix: "rbac_", DB: db})

	// first create perm
	err := auth.NewPerm("perm-a")
	if err != nil {
		t.Error("unexpected error while creating perm.", err)
	}
	err = auth.NewPerm("perm-b")
	if err != nil {
		t.Error("unexpected error while creating perm.", err)
	}

	// test
	perms, err := auth.GetPerms()
	// check
	if len(perms) != 2 {
		t.Error("failed assert getting perm")
	}
	db.Where("name=?", "perm-a").Delete(rbac.Perm{})
	db.Where("name=?", "perm-b").Delete(rbac.Perm{})
}

func TestDeleteRole(t *testing.T) {
	auth := rbac.New(rbac.Options{TablesPrefix: "rbac_", DB: db})

	err := auth.NewRole("role-a")
	if err != nil {
		t.Error("unexpected error while creating role.", err)
	}

	// test delete a missing role
	err = auth.DeleteRole("role-aa")
	if err == nil {
		t.Error("expecting an error when deleting a missing role")
	}

	// test delete an assigned role
	auth.AssignRole(1, "role-a")
	err = auth.DeleteRole("role-a")
	if err == nil {
		t.Error("expecting an error when deleting an assigned role")
	}
	auth.RevokeRole(1, "role-a")

	err = auth.DeleteRole("role-a")
	if err != nil {
		t.Error("unexpected error while deleting role.", err)
	}

	var c int64
	db.Model(rbac.Role{}).Count(&c)
	if c != 0 {
		t.Error("failed assert deleting role")
	}
}

func TestDeletePerm(t *testing.T) {
	auth := rbac.New(rbac.Options{TablesPrefix: "rbac_", DB: db})

	err := auth.NewPerm("perm-a")
	if err != nil {
		t.Error("unexpected error while creating perm.", err)
	}

	// delete missing perm
	err = auth.DeletePerm("perm-aa")
	if err == nil {
		t.Error("expecting an error when deleting a missing perm")
	}

	// delete an assigned perm
	auth.NewRole("role-a")
	auth.AssignPerms("role-a", "perm-a")

	// delete assinged perm
	err = auth.DeletePerm("perm-a")
	if err == nil {
		t.Error("expecting an error when deleting assigned perm")
	}

	auth.RevokeRolePerm("role-a", "perm-a")

	err = auth.DeletePerm("perm-a")
	if err != nil {
		t.Error("unexpected error while deleting perm.", err)
	}

	var c int64
	db.Model(rbac.Perm{}).Count(&c)
	if c != 0 {
		t.Error("failed assert deleting perm")
	}

	// clean up
	auth.DeleteRole("role-a")
}

func TestGetUserRoles(t *testing.T) {
	auth := rbac.New(rbac.Options{TablesPrefix: "rbac_", DB: db})

	// first create a role
	auth.NewRole("role-a")
	auth.NewRole("role-b")
	auth.AssignRole(1, "role-a")
	auth.AssignRole(1, "role-b")

	roles, _ := auth.GetUserRoles(1)
	if len(roles) != 2 {
		t.Error("expeting two roles to be returned")
	}

	if !sliceHasString(roles, "role-a") {
		t.Error("missing role in returned roles")
	}

	if !sliceHasString(roles, "role-b") {
		t.Error("missing role in returned roles")
	}

	db.Where("user_id=?", 1).Delete(rbac.UserRole{})
	db.Where("name=?", "role-a").Delete(rbac.Role{})
	db.Where("name=?", "role-b").Delete(rbac.Role{})
}

func sliceHasString(s []string, val string) bool {
	for _, v := range s {
		if v == val {
			return true
		}
	}

	return false
}
