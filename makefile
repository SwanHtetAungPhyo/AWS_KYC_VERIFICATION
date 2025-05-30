.PHONY: build clean deploy delete test invoke-local

STACK_NAME ?= kyc-verification
ENVIRONMENT ?= dev
REGION ?= us-east-1

build:
	@echo "Building Go binary for Lambda..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o main main.go

clean:
	@echo "Cleaning build artifacts..."
	rm -f main
	rm -rf .aws-sam/

validate:
	@echo "Validating SAM template..."
	sam validate

sam-build: clean
	@echo "Building with SAM..."
	sam build

deploy: sam-build
	@echo "Deploying to AWS..."
	sam deploy \
		--stack-name $(STACK_NAME)-$(ENVIRONMENT) \
		--region $(REGION) \
		--parameter-overrides Environment=$(ENVIRONMENT) \
		--capabilities CAPABILITY_IAM \
		--resolve-s3 \
		--confirm-changeset

deploy-guided: sam-build
	@echo "Deploying with guided setup..."
	sam deploy --guided

delete:
	@echo "Deleting stack $(STACK_NAME)-$(ENVIRONMENT)..."
	aws cloudformation delete-stack \
		--stack-name $(STACK_NAME)-$(ENVIRONMENT) \
		--region $(REGION)

start-api: sam-build
	@echo "Starting local API..."
	sam local start-api --port 3000

invoke-local: sam-build
	@echo "Invoking function locally..."
	sam local invoke KYCFunction --event events/api-gateway-event.json

test-local:
	@echo "Testing local API..."
	curl -X POST http://localhost:3000/kyc \
		-F "email=test@example.com" \
		-F "id_image=@./test-images/id.jpg" \
		-F "selfie=@./test-images/selfie.jpg"

outputs:
	@echo "Getting stack outputs..."
	aws cloudformation describe-stacks \
		--stack-name $(STACK_NAME)-$(ENVIRONMENT) \
		--region $(REGION) \
		--query 'Stacks[0].Outputs' \
		--output table

# Get logs
logs:
	@echo "Getting function logs..."
	sam logs --stack-name $(STACK_NAME)-$(ENVIRONMENT) --tail

# Install dependencies
deps:
	@echo "Installing Go dependencies..."
	go mod tidy
	go mod download

# Format code
fmt:
	@echo "Formatting Go code..."
	go fmt ./...

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Package for deployment
package: sam-build
	@echo "Packaging application..."
	sam package \
		--s3-bucket $(S3_BUCKET) \
		--output-template-file packaged.yaml

help:
	@echo "Available commands:"
	@echo "  build        - Build Go binary for Lambda"
	@echo "  clean        - Clean build artifacts"
	@echo "  validate     - Validate SAM template"
	@echo "  sam-build    - Build using SAM"
	@echo "  deploy       - Deploy to AWS"
	@echo "  deploy-guided- Deploy with guided setup"
	@echo "  delete       - Delete the stack"
	@echo "  start-api    - Start local API for testing"
	@echo "  invoke-local - Invoke function locally"
	@echo "  test-local   - Test local API with curl"
	@echo "  outputs      - Get stack outputs"
	@echo "  logs         - Get function logs"
	@echo "  deps         - Install dependencies"
	@echo "  fmt          - Format code"
	@echo "  test         - Run tests"
	@echo "  help         - Show this help"