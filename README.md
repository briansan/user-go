# BreadTech Base User Model

## Introduction
The start of any great project (and company for that matter) begins with a solid representation of its user base. This particular framework aims to provide some go-based boilerplate code that one can use to bootstrap a user-based project.

# Getting Started
This project uses Mongo as the backend so for now, you'll have to commit to using this fine NoSQL database. While there may or may not be future plans to support alternative database technologies, We'd like to express my appreciation for Mongo in its ability to solve many data persistence problems while providing a rather intuitive interface.

## Provision

### Mongo
- The fastest way to get a Mongo database up and running is with [Docker](https://get.docker.com).

```
$ docker run -p 27017:27017 mongo # use -d flag to run as detached (in background)
```

- Up next is to provision the database. For your convenience, a provision script has been provided in `infrastructure/provision_mongo.js` that uses some sample credentials and but feel free to customize it to your liking.

```
$ mongo < infrastructure/provision_mongo.js
```

## Run
- This framework uses [viper](https://github.com/spf13/viper) for configuration management which allows the use of configuration files, environment variables, and more to configure a project. We recommending using config files for this job. By default, the project searches for a file called `bt-config[.yml|.json|.toml]` but feel to change the name by modifying the `ConfigFileName` variable in `config/config.go`.
- Once you're ready to roll, run with:

```
$ go run main.go
```

## Usage

### Creating a user
```
$ curl localhost:8888/api/v1/users -XPOST -HContent-type:application/json -d '{"username": "bk", "password": "applebananacoke", "email": "bk@example.com"}'
```

### Authentication
```
$ curl bk:applebananacoke@localhost:8888/api/v1/login
```

- The returned JSON object has the field `session` which you can use to authentication instead of passing along credentials on every request. For command-line usgae, we recommend using an environment variable to store the session.

```
$ TOKEN=$(curl bk:applebananacoke@localhost:8888/api/v1/login | jq -r .session)
$ curl -H "Authorization: Bearer $TOKEN" localhost:8888/api/v1/users/bk
```

### Modification
```
$ # Modify
$ curl -H "Authorization: Bearer $TOKEN" localhost:8888/api/v1/users/bk -XPATCH -HContent-type:application/json -d '{"email": "kb@example.com"}' 
$ # Delete
$ curl -H "Authorization: Bearer $TOKEN" localhost:8888/api/v1/users/bk -XDELETE
```

### Administration
- On first time execution, a full-privileged admin account is created and the password is set to the value of `SECRET` in the config file (`BT_SECRET` for environment variable). You can login as the boss with:

```
$ curl boss:gammahouseigloo@localhost:8888/api/v1/login
```
