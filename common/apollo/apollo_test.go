package apollo

import (
	"fmt"
	"testing"

	"github.com/philchia/agollo/v4"
)

func TestApollo(t *testing.T) {
	_ = GetInstance()

	res := agollo.GetString("spring.datasource.api.url", agollo.WithNamespace("datasource"))
	fmt.Println(res)
}

func TestBuildFacebookMarketingDma(t *testing.T) {
}
