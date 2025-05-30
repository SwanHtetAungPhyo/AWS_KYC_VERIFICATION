#!/bin/bash

FUNCTION_NAME="kyc-verification"
ZIP_NAME="function.zip"
ROLE_ARN="arn:aws:iam::162047532564:role/lambda-execution-role"
ARCHITECTURE="arm64"

set -e

echo "🔨 Building Go binary for Linux..."
GOOS=linux GOARCH=arm64 go build -o bootstrap main.go

echo "📦 Packaging into $ZIP_NAME..."
zip -q $ZIP_NAME bootstrap

echo "🔍 Checking if Lambda function '$FUNCTION_NAME' exists..."
if aws lambda get-function --function-name "$FUNCTION_NAME" >/dev/null 2>&1; then
  echo "🔁 Updating existing Lambda function..."
  aws lambda update-function-code \
    --function-name "$FUNCTION_NAME" \
    --zip-file fileb://$ZIP_NAME
else
  echo "🚀 Creating new Lambda function..."
  aws lambda create-function \
    --function-name "$FUNCTION_NAME" \
    --runtime provided.al2 \
    --handler bootstrap \
    --role "$ROLE_ARN" \
    --zip-file fileb://$ZIP_NAME \
    --architectures "$ARCHITECTURE" \
    --tracing-config Mode=Active
fi

echo "✅ Deployment complete!"