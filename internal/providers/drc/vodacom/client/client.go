package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/abubakar508/taas/internal/domain/credentials"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Meta struct {
	BaseURL               string `json:"base_url"`
	BasicToken            string `json:"basic_token"`
	OfferUsername         string `json:"offer_username"`
	OfferPassword         string `json:"offer_password"`
	InsecureSkipTLSVerify bool   `json:"insecure_skip_tls_verify"`

	// SOAP / airtime fields
	AirtimeSOAPURL      string `json:"airtime_soap_url"`
	ThirdPartyID        string `json:"third_party_id"`
	ThirdPartyPassword  string `json:"third_party_password"`
	InitiatorIdentifier string `json:"initiator_identifier"`
	SecurityCredential  string `json:"security_credential"`
	ShortCode           string `json:"short_code"`
	CallerType          string `json:"caller_type"`
	KeyOwner            string `json:"key_owner"`
	CommandID           string `json:"command_id"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   string `json:"expires_in"`
}

type Client struct {
	credRepo credentials.Repository
	http     *http.Client
}

func New(pool *pgxpool.Pool) *Client {
	return &Client{
		credRepo: credentials.NewRepository(pool),
		http: &http.Client{
			Timeout: 25 * time.Second,
		},
	}
}

func (c *Client) LoadMeta(ctx context.Context, partnerID string) (*Meta, error) {
	cred, err := c.credRepo.FindActiveByPartnerCountryProvider(ctx, partnerID, "drc", "vodacom")
	if err != nil {
		return nil, fmt.Errorf("provider credentials not configured")
	}

	meta := &Meta{}
	if len(cred.Meta) > 0 {
		if err := json.Unmarshal(cred.Meta, meta); err != nil {
			return nil, fmt.Errorf("invalid provider credentials")
		}
	}

	if meta.BaseURL == "" {
		//the provided url will be of this format  "https://testenv.m-pesa.vodacom.cd:PORT" unless changed. Ensure you use  a test url before production
		meta.BaseURL = "https://testenv.m-pesa.vodacom.cd:5443"
	}

	if meta.AirtimeSOAPURL == "" {
		//the provided url will be of this format  "https://IP_ADDRESS:PORT/payment/services/SYNCAPIRequestMgrService" unless changed. Ensure you use  a test url before production
		meta.AirtimeSOAPURL = "https://192.168.1.100:30002/payment/services/SYNCAPIRequestMgrService"
	}

	if meta.CallerType == "" {
		meta.CallerType = "2"
	}

	if meta.KeyOwner == "" {
		meta.KeyOwner = "1"
	}

	if meta.CommandID == "" {
		meta.CommandID = "InitTrans_2108"
	}

	return meta, nil
}

func (c *Client) HTTP(meta *Meta) *http.Client {
	if !meta.InsecureSkipTLSVerify {
		return c.http
	}

	return &http.Client{
		Timeout: 25 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}

func (c *Client) GetAccessToken(ctx context.Context, meta *Meta) (string, error) {
	url := meta.BaseURL + "/v1/token/generate?grant_type=client_credentials"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("authorization", "Basic "+meta.BasicToken)
	req.Header.Set("content-type", "application/json")
	req.Header.Set("user-agent", "taas-go-client")

	resp, err := c.HTTP(meta).Do(req)
	if err != nil {
		return "", fmt.Errorf("token request failed")
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("token request failed: %s", string(body))
	}

	var tr TokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return "", fmt.Errorf("token response parse failed")
	}
	if tr.AccessToken == "" {
		return "", fmt.Errorf("access token missing")
	}

	return tr.AccessToken, nil
}

func (c *Client) PostJSON(ctx context.Context, meta *Meta, url string, token string, payload any) ([]byte, int, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build provider request")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build provider request")
	}

	req.Header.Set("authorization", "Bearer "+token)
	req.Header.Set("content-type", "application/json")
	req.Header.Set("user-agent", "taas-go-client")

	resp, err := c.HTTP(meta).Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("provider request failed")
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	return raw, resp.StatusCode, nil
}

func (c *Client) PostSOAP(ctx context.Context, meta *Meta, payload []byte) ([]byte, int, error) {
	url := meta.AirtimeSOAPURL
	if url == "" {
		url = "https://41.78.195.155:30002/payment/services/SYNCAPIRequestMgrService"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build soap request")
	}

	req.Header.Set("content-type", "text/xml; charset=utf-8")
	req.Header.Set("accept", "text/xml")
	req.Header.Set("user-agent", "taas-go-client")

	resp, err := c.HTTP(meta).Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("soap request failed")
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	return raw, resp.StatusCode, nil
}

func MarshalSOAP(v any) ([]byte, error) {
	b, err := xml.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), b...), nil
}
