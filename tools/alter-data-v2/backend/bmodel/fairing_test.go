package bmodel

import (
	"fmt"
	"testing"
)

func TestQuery(t *testing.T) {
	b := GetDataWithPlatform("shopify")
	fmt.Println(b)
}
