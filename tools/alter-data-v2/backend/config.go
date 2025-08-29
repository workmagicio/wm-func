package backend

const (
	PLATFORM_GOOGLE           = "googleAds"
	PLATFORM_FACEBOOK         = "facebookMarketing"
	PLATFORN_TIKTIK_MARKETING = "tiktokMarketing"

	ADS_PLATFORM_GOOGLE = "Google"
)

var PlatformMap = map[string]string{
	PLATFORM_GOOGLE: ADS_PLATFORM_GOOGLE,
}
