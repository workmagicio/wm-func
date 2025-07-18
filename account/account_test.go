package account

import (
	"fmt"
	"testing"
)

func TestGetShopifyAccount(t *testing.T) {
	res := GetShopifyAccount()
	fmt.Println(res)
}
