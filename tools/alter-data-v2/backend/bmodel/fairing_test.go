package bmodel

import (
	"fmt"
	"testing"
)

func TestQuery(t *testing.T) {
	b := GetSingleDataWithPlatform("amazonVendorPartner")
	fmt.Println(b)
}
