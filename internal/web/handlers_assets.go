package web

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/weiyong1024/clawsandbox/internal/state"
)

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *Server) loadAssets() (*state.AssetStore, error) {
	return state.LoadAssets()
}

// --- Model Asset Handlers ---

func (s *Server) handleListModelAssets(w http.ResponseWriter, r *http.Request) {
	store, err := s.loadAssets()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": store.ListModels()})
}

type createModelRequest struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	APIKey   string `json:"api_key"`
	Model    string `json:"model"`
}

func (s *Server) handleCreateModelAsset(w http.ResponseWriter, r *http.Request) {
	var req createModelRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Provider == "" || req.APIKey == "" || req.Model == "" {
		writeError(w, http.StatusBadRequest, "provider, api_key, and model are required")
		return
	}

	store, err := s.loadAssets()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	name := req.Name
	if name == "" {
		name = fmt.Sprintf("%s %s", providerDisplayName(req.Provider), req.Model)
	}

	asset := &state.ModelAsset{
		ID:        generateID(),
		Name:      name,
		Provider:  req.Provider,
		APIKey:    req.APIKey,
		Model:     req.Model,
		Validated: true, // Only saved after validation passes
	}

	store.AddModel(asset)
	if err := store.SaveAssets(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": asset})
}

func (s *Server) handleUpdateModelAsset(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req createModelRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	store, err := s.loadAssets()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	existing := store.GetModel(id)
	if existing == nil {
		writeError(w, http.StatusNotFound, "model asset not found")
		return
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Provider != "" {
		existing.Provider = req.Provider
	}
	if req.APIKey != "" {
		existing.APIKey = req.APIKey
	}
	if req.Model != "" {
		existing.Model = req.Model
	}
	existing.Validated = true

	if !store.UpdateModel(existing) {
		writeError(w, http.StatusNotFound, "model asset not found")
		return
	}
	if err := store.SaveAssets(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": existing})
}

func (s *Server) handleDeleteModelAsset(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	store, err := s.loadAssets()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !store.RemoveModel(id) {
		writeError(w, http.StatusNotFound, "model asset not found")
		return
	}
	if err := store.SaveAssets(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]string{"status": "deleted"}})
}

type testModelRequest struct {
	Provider string `json:"provider"`
	APIKey   string `json:"api_key"`
	Model    string `json:"model"`
}

func (s *Server) handleTestModelAsset(w http.ResponseWriter, r *http.Request) {
	var req testModelRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	err := ValidateModelKey(req.Provider, req.APIKey, req.Model)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"data": map[string]any{"valid": false, "error": err.Error()},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{"valid": true},
	})
}

// --- Channel Asset Handlers ---

func (s *Server) handleListChannelAssets(w http.ResponseWriter, r *http.Request) {
	store, err := s.loadAssets()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": store.ListChannels()})
}

type createChannelRequest struct {
	Name      string `json:"name"`
	Channel   string `json:"channel"`
	Token     string `json:"token"`
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
}

func (s *Server) handleCreateChannelAsset(w http.ResponseWriter, r *http.Request) {
	var req createChannelRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Channel == "" {
		writeError(w, http.StatusBadRequest, "channel is required")
		return
	}
	if req.Channel == "lark" {
		if req.AppID == "" || req.AppSecret == "" {
			writeError(w, http.StatusBadRequest, "app_id and app_secret are required for Lark")
			return
		}
	} else if req.Token == "" {
		writeError(w, http.StatusBadRequest, "token is required")
		return
	}

	store, err := s.loadAssets()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	name := req.Name
	if name == "" {
		name = fmt.Sprintf("%s Bot", channelDisplayName(req.Channel))
	}

	asset := &state.ChannelAsset{
		ID:        generateID(),
		Name:      name,
		Channel:   req.Channel,
		Token:     req.Token,
		AppID:     req.AppID,
		AppSecret: req.AppSecret,
		Validated: true,
	}

	store.AddChannel(asset)
	if err := store.SaveAssets(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": asset})
}

func (s *Server) handleUpdateChannelAsset(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req createChannelRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	store, err := s.loadAssets()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	existing := store.GetChannel(id)
	if existing == nil {
		writeError(w, http.StatusNotFound, "channel asset not found")
		return
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Channel != "" {
		existing.Channel = req.Channel
	}
	if req.Token != "" {
		existing.Token = req.Token
	}
	if req.AppID != "" {
		existing.AppID = req.AppID
	}
	if req.AppSecret != "" {
		existing.AppSecret = req.AppSecret
	}
	existing.Validated = true

	if !store.UpdateChannel(existing) {
		writeError(w, http.StatusNotFound, "channel asset not found")
		return
	}
	if err := store.SaveAssets(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": existing})
}

func (s *Server) handleDeleteChannelAsset(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	store, err := s.loadAssets()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !store.RemoveChannel(id) {
		writeError(w, http.StatusNotFound, "channel asset not found")
		return
	}
	if err := store.SaveAssets(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]string{"status": "deleted"}})
}

type testChannelRequest struct {
	Channel   string `json:"channel"`
	Token     string `json:"token"`
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
}

func (s *Server) handleTestChannelAsset(w http.ResponseWriter, r *http.Request) {
	var req testChannelRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	err := ValidateChannelToken(req.Channel, req.Token, req.AppID, req.AppSecret)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"data": map[string]any{"valid": false, "error": err.Error()},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{"valid": true},
	})
}

func providerDisplayName(provider string) string {
	switch provider {
	case "anthropic":
		return "Anthropic"
	case "openai":
		return "OpenAI"
	case "google":
		return "Google"
	case "deepseek":
		return "DeepSeek"
	default:
		return provider
	}
}

func channelDisplayName(channel string) string {
	switch channel {
	case "telegram":
		return "Telegram"
	case "discord":
		return "Discord"
	case "slack":
		return "Slack"
	case "lark":
		return "Lark"
	default:
		return channel
	}
}
