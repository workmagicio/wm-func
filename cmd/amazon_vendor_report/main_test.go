package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestCount(t *testing.T) {
	md := DocumentResponse{}

	b, err := os.ReadFile("/Users/xukai/Downloads/685de5e5-24de-4c05-aac7-b87b29c25dac.amzn1.tortuga.4.na.T318HIE8CXXRPO.json")
	if err != nil {
		panic(err)
	}

	json.Unmarshal(b, &md)

	cnt := 0
	sales := 0.0
	for _, metrics := range md.Metrics {
		cnt += metrics.Metrics.ShippedOrders.ShippedUnitsWithRevenue.Units
		sales += metrics.Metrics.ShippedOrders.ShippedUnitsWithRevenue.Value.Amount

	}

	fmt.Println(cnt, sales)
}

func TestGetdocumentResult(t *testing.T) {
	url := "https://tortuga-prod-na.s3-external-1.amazonaws.com/2ddbe860-5f20-4a1c-8f17-5d773c8504e7.amzn1.tortuga.4.na.T1QO94ZI8JN4A3?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Date=20250620T035254Z&X-Amz-SignedHeaders=host&X-Amz-Expires=300&X-Amz-Credential=AKIA5U6MO6RADQRQYCSG%2F20250620%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Signature=a975e6b945f9f72514415210cf8258aecd4a251822e9018b2f006aecad3ca539"
	getDocumentResult(url)
}

func TestStreamSlice(t *testing.T) {
	start := time.Now().Add(time.Hour * 24 * 180 * -1)
	var streamSlice []StreamSlice

	if start.Add(time.Hour * 24 * 30).After(time.Now()) {
		start = start.Add(time.Hour * 24 * 30 * -1)
	}
	now := time.Now()

	r := now.Sub(start)
	fmt.Println(r.Hours())
	for start.Before(now) && now.Sub(start).Hours() > 1 {
		end := start.Add(time.Hour * 24 * 30)
		if end.After(now) {
			end = now
		}

		streamSlice = append(streamSlice, StreamSlice{
			Start: start,
			End:   end,
		})

		if end.Equal(now) {
			break
		}
		start = end
	}
	fmt.Println(streamSlice)

}
