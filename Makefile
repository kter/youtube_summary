.PHONY: init-dev init-prd plan-dev plan-prd apply-dev apply-prd destroy-dev destroy-prd \
	build-layer build-frontend deploy-frontend-dev deploy-frontend-prd \
	invoke-batch-local clean

# =============================================================================
# Terraform Commands
# =============================================================================

init-dev:
	cd terraform && AWS_PROFILE=dev terraform init && (AWS_PROFILE=dev terraform workspace select dev || AWS_PROFILE=dev terraform workspace new dev)

init-prd:
	cd terraform && AWS_PROFILE=prd terraform init && (AWS_PROFILE=prd terraform workspace select prd || AWS_PROFILE=prd terraform workspace new prd)

plan-dev:
	cd terraform && AWS_PROFILE=dev terraform workspace select dev && AWS_PROFILE=dev terraform plan

plan-prd:
	cd terraform && AWS_PROFILE=prd terraform workspace select prd && AWS_PROFILE=prd terraform plan

apply-dev:
	cd terraform && AWS_PROFILE=dev terraform workspace select dev && AWS_PROFILE=dev terraform apply -auto-approve

apply-prd:
	cd terraform && AWS_PROFILE=prd terraform workspace select prd && AWS_PROFILE=prd terraform apply -auto-approve

destroy-dev:
	cd terraform && AWS_PROFILE=dev terraform workspace select dev && AWS_PROFILE=dev terraform destroy -auto-approve

destroy-prd:
	cd terraform && AWS_PROFILE=prd terraform workspace select prd && AWS_PROFILE=prd terraform destroy -auto-approve

output-dev:
	cd terraform && terraform workspace select dev && terraform output

output-prd:
	cd terraform && terraform workspace select prd && terraform output

# =============================================================================
# Lambda Layer Build
# =============================================================================

build-layer:
	cd backend && chmod +x build_layer.sh && ./build_layer.sh

# =============================================================================
# Frontend Commands
# =============================================================================

install-frontend:
	cd frontend && npm install

build-frontend: install-frontend
	cd frontend && npm run build

dev-frontend: install-frontend
	cd frontend && npm run dev

deploy-frontend-dev: build-frontend
	$(eval BUCKET := $(shell cd terraform && AWS_PROFILE=dev terraform workspace select dev > /dev/null && AWS_PROFILE=dev terraform output -raw s3_bucket_name))
	$(eval DIST_ID := $(shell cd terraform && AWS_PROFILE=dev terraform workspace select dev > /dev/null && AWS_PROFILE=dev terraform output -raw cloudfront_distribution_id))
	aws s3 sync frontend/dist s3://$(BUCKET) --delete --profile dev
	aws cloudfront create-invalidation --distribution-id $(DIST_ID) --paths "/*" --profile dev

deploy-frontend-prd: build-frontend
	$(eval BUCKET := $(shell cd terraform && AWS_PROFILE=prd terraform workspace select prd > /dev/null && AWS_PROFILE=prd terraform output -raw s3_bucket_name))
	$(eval DIST_ID := $(shell cd terraform && AWS_PROFILE=prd terraform workspace select prd > /dev/null && AWS_PROFILE=prd terraform output -raw cloudfront_distribution_id))
	aws s3 sync frontend/dist s3://$(BUCKET) --delete --profile prd
	aws cloudfront create-invalidation --distribution-id $(DIST_ID) --paths "/*" --profile prd

# =============================================================================
# Lambda Commands
# =============================================================================



invoke-batch-local:
	cd backend_go && AWS_PROFILE=dev LOCAL_RUN=true go run cmd/batch/main.go

logs-api-dev:
	$(eval FUNC_NAME := $(shell cd terraform && AWS_PROFILE=dev terraform workspace select dev > /dev/null && AWS_PROFILE=dev terraform output -raw api_lambda_function_name))
	aws logs tail /aws/lambda/$(FUNC_NAME) --follow --profile dev

logs-api-prd:
	$(eval FUNC_NAME := $(shell cd terraform && AWS_PROFILE=prd terraform workspace select prd > /dev/null && AWS_PROFILE=prd terraform output -raw api_lambda_function_name))
	aws logs tail /aws/lambda/$(FUNC_NAME) --follow --profile prd

# =============================================================================
# Build Commands
# =============================================================================

# Build Go binaries
build-go:
	mkdir -p .build

	# API Lambda
	cd backend_go && GOOS=linux GOARCH=amd64 go build -o bootstrap cmd/api/main.go
	cd backend_go && zip -j ../.build/api_lambda.zip bootstrap
	rm backend_go/bootstrap

build-frontend:
	cd frontend && npm install
	cd frontend && npm run build

# =============================================================================
# Utility Commands
# =============================================================================

clean:
	rm -rf .build
	rm -rf backend/layer/python
	rm -rf frontend/dist
	rm -rf frontend/node_modules

validate:
	cd terraform && terraform validate

fmt:
	cd terraform && terraform fmt -recursive

# =============================================================================
# Full Deploy Commands
# =============================================================================

# Build everything
build-all: build-go build-frontend

deploy-all-dev: build-all apply-dev deploy-frontend-dev
	@echo "Development environment deployed successfully!"

deploy-all-prd: build-all apply-prd deploy-frontend-prd
	@echo "Production environment deployed successfully!"
