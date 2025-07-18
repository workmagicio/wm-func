package state

import (
	"fmt"
	"testing"
)

func TestGetSyncInfo(t *testing.T) {
	//133944,1011512027331849,amazonAds,ProductStream

	res := GetSyncInfo(133944, "1011512027331849", "amazonAds", "ProductStream")
	fmt.Println(res)

}
