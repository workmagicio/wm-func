package cac

import (
	"testing"
	"wm-func/tools/alter-data-v2/backend"
)

func TestCac(t *testing.T) {
	GetAlterDataWithPlatform(backend.PLATFORM_GOOGLE, true)
}
