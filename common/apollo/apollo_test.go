package apollo

import (
	"fmt"
	"github.com/philchia/agollo/v4"
	"testing"
)

func init() {
	Init()
}

func TestApollo(t *testing.T) {
	Init()
	res := agollo.GetString("application.service.AWS_ACCESS_KEY_ID")

	fmt.Println(res)
}

func TestBuildFacebookMarketingDma(t *testing.T) {
}
