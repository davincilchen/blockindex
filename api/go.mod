module api

go 1.16

replace example.com/routes => ./routes

replace example.com/models => ./models

replace example.com/controllers => ./controllers

require (
	example.com/routes v0.0.0-00010101000000-000000000000
	gorm.io/gorm v1.23.4
)

require (
	example.com/models v0.0.0-00010101000000-000000000000
	github.com/caarlos0/env v3.5.0+incompatible
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/gomodule/redigo v1.8.8 // indirect
	github.com/joho/godotenv v1.4.0
	golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gorm.io/driver/postgres v1.3.3
)
