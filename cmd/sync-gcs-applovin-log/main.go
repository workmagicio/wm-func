package main

import (
	"log"
	"time"
)

func main() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		for tenantId := range tenantAccountMap {
			p := NewApplovin(tenantId)
			p.Sync()
			log.Printf("Sync Applovin log for tenant %d completed", tenantId)
		}
		log.Println("--------------------------------")
		log.Printf("Sync Applovin log for all tenants completed")
	}

}
