# rbac

forkd from https://github.com/harranali/authority

Role Based Access Control (RBAC) Go package with database persistence

# Features

- Create Roles/Permissions
- Assign Permissions to Roles/Multiple Roles to Users
- Check User's Roles/Permissions/Role's Permissions
- Revoke User's Roles/User's Permissions/ole's permissions
- List User's Roles/All Roles/All Permissions
- Delete Roles/Permissions

## Test

1. `docker run --name mysql -e MYSQL_ROOT_PASSWORD=root -p 3306:3306 -d mysql:5.7.34`
1. `go test`

# Usage

```go
// initiate the database (using mysql)
dsn := "dbuser:dbpassword@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})

// initiate authority
auth := authority.New(authority.Options{ TablesPrefix: "authority_", DB: db })

// create role
err := auth.CreateRole("role-1")

// create permissions
err := auth.NewPerm("permission-1")
err = auth.NewPerm("permission-2")
err = auth.NewPerm("permission-3")

// assign the permissions to the role
err := auth.AssignPerm("role-1", "permission-1", "permission-2", "permission-3")

// assign a role to user (user id = 1) 
err = auth.AssignRole(1, "role-a")

// check if the user have a given role
ok, err := auth.CheckRole(1, "role-a")

// check if a user have a given permission 
ok, err := auth.CheckPerm(1, "permission-d")

// check if a role have a given permission
ok, err := auth.CheckRolePerm("role-a", "permission-a")
```
