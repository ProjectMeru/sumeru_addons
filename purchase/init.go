package purchase

import (
	"log"

	_ "sumeru_addons/purchase/models"
)

func init() {
	log.Println("Purchase Addon Loaded")
}
