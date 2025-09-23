package main

import (
	"log"
	"time"
)

func main() {
	run()
	ticker := time.NewTicker(10 * time.Minute)
	for range ticker.C {
		run()
	}

}

func run() {
	for tenantId := range tenantAccountMap {
		p := NewApplovin(tenantId)
		p.Sync()
		log.Printf("Sync Applovin log for tenant %d completed", tenantId)
	}
	log.Println("--------------------------------")
	log.Printf("Sync Applovin log for all tenants completed")

}
