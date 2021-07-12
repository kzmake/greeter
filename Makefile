.PHONY: all
all: tidy fmt lint proto

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: fmt
fmt:
	goimports -l -local "github.com/kzmake/greeter" -w .

.PHONY: lint
lint:
	golangci-lint run

.PHONY: proto
proto:
	@for file in api/**/**/*.proto; do \
		protoc \
			--proto_path=.:. \
			--go_out=paths=source_relative:. \
			--go-grpc_out=paths=source_relative:. \
			--grpc-gateway_out=logtostderr=true,paths=source_relative:. \
			--openapiv2_out=logtostderr=true:. \
			$$file; \
		echo "generated from $$file"; \
	done


.PHONY: build
build: certs
	docker build -t kzmake/greeter .

.PHONY: publish
publish: build
	docker push kzmake/greeter

.PHONY: certs
certs: certs/gen certs/view

.PHONY: certs/gen
certs/gen: certs/gen/ca certs/gen/server certs/gen/client

.PHONY: certs/gen/ca
certs/gen/ca:
	openssl genrsa -out certs/ca.key 2048
	openssl req -x509 -new -nodes -key certs/ca.key -subj "/CN=kzmake.example.com" -days 36500 -out certs/ca.crt

.PHONY: certs/gen/server
certs/gen/server: certs/gen/server/greeter

.PHONY: certs/gen/server/greeter
certs/gen/server/greeter:
	openssl genrsa -out certs/server.greeter.key 2048
	openssl req -new -key certs/server.greeter.key -subj "/CN=*.greeter.default.svc.cluster.local"> certs/server.greeter.csr
	openssl x509 -req -extfile certs/server.greeter.ext.conf -in certs/server.greeter.csr -CA certs/ca.crt -CAkey certs/ca.key -CAcreateserial -days 36500 -out certs/server.greeter.crt

.PHONY: certs/gen/server
certs/gen/server: certs/gen/client/gateway

.PHONY: certs/gen/client/gateway
certs/gen/client:
	openssl genrsa -out certs/client.gateway.key 2048
	openssl req -new -key certs/client.gateway.key -subj "/CN=gateway.default.svc.cluster.local" > certs/client.gateway.csr
	openssl x509 -req -in certs/client.gateway.csr -CA certs/ca.crt -CAkey certs/ca.key -CAcreateserial -days 36500 -out certs/client.gateway.crt

.PHONY: certs/view
certs/view:
	openssl rsa -text -noout -in certs/ca.key
	openssl x509 -text -noout -in certs/ca.crt
	openssl rsa -text -noout -in certs/server.greeter.key
	openssl x509 -text -noout -in certs/server.greeter.crt
	openssl rsa -text -noout -in certs/client.gateway.key
	openssl x509 -text -noout -in certs/client.gateway.crt

.PHONY: kind
kind:
	kind create cluster --config kind.yaml || true

.PHONY: kind/clean
kind/clean:
	kind delete clusters greeter

.PHONY: dev
dev:
	skaffold run

.PHONY: request
request:
	curl -i localhost:8080/hello -H "Content-Type: application/json" -d '{"name": "alice"}'
