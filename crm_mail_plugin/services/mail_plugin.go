package services

import (
	"encoding/json"
	"io"
	"net/http"

	"sumeru/core/orm"
	"sumeru/core/server/router"
)

func init() {
	router.Register(http.MethodPost, "/crm/mail_plugin/lead", router.AuthSession, mailPluginLeadHandler)
}

type mailPluginLeadRequest struct {
	Subject string `json:"subject"`
	Email   string `json:"email"`
	Body    string `json:"body"`
}

type mailPluginLeadResponse struct {
	LeadID int    `json:"lead_id"`
	Name   string `json:"name"`
}

func mailPluginLeadHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	var req mailPluginLeadRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	name := req.Subject
	if name == "" {
		name = req.Email
	}
	if name == "" {
		name = "Email Lead"
	}
	bypass := orm.ContextWithBypass(r.Context(), true)
	leadModel, ok := orm.Registry["crm.lead"]
	if !ok {
		http.Error(w, "crm.lead missing", http.StatusInternalServerError)
		return
	}
	leadID, err := orm.Create(bypass, leadModel, map[string]interface{}{
		"name":        name,
		"type":        "lead",
		"email_from":  req.Email,
		"description": req.Body,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(mailPluginLeadResponse{LeadID: leadID, Name: name})
}
