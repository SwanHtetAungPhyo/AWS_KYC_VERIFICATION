#!/bin/bash

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
