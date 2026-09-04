package services

import (
	"sumeru/core/orm"
)

// IAPService describes a registered IAP provider stub.
type IAPService struct {
	Name  string
	Label string
}

// ServiceRegistry maps service names to stub provider metadata.
var ServiceRegistry = map[string]IAPService{}

// RegisterService adds a stub IAP service to the registry.
func RegisterService(service IAPService) {
	if service.Name == "" {
		return
	}
	ServiceRegistry[service.Name] = service
}

func init() {
	RegisterService(IAPService{Name: "clearbit", Label: "Lead Enrichment"})
	RegisterService(IAPService{Name: "reveal", Label: "Lead Mining"})
	orm.RegisterObjectAction("crm.iap.lead.mining.request", "action_mine_leads", actionMineLeads)
}
