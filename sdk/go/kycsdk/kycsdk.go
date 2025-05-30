package kycsdk

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
)

type KYCRequest struct {
	Email   string `json:"email"`
	IDImage []byte `json:"id_image"` // Byte array for ID image
	Sefile  []byte `json:"selfie"`   // Byte array for selfie image
}
type KYCResponse struct {
	Success    bool    `json:"success"`
	Verified   bool    `json:"verified"`
	Similarity float64 `json:"similarity"`
	Message    *string `json:"message,omitempty"`
}

// SKYClient is a client for interacting with the KYC API.
type SKYClient struct {
	baseURL string
	client  *http.Client
}

// NewSKYClient creates a new SKYClient with the specified base URL.
// If baseURL is empty, it defaults to "https://default-api-url".
func NewSKYClient(baseURL string) *SKYClient {
	if baseURL == "" {
		baseURL = "https://default-api-url"
	}
	return &SKYClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// SubmitKYC sends a KYC verification request with the provided email and image data.
func (c *SKYClient) SubmitKYC(req KYCRequest) (*KYCResponse, error) {
	// Create a buffer for the multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add email field
	if err := writer.WriteField("email", req.Email); err != nil {
		return nil, err
	}

	// Add id_image file
	idImagePart, err := writer.CreateFormFile("id_image", "id_image.jpeg")
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(idImagePart, bytes.NewReader(req.IDImage)); err != nil {
		return nil, err
	}

	// Add selfie file
	selfiePart, err := writer.CreateFormFile("selfie", "selfie.jpeg")
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(selfiePart, bytes.NewReader(req.Sefile)); err != nil {
		return nil, err
	}

	// Close the multipart writer
	if err := writer.Close(); err != nil {
		return nil, err
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", c.baseURL+"/kyc", body)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read and parse response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var kycResp KYCResponse
	if err := json.Unmarshal(respBody, &kycResp); err != nil {
		return nil, err
	}

	return &kycResp, nil
}
