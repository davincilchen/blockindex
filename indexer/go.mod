module example2.com/indexer

go 1.16

replace example2.com/ethprocess => ./ethprocess

replace example2.com/models => ./models

require (
	github.com/caarlos0/env v3.5.0+incompatible
	github.com/ethereum/go-ethereum v1.10.17
	github.com/joho/godotenv v1.4.0
	gorm.io/driver/postgres v1.3.1 // indirect
)

require (
	example2.com/ethprocess v0.0.0-00010101000000-000000000000
	example2.com/models v0.0.0-00010101000000-000000000000
)
