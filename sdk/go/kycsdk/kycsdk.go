package kycsdk

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
)

type SKYClient struct {
	BaseURL    string
	APIKey     string
	httpClient *http.Client
}

type KYCRequest struct {
	Email   string
	IDImage []byte
	Sefile  []byte
}

type KYCResponse struct {
	Success    bool    `json:"success"`
	Verified   bool    `json:"verified"`
	Similarity float32 `json:"similarity"`
	Message    *string `json:"message"`
}

func NewSKYClient(baseURL, apiKey string) *SKYClient {
	return &SKYClient{
		BaseURL:    baseURL,
		APIKey:     apiKey,
		httpClient: &http.Client{},
	}
}

func (c *SKYClient) SubmitKYC(req KYCRequest) (*KYCResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if err := writer.WriteField("email", req.Email); err != nil {
		return nil, err
	}

	part, err := writer.CreateFormFile("id_image", "id_image.jpeg")
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(req.IDImage); err != nil {
		return nil, err
	}

	part, err = writer.CreateFormFile("selfie", "selfie.jpeg")
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(req.Sefile); err != nil {
		return nil, err
	}

	writer.Close()

	request, err := http.NewRequest("POST", c.BaseURL+"/kyc", body)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var kycResp KYCResponse
	if err := json.NewDecoder(resp.Body).Decode(&kycResp); err != nil {
		return nil, err
	}

	return &kycResp, nil
}
