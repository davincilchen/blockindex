# Server
## Usage


### Start service

```docker
docker-compose up -d
```

```curl
curl http://localhost:8888/v1/blocks?limit=20
curl http://localhost:8888/v1/blocks?limit=20&&chain_id=97

curl http://localhost:8888/v1/blocks/:num

curl http://localhost:8888/v1/transaction/:txHash
```

### Stop service

```docker

docker-compose down
docker rmi blockindex_api-service blockindex_api-service

```
