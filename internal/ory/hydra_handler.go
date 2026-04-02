package ory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/resoul/studio.go.api/internal/httpx"
)

type HydraHandler struct {
	adminURL string
	client   *http.Client
}

func NewHydraHandler(adminURL string) *HydraHandler {
	return &HydraHandler{
		adminURL: strings.TrimRight(adminURL, "/"),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type hydraError struct {
	StatusCode int
	Message    string
}

func (e *hydraError) Error() string {
	return e.Message
}

type acceptLoginRequest struct {
	LoginChallenge string         `json:"login_challenge" binding:"required"`
	Subject        string         `json:"subject" binding:"required"`
	Remember       bool           `json:"remember"`
	RememberFor    int64          `json:"remember_for"`
	ACR            string         `json:"acr,omitempty"`
	Context        map[string]any `json:"context,omitempty"`
}

type rejectLoginRequest struct {
	LoginChallenge string `json:"login_challenge" binding:"required"`
	Error          string `json:"error,omitempty"`
	ErrorDebug     string `json:"error_debug,omitempty"`
	ErrorHint      string `json:"error_hint,omitempty"`
	ErrorDesc      string `json:"error_description,omitempty"`
	StatusCode     int    `json:"status_code,omitempty"`
}

type acceptConsentRequest struct {
	ConsentChallenge         string         `json:"consent_challenge" binding:"required"`
	GrantScope               []string       `json:"grant_scope"`
	GrantAccessTokenAudience []string       `json:"grant_access_token_audience"`
	Remember                 bool           `json:"remember"`
	RememberFor              int64          `json:"remember_for"`
	Session                  map[string]any `json:"session,omitempty"`
}

type rejectConsentRequest struct {
	ConsentChallenge string `json:"consent_challenge" binding:"required"`
	Error            string `json:"error,omitempty"`
	ErrorDebug       string `json:"error_debug,omitempty"`
	ErrorHint        string `json:"error_hint,omitempty"`
	ErrorDesc        string `json:"error_description,omitempty"`
	StatusCode       int    `json:"status_code,omitempty"`
}

func (h *HydraHandler) GetLoginRequest(c *gin.Context) {
	challenge := strings.TrimSpace(c.Query("login_challenge"))
	if challenge == "" {
		httpx.RespondError(c, http.StatusBadRequest, "invalid_request", "login_challenge is required")
		return
	}

	var out map[string]any
	if err := h.callHydra(
		c,
		http.MethodGet,
		"/admin/oauth2/auth/requests/login",
		url.Values{"login_challenge": []string{challenge}},
		nil,
		&out,
	); err != nil {
		h.respondHydraError(c, err)
		return
	}

	httpx.RespondOK(c, out)
}

func (h *HydraHandler) AcceptLoginRequest(c *gin.Context) {
	var in acceptLoginRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	payload := map[string]any{
		"subject":      in.Subject,
		"remember":     in.Remember,
		"remember_for": in.RememberFor,
	}
	if in.ACR != "" {
		payload["acr"] = in.ACR
	}
	if len(in.Context) > 0 {
		payload["context"] = in.Context
	}

	var out map[string]any
	if err := h.callHydra(
		c,
		http.MethodPut,
		"/admin/oauth2/auth/requests/login/accept",
		url.Values{"login_challenge": []string{in.LoginChallenge}},
		payload,
		&out,
	); err != nil {
		h.respondHydraError(c, err)
		return
	}

	httpx.RespondOK(c, out)
}

func (h *HydraHandler) RejectLoginRequest(c *gin.Context) {
	var in rejectLoginRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	payload := map[string]any{}
	if in.Error != "" {
		payload["error"] = in.Error
	}
	if in.ErrorDebug != "" {
		payload["error_debug"] = in.ErrorDebug
	}
	if in.ErrorHint != "" {
		payload["error_hint"] = in.ErrorHint
	}
	if in.ErrorDesc != "" {
		payload["error_description"] = in.ErrorDesc
	}
	if in.StatusCode > 0 {
		payload["status_code"] = in.StatusCode
	}

	var out map[string]any
	if err := h.callHydra(
		c,
		http.MethodPut,
		"/admin/oauth2/auth/requests/login/reject",
		url.Values{"login_challenge": []string{in.LoginChallenge}},
		payload,
		&out,
	); err != nil {
		h.respondHydraError(c, err)
		return
	}

	httpx.RespondOK(c, out)
}

func (h *HydraHandler) GetConsentRequest(c *gin.Context) {
	challenge := strings.TrimSpace(c.Query("consent_challenge"))
	if challenge == "" {
		httpx.RespondError(c, http.StatusBadRequest, "invalid_request", "consent_challenge is required")
		return
	}

	var out map[string]any
	if err := h.callHydra(
		c,
		http.MethodGet,
		"/admin/oauth2/auth/requests/consent",
		url.Values{"consent_challenge": []string{challenge}},
		nil,
		&out,
	); err != nil {
		h.respondHydraError(c, err)
		return
	}

	httpx.RespondOK(c, out)
}

func (h *HydraHandler) AcceptConsentRequest(c *gin.Context) {
	var in acceptConsentRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	payload := map[string]any{
		"grant_scope":                 in.GrantScope,
		"grant_access_token_audience": in.GrantAccessTokenAudience,
		"remember":                    in.Remember,
		"remember_for":                in.RememberFor,
	}
	if len(in.Session) > 0 {
		payload["session"] = in.Session
	}

	var out map[string]any
	if err := h.callHydra(
		c,
		http.MethodPut,
		"/admin/oauth2/auth/requests/consent/accept",
		url.Values{"consent_challenge": []string{in.ConsentChallenge}},
		payload,
		&out,
	); err != nil {
		h.respondHydraError(c, err)
		return
	}

	httpx.RespondOK(c, out)
}

func (h *HydraHandler) RejectConsentRequest(c *gin.Context) {
	var in rejectConsentRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		httpx.RespondError(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	payload := map[string]any{}
	if in.Error != "" {
		payload["error"] = in.Error
	}
	if in.ErrorDebug != "" {
		payload["error_debug"] = in.ErrorDebug
	}
	if in.ErrorHint != "" {
		payload["error_hint"] = in.ErrorHint
	}
	if in.ErrorDesc != "" {
		payload["error_description"] = in.ErrorDesc
	}
	if in.StatusCode > 0 {
		payload["status_code"] = in.StatusCode
	}

	var out map[string]any
	if err := h.callHydra(
		c,
		http.MethodPut,
		"/admin/oauth2/auth/requests/consent/reject",
		url.Values{"consent_challenge": []string{in.ConsentChallenge}},
		payload,
		&out,
	); err != nil {
		h.respondHydraError(c, err)
		return
	}

	httpx.RespondOK(c, out)
}

func (h *HydraHandler) callHydra(
	c *gin.Context,
	method string,
	apiPath string,
	query url.Values,
	body any,
	out any,
) error {
	if h == nil {
		return &hydraError{StatusCode: http.StatusInternalServerError, Message: "hydra handler is not initialized"}
	}

	baseURL, err := url.Parse(h.adminURL)
	if err != nil {
		return &hydraError{StatusCode: http.StatusInternalServerError, Message: "invalid hydra admin url"}
	}

	baseURL.Path = path.Clean(strings.TrimSuffix(baseURL.Path, "/") + "/" + strings.TrimPrefix(apiPath, "/"))
	baseURL.RawQuery = query.Encode()

	var bodyReader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return &hydraError{StatusCode: http.StatusInternalServerError, Message: "failed to marshal hydra payload"}
		}
		bodyReader = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), method, baseURL.String(), bodyReader)
	if err != nil {
		return &hydraError{StatusCode: http.StatusInternalServerError, Message: "failed to build hydra request"}
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return &hydraError{StatusCode: http.StatusBadGateway, Message: fmt.Sprintf("hydra request failed: %v", err)}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &hydraError{StatusCode: http.StatusBadGateway, Message: "failed to read hydra response"}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := strings.TrimSpace(string(respBody))
		if msg == "" {
			msg = fmt.Sprintf("hydra returned status %d", resp.StatusCode)
		}
		return &hydraError{StatusCode: resp.StatusCode, Message: msg}
	}

	if out == nil || len(respBody) == 0 {
		return nil
	}

	if err := json.Unmarshal(respBody, out); err != nil {
		return &hydraError{StatusCode: http.StatusBadGateway, Message: "failed to parse hydra response"}
	}

	return nil
}

func (h *HydraHandler) respondHydraError(c *gin.Context, err error) {
	var he *hydraError
	if ok := asHydraError(err, &he); ok {
		httpx.RespondError(c, he.StatusCode, "hydra_error", he.Message)
		return
	}

	httpx.RespondError(c, http.StatusBadGateway, "hydra_error", err.Error())
}

func asHydraError(err error, target **hydraError) bool {
	if err == nil || target == nil {
		return false
	}

	v, ok := err.(*hydraError)
	if !ok {
		return false
	}

	*target = v
	return true
}
