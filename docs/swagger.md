# Generate swagger documentation

see: https://github.com/swaggo/swag

```sh
go install github.com/swaggo/swag/cmd/swag@v1.16.6
go get github.com/swaggo/gin-swagger
go get github.com/swaggo/files
swag init
```

result can be obtained as: http://localhost:8080/swagger/index.html