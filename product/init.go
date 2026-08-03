package product

import (
	"log"

	_ "sumeru_addons/product/models"
)

func init() {
	log.Println("Product Addon Loaded")
}
