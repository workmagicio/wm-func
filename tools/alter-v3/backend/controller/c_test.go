package controller

import "testing"

func TestController_GetData(t *testing.T) {
	name := "metaAds"
	c := NewController(name)
	c.Cac()
}
