package controller

import (
	"testing"
	"wm-func/tools/alter-data-v2/backend"
)

func TestGetAlterDataWithPlatform(t *testing.T) {
	GetAlterDataWithPlatform(false, backend.PLATFORM_GOOGLE)
}
