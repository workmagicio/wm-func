package controller

import "testing"

func TestController_GetData(t *testing.T) {
	name := "amazon vendor partner"
	c := NewController(name)
	c.Cac()
}
