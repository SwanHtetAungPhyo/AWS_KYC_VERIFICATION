KYC Verification API
This is a Go-based REST API built with the Fiber framework for performing Know Your Customer (KYC) verification. It uses AWS Textract to analyze ID documents and AWS Rekognition to perform facial recognition by comparing an ID image with a selfie. The API accepts multipart form data containing an email, an ID image, and a selfie, and returns a verification result based on facial similarity.
Features

Validates ID documents using AWS Textract.
Detects and validates faces in selfies using AWS Rekognition.
Compares faces between ID and selfie with a minimum similarity threshold of 70%.
Logs verification results and errors using Logrus.
Supports Docker for easy deployment.
Handles multipart form data for file uploads.

Prerequisites

Go: Version 1.22 or later.
AWS Account: With access to Textract and Rekognition services.
Docker: Optional, for containerized deployment.
Environment Variables:
AWS_ACCESS_KEY_ID: AWS access key.
AWS_SECRET_ACCESS_KEY: AWS secret key.
AWS_REGION: AWS region (e.g., us-east-1).
PORT: Optional, defaults to 3000.



Installation

Clone the Repository:
git clone <repository-url>
cd kyc-api


Set Up Dependencies:Ensure the go.mod file includes:
module kyc-api

go 1.22

require (
    github.com/aws/aws-sdk-go-v2 v1.30.3
    github.com/aws/aws-sdk-go-v2/config v1.27.27
    github.com/aws/aws-sdk-go-v2/credentials v1.17.27
    github.com/aws/aws-sdk-go-v2/service/rekognition v1.40.2
    github.com/aws/aws-sdk-go-v2/service/textract v1.34.2
    github.com/gofiber/fiber/v2 v2.52.5
    github.com/sirupsen/logrus v1.9.3
)

Run:
go mod tidy


Set Environment Variables:
export AWS_ACCESS_KEY_ID=<your_access_key>
export AWS_SECRET_ACCESS_KEY=<your_secret_key>
export AWS_REGION=us-east-1
export PORT=3000


Run the Application:
go run main.go

The API will be available at http://localhost:3000.


Docker Setup

Build the Docker Image:
docker build -t kyc-api .


Run the Container:
docker run -p 3000:3000 \
  -e AWS_ACCESS_KEY_ID=<your_access_key> \
  -e AWS_SECRET_ACCESS_KEY=<your_secret_key> \
  -e AWS_REGION=us-east-1 \
  kyc-api

The API will be accessible at http://localhost:3000.


API Endpoint
POST /kyc
Performs KYC verification by processing an email, ID image, and selfie.
Request

Content-Type: multipart/form-data
Form Fields:
email (string, required): User's email address.
id_image (file, required): ID document image (e.g., PNG, JPG).
selfie (file, required): Selfie image for facial comparison.



Example:
curl -X POST http://localhost:3000/kyc \
  -F "email=user@example.com" \
  -F "id_image=@/path/to/id.jpg" \
  -F "selfie=@/path/to/selfie.jpg"

Response

Content-Type: application/json
Status Codes:
200 OK: Verification completed.
400 Bad Request: Invalid or missing input.
500 Internal Server Error: Server-side error.



Success Response:
{
  "success": true,
  "verified": true,
  "similarity": 85.5,
  "message": "KYC verification completed"
}

Error Response:
{
  "success": false,
  "error": "Email is required"
}

Verification Process

Input Validation: Checks for valid email and non-empty image files.
ID Analysis: Uses AWS Textract to validate the ID document.
Face Detection: Uses AWS Rekognition to detect exactly one face in the selfie and validate its quality (confidence ≥ 90%, brightness ≥ 50, sharpness ≥ 50).
Face Comparison: Compares faces between the ID and selfie, requiring a similarity score ≥ 70% for verification.
Logging: Logs all steps and errors using Logrus.

Error Handling

400 Bad Request: Missing or invalid email, missing files, or invalid form data.
500 Internal Server Error: File reading errors, AWS service failures, or verification issues (e.g., no face matches, poor image quality).
Errors are logged with detailed context for debugging.

Project Structure
kyc-api/
├── main.go          # Main application code
├── go.mod          # Go module dependencies
├── go.sum          # Dependency checksums
├── Dockerfile       # Docker configuration
└── README.md       # This file

Testing
To test the API:

Ensure AWS credentials are valid and have permissions for Textract and Rekognition.
Use a tool like curl or Postman to send a POST /kyc request with valid form data.
Check logs for detailed debugging information.

Contributing
Contributions are welcome! Please submit a pull request or open an issue for bugs, features, or improvements.
License
This project is licensed under the MIT License.
