package crm

import (
	"log"

	_ "sumeru_addons/crm/models"
	_ "sumeru_addons/crm/services"
	_ "sumeru_addons/crm/wizard"
)

func init() {
	log.Println("CRM Addon Loaded")
}
