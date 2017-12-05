# BreadTech User REST API

### User
```
id             bson.ObjectID
username       string
password       string
email          string
role           int
```

### Task
This is a non-existent data model, but for the sake of providing
some context of user permissions, imagine that it is some object
that has a one-to-many relationship with the user.

## Permissions
```
CreateUser:
  can create user
ModifySelfTasks:
  can view/modify self
  can CRUD tasks where task.user_id = self
ModifyAllUsers: 
  can CRUD all users
ModifyAllUsersRestricted:
  like ModifyAllUsers except:
    cannot modify users where role = Admin
    cannot modify user.role
ViewAllTasks:
  can read all tasks
ModifyAllTasks: 
  can CRUD all tasks
```

## Roles
```
Anon: CreateUser
User: ModifySelfTasks
Manager: User + ModifyAllUserRestricted + ViewAllTasks
Admin: Manager + ModifyAllUsers + ModifyAllTasks
```

## API
all routes mounted on `/api/v1`

### GET /service/ping
- allows: All
- details: healthcheck endpoint reporting version

### GET /login
- allows: All
- details: presents authenticated user with 1 hr jwt session
- requires: BasicAuth

### GET /users
- allows: Manager, Admin
- details: retrieves all users
- requires: Bearer JWT Auth

### POST /users
- allows: Anon, Manager, Admin
- details: creates a user
- requires: Bearer JWT Auth

### GET /users/:userID
- allows: User\*, Manager, Admin
- details: retrieves a user by id
- requires: Bearer JWT Auth

### PATCH /users/:userID
- allows: User\*, Manager, Admin
- details: updates a user by field
- requires: Bearer JWT Auth

### DELETE /users/:userID
- allows: User\*, Manager, Admin
- details: deletes a user and all associated tasks
- requires: Bearer JWT Auth

[^*]: only allowed for resources owned by that role's user
