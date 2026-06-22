package crm

import (
	"log"

	_ "sumeru_addons/crm/models"
)

func init() {
	log.Println("CRM Addon Loaded")
}
