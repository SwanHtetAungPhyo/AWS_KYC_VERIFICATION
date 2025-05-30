#!/bin/bash

<<<<<<< HEAD
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
=======
# === CONFIGURATION ===
FUNCTION_NAME="my-go-lambda"
ROLE_ARN="arn:aws:iam::162047532564:role/lambda-execution-role"
RUNTIME="provided.al2023"
ZIP_FILE="function.zip"
BINARY_NAME="bootstrap"

# === BUILD ===
echo "🔨 Building Go binary for Linux..."
GOOS=linux GOARCH=amd64 go build -o $BINARY_NAME main.go || { echo "Build failed"; exit 1; }

# === ZIP ===
echo "📦 Zipping the binary..."
zip -j $ZIP_FILE $BINARY_NAME > /dev/null

# === CHECK IF FUNCTION EXISTS ===
echo "🔍 Checking if Lambda function '$FUNCTION_NAME' exists..."
aws lambda get-function --function-name "$FUNCTION_NAME" > /dev/null 2>&1
EXISTS=$?

if [ $EXISTS -eq 0 ]; then
  # === UPDATE FUNCTION ===
  echo "♻️ Updating existing Lambda function '$FUNCTION_NAME'..."
  aws lambda update-function-code \
    --function-name "$FUNCTION_NAME" \
    --zip-file "fileb://$ZIP_FILE"
else
  # === CREATE FUNCTION ===
  echo "🚀 Creating new Lambda function '$FUNCTION_NAME'..."
  aws lambda create-function \
    --function-name "$FUNCTION_NAME" \
    --runtime "$RUNTIME" \
    --handler "$BINARY_NAME" \
    --zip-file "fileb://$ZIP_FILE" \
    --role "$ROLE_ARN"
fi

echo "✅ Done!"
>>>>>>> eed52768e0491698ce2a6003b4760324695cc1c5
