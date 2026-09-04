package services

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"sumeru/core/orm"
	"sumeru/core/server/router"
)

func init() {
	router.Register(http.MethodPost, "/website/crm/form/submit", router.AuthNone, websiteCrmFormSubmitHandler)
}

type websiteCrmFormSubmitRequest struct {
	FormID     int    `json:"form_id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Message    string `json:"message"`
	CampaignID int    `json:"campaign_id"`
	MediumID   int    `json:"medium_id"`
	SourceID   int    `json:"source_id"`
}

type websiteCrmFormSubmitResponse struct {
	LeadID int    `json:"lead_id"`
	Name   string `json:"name"`
}

func websiteCrmFormSubmitHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	var req websiteCrmFormSubmitRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}
	bypass := orm.ContextWithBypass(r.Context(), true)
	if req.FormID > 0 {
		if _, err := orm.SearchOne(bypass, "website.crm.form", map[string]interface{}{
			"id":     req.FormID,
			"active": true,
		}); err != nil {
			http.Error(w, "form not found", http.StatusNotFound)
			return
		}
	}
	leadModel, ok := orm.Registry["crm.lead"]
	if !ok {
		http.Error(w, "crm.lead missing", http.StatusInternalServerError)
		return
	}
	leadVals := map[string]interface{}{
		"name":        req.Name,
		"type":        "lead",
		"email_from":  req.Email,
		"phone":       req.Phone,
		"description": req.Message,
	}
	if req.FormID > 0 {
		if form, err := orm.SearchOne(bypass, "website.crm.form", map[string]interface{}{"id": req.FormID}); err == nil {
			if teamID, ok := orm.CoerceInt64(form["team_id"]); ok && teamID > 0 {
				leadVals["team_id"] = teamID
			}
		}
	}
	if req.CampaignID > 0 {
		leadVals["campaign_id"] = req.CampaignID
	}
	if req.MediumID > 0 {
		leadVals["medium_id"] = req.MediumID
	}
	if req.SourceID > 0 {
		leadVals["source_id"] = req.SourceID
	}
	leadID, err := orm.Create(bypass, leadModel, leadVals)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(websiteCrmFormSubmitResponse{LeadID: leadID, Name: req.Name})
}

// FormSubmitURL returns the public submission endpoint for a form (stub helper).
func FormSubmitURL(formID int) string {
	return "/website/crm/form/submit?form_id=" + strconv.Itoa(formID)
}
