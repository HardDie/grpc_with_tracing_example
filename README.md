# How to run

## Run jaeger server
```
docker run --rm -d -p 16686:16686 -p 14268:14268 --name jaeger jaegertracing/all-in-one
```

## Compile go files from proto
```
make proto_server
make proto_client
```

## Run app
```
# In first console tab
go run server/main.go

# In second console tab
go run client/main.go

# In third console tab
grpcurl -plaintext localhost:9001 client.Client.Test
```
