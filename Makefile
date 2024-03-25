SHELL=cmd.exe
APP_NAME=sales
IMAGE_NAME=sales-api-image
VERSION=1.0.0
NAMESPACE=sales-system
DB_NAMESPACE=database-system
CLUSTER_NAME=test-cluster

##
##
## set build variable in main function to local
##
##
## go build --ldflags "-X main.build=local"
##
################################################################
### Testing Running System
### For testing a simple query on system, Dont forget to `make seed` first
### curl --user "admin@example.com:gophers" https://localhost:3000/v1/users/token
### export TOKEN=$prev_val
### curl -H "Authorization: Bearer ${TOKEN} https://localhost:3000/v1/users/1/2
###
### Access Zipkin
### zipkin: http://localhost:9411
###
### Access metrics directly (4000) or through the sidecar (3001)
### expvarmon -ports=":4000" -vars="build,requests,goroutines,errors,panics,mem.memstats.Alloc"
### expvarmon -ports=":3001" -endpoint="/metrics" -vars="build,requests,goroutines,errors,panics,mem.memstats.Alloc"
### How to install
### go get github.com/divan/expvarmon@latest
###
###
# hey -m GET -c 100 -n 10000 http://localhost:3000/v1/test
#
# to generate private/public key PEM file
# openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:2048
# openssl rsa -pubout -in private.pem -out public.pem
#

###db-lab a cli too to interact with database
##db-lab --host 0.0.0.0 --user postgres --db postgres --pass postgres --ssl disable --port 5432 --driver postgres

all: build-image

build:
	@echo "Building ..."
	chdir ..\${APP_NAME}  && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0 && go build -o ${APP_NAME} .
	@echo "Done!"

admin:
	go run tooling/admin/main.go

#build docker image
#Exp service:1.0.0
build-image:
	docker build --no-cache \
	-f zarf/docker/service.dockerfile \
	-t ${IMAGE_NAME}:${VERSION} \
	.

run:
	docker run -it ${IMAGE_NAME}:${VERSION}

KIND_CLUSTER := ${CLUSTER_NAME}

gen-key:
	go run tooling/sales-admin/main.go

kind-up:
	kind create cluster \
		--image kindest/node:v1.21.14@sha256:8a4e9bb3f415d2bb81629ce33ef9c76ba514c14d707f9797a01e3216376ba093 \
		--name ${KIND_CLUSTER} \
		--config zarf/k8s/kind/kind-config.yaml
	kubectl config set-context --current --namespace=${NAMESPACE}

kind-down:
	kind delete cluster --name ${KIND_CLUSTER}

kind-status:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch --all-namespaces

# first line will edit kind config and add this version to it
kind-apply:
	kustomize build zarf/k8s/kind/database-pod | kubectl apply -f -
# wait for database pod to be available and then brings the sales pod up
	kubectl wait --namespace=${DB_NAMESPACE} --timeout=120s --for=condition=Available deployment/database-pod
	kustomize build zarf/k8s/kind/sales-pod | kubectl apply -f -

kind-load:
	cd zarf/k8s/kind/sales-pod & kustomize edit set image ${IMAGE_NAME}=sales-api:${VERSION}
	kind load docker-image ${IMAGE_NAME}:${VERSION} --name ${KIND_CLUSTER}

kind-logs:
	kubectl logs -l app=${APP_NAME} --all-containers=true -f --tail=100 --namespace=${NAMESPACE}

kind-status-service:
	kubectl get pods -o wide --watch --namespace=${NAMESPACE}

kind-status-db:
	kubectl get pods -o wide --watch --namespace=${DB_NAMESPACE}

kind-restart:
	kubectl rollout restart deployment sales-pod --namespace=${NAMESPACE}

kind-update: all kind-load kind-restart

#first apply must be called - then load
kind-update-apply: all kind-load kind-apply

kind-update-all: all kind-load kind-apply kind-restart

kind-describe:
	kubectl describe pod -l app=${APP_NAME}

tidy:
	go mod tidy
	go mod vendor

