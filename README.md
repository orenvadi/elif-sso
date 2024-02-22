# Elif Digital SSO 


## How to run

To run the project you must have a Postgres database, configs are located in config/config.yaml

**Migrate the database**

```sh
make migrate
```

if it does not work, you have to adjust the main Makefile

**Run the server on local machine**

```sh
make run_local
```


## How to change and regenerate protos

### Requirements

- [buf cli](https://buf.build/docs/installation) 


At first you should change directory

```sh
cd protos
```

Then

```sh
buf mod update
```

```sh
buf generate
```


## Important note!!!

I could not solve issue with generated `sso.pb.go`
so every time you regenerate protos

**Replace**

```go
	_ "buf/validate"
```

**With this**
```go
	_ "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
```






## TODO 

-[x] Finish update of users

-[] Email confirm
-[] Reset password

-[] Get user data
-[] Refresh token










