package main

import (
	"fmt"
	"os"

	"github.com/SwanHtetAungPhyo/swan-kyc-sdk/kycsdk"
)

func main() {
	idImage, err := os.ReadFile("./test/passort.jpeg")
	if err != nil {
		fmt.Printf("Error reading id_image: %v\n", err)
		return
	}
	selfie, err := os.ReadFile("./test/selfile.jpeg")
	if err != nil {
		fmt.Printf("Error reading selfie: %v\n", err)
		return
	}

	client := kycsdk.NewSKYClient("https://aws-kyc-verification.onrender.com")
	req := kycsdk.KYCRequest{
		Email:   "john.doe@example.com",
		IDImage: idImage,
		Sefile:  selfie,
	}

	resp, err := client.SubmitKYC(req)
	if err != nil {
		fmt.Printf("Error submitting KYC: %v\n", err)
		return
	}

	fmt.Printf("Success: %v\n", resp.Success)
	fmt.Printf("Verified: %v\n", resp.Verified)
	fmt.Printf("Similarity: %f\n", resp.Similarity)
	if resp.Message != nil {
		fmt.Printf("Message: %s\n", *resp.Message)
	}
}
