# Elif Digital SSO 


## Important note

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




