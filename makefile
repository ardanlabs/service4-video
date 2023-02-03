# For full Kind v0.17 release notes: https://github.com/kubernetes-sigs/kind/releases/tag/v0.17.0
#
# Other commands to install.
# go install github.com/divan/expvarmon@latest
# go install github.com/rakyll/hey@latest
#
# http://sales-service.sales-system.svc.cluster.local:4000/debug/pprof
# curl -il sales-service.sales-system.svc.cluster.local:4000/debug/vars
# curl -il sales-service.sales-system.svc.cluster.local:3000/status
#
# RSA Keys
# 	To generate a private/public key PEM file.
# 	$ openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:2048
# 	$ openssl rsa -pubout -in private.pem -out public.pem
#
# Testing Coverage
# 	$ go test -coverprofile p.out
# 	$ go tool cover -html p.out

db:
	go run app/scratch/db/main.go

live:
	curl -sS sales-service.sales-system.svc.cluster.local:4000/debug/liveness | jq

live-local:
	curl -sS localhost:4000/debug/liveness | jq

pgcli:
	pgcli postgresql://postgres:postgres@database-service.sales-system.svc.cluster.local

pgcli-local:
	pgcli postgresql://postgres:postgres@localhost

jwt:
	go run app/scratch/jwt/main.go

status:
	curl -il sales-service.sales-system.svc.cluster.local:3000/status

auth:
	curl -il -H "Authorization: Bearer ${TOKEN}" sales-service.sales-system.svc.cluster.local:3000/auth

auth-local:
	curl -il -H "Authorization: Bearer ${TOKEN}" localhost:3000/auth

run:
	go run app/services/sales-api/main.go | go run app/tooling/logfmt/main.go

run-help:
	go run app/services/sales-api/main.go --help

tidy:
	go mod tidy
	go mod vendor

metrics-local:
	expvarmon -ports=":4000" -vars="build,requests,goroutines,errors,panics,mem:memstats.Alloc"

metrics-view:
	expvarmon -ports="sales-service.sales-system.svc.cluster.local:4000" -vars="build,requests,goroutines,errors,panics,mem:memstats.Alloc"

test-load-local:
	hey -m GET -c 100 -n 10000 http://localhost:3000/status

test-load:
	hey -m GET -c 100 -n 10000 http://sales-service.sales-system.svc.cluster.local:3000/status

test-token-local:
	curl -il --user "admin@example.com:gophers" http://localhost:3000/users/token/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1

test-token:
	curl -il --user "admin@example.com:gophers" http://sales-service.sales-system.svc.cluster.local:3000/users/token/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1

# export TOKEN="COPY TOKEN STRING FROM LAST CALL"

test-users-local:
	curl -il -H "Authorization: Bearer ${TOKEN}" http://localhost:3000/users/1/2

test-users:
	curl -il -H "Authorization: Bearer ${TOKEN}" http://sales-service.sales-system.svc.cluster.local:3000/users/1/2

# ==============================================================================
# Running tests within the local computer
# go install honnef.co/go/tools/cmd/staticcheck@latest
# go install golang.org/x/vuln/cmd/govulncheck@latest

test:
	CGO_ENABLED=0 go test -count=1 ./...
	CGO_ENABLED=0 go vet ./...
	staticcheck -checks=all ./...
	govulncheck ./...

# ==============================================================================
# Building containers

# $(shell git rev-parse --short HEAD)
VERSION := 1.0

all: sales

sales:
	docker build \
		-f zarf/docker/dockerfile.sales-api \
		-t sales-api:$(VERSION) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

# ==============================================================================
# Running from within k8s/kind

GOLANG       := golang:1.19
ALPINE       := alpine:3.17
KIND         := kindest/node:v1.25.3
POSTGRES     := postgres:15-alpine
VAULT        := hashicorp/vault:1.12
ZIPKIN       := openzipkin/zipkin:2.23
TELEPRESENCE := docker.io/datawire/tel2:2.10.4

KIND_CLUSTER := ardan-starter-cluster

dev-kind-up:
	kind create cluster \
		--image kindest/node:v1.25.3@sha256:f52781bc0d7a19fb6c405c2af83abfeb311f130707a0e219175677e366cc45d1 \
		--name $(KIND_CLUSTER) \
		--config zarf/k8s/dev/kind-config.yaml
	kubectl wait --timeout=120s --namespace=local-path-storage --for=condition=Available deployment/local-path-provisioner

	kind load docker-image $(POSTGRES) --name $(KIND_CLUSTER)

dev-up: dev-kind-up
	kind load docker-image $(TELEPRESENCE) --name $(KIND_CLUSTER)
	telepresence --context=kind-$(KIND_CLUSTER) helm install
	telepresence --context=kind-$(KIND_CLUSTER) connect

dev-up-wsl2: dev-kind-up
	kind load docker-image $(TELEPRESENCE) --name $(KIND_CLUSTER)
	telepresence --context=kind-$(KIND_CLUSTER) helm install
	sudo telepresence --context=kind-$(KIND_CLUSTER) connect

dev-down:
	telepresence quit -s
	kind delete cluster --name $(KIND_CLUSTER)

dev-kind-down:
	kind delete cluster --name $(KIND_CLUSTER)

dev-status:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch --all-namespaces

dev-load:
	kind load docker-image sales-api:$(VERSION) --name $(KIND_CLUSTER)

dev-apply:
	kustomize build zarf/k8s/dev/database | kubectl apply -f -
	kubectl wait --timeout=120s --namespace=sales-system --for=condition=Available deployment/database

	kustomize build zarf/k8s/dev/sales | kubectl apply -f -
	kubectl wait --timeout=120s --namespace=sales-system --for=condition=Available deployment/sales

dev-restart:
	kubectl rollout restart deployment sales --namespace=sales-system

dev-logs:
	kubectl logs --namespace=sales-system -l app=sales --all-containers=true -f --tail=100 --max-log-requests=6 | go run app/tooling/logfmt/main.go -service=SALES-API

dev-describe:
	kubectl describe nodes
	kubectl describe svc

dev-describe-deployment:
	kubectl describe deployment --namespace=sales-system sales

dev-describe-sales:
	kubectl describe pod --namespace=sales-system -l app=sales

dev-describe-tel:
	kubectl describe pod --namespace=ambassador -l app=traffic-manager

dev-update: all dev-load dev-restart

dev-update-apply: all dev-load dev-apply

dev-logs-db:
	kubectl logs --namespace=sales-system -l app=database --all-containers=true -f --tail=100

dev-logs-init:
	kubectl logs --namespace=sales-system -l app=sales -f --tail=100 -c init-db
