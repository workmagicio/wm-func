package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
	"wm-func/wm_account"
)

var platform = "facebookMarketing"

func main() {
	account := wm_account.GetAccountsWithPlatform(platform)

	accountMap := make(map[int64]wm_account.Account)
	for _, ac := range account {
		accountMap[ac.TenantId] = ac
	}

	//getCampaign(accountMap)
	getAdSet(accountMap)
	//ads := GetAllAds()
	//fmt.Println(ads)
}
func getCampaign(ac map[int64]wm_account.Account) {
	ads := GetCampaignFromFile()
	for i, ad := range ads {
		fmt.Println("==============", i, "============")
		account := ac[ad.TenantId]
		b := Get(ad.CampaignId, account.AccessToken, campaign_fields)
		now := time.Now().Format("2006-01-02 15:04:05")

		saveFirstOrderCustomers(account, []AirbyteData{
			{
				TenantId:            account.TenantId,
				AirbyteRawId:        ad.CampaignId,
				AirbyteData:         b,
				AirbyteExtractedAt:  now,
				AirbyteLoadedAt:     now,
				AirbyteMeta:         "{}",
				AirbyteGenerationId: 0,
				ItemType:            "-",
			},
		})
	}

}
func getAdSet(ac map[int64]wm_account.Account) {
	ads := GetAdSetFromFile()
	for i, ad := range ads {
		fmt.Println("==============", i, "============")
		account := ac[ad.TenantId]
		b := Get(ad.AdSetId, account.AccessToken, adset_fields)
		now := time.Now().Format("2006-01-02 15:04:05")

		saveFirstOrderCustomers(account, []AirbyteData{
			{
				TenantId:            account.TenantId,
				AirbyteRawId:        ad.AdSetId,
				AirbyteData:         b,
				AirbyteExtractedAt:  now,
				AirbyteLoadedAt:     now,
				AirbyteMeta:         "{}",
				AirbyteGenerationId: 0,
				ItemType:            "-",
			},
		})
	}

}
func getAd(ac map[int64]wm_account.Account) {
	ads := GetAsdFromFile()
	for i, ad := range ads {
		fmt.Println("==============", i, "============")
		account := ac[ad.TenantId]
		b := Get(ad.AdId, account.AccessToken, ad_fields)
		now := time.Now().Format("2006-01-02 15:04:05")

		saveFirstOrderCustomers(account, []AirbyteData{
			{
				TenantId:            account.TenantId,
				AirbyteRawId:        ad.AdId,
				AirbyteData:         b,
				AirbyteExtractedAt:  now,
				AirbyteLoadedAt:     now,
				AirbyteMeta:         "{}",
				AirbyteGenerationId: 0,
				ItemType:            "-",
			},
		})
	}

}

func GetAsdFromFile() []Ad {
	b, err := os.ReadFile("/Users/xukai/workspace/workmagic/wm-func/cmd/check_and_add_facebook_setting/ad.json")
	if err != nil {
		panic(err)
	}
	res := []Ad{}
	if err := json.Unmarshal(b, &res); err != nil {
		panic(err)
	}
	return res
}

func GetAdSetFromFile() []AdSet {
	b, err := os.ReadFile("/Users/xukai/workspace/workmagic/wm-func/cmd/check_and_add_facebook_setting/adset.json")
	if err != nil {
		panic(err)
	}
	res := []AdSet{}
	if err := json.Unmarshal(b, &res); err != nil {
		panic(err)
	}
	return res
}
func GetCampaignFromFile() []Campaign {
	b, err := os.ReadFile("/Users/xukai/workspace/workmagic/wm-func/cmd/check_and_add_facebook_setting/campaign.json")
	if err != nil {
		panic(err)
	}
	res := []Campaign{}
	if err := json.Unmarshal(b, &res); err != nil {
		panic(err)
	}
	return res
}
func Get(id, accessToken, fields string) []byte {
	url := "https://graph.facebook.com/v23.0/"
	url = fmt.Sprintf("%s%s?access_token=%s&fields=%s", url, id, accessToken, fields)
	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	if !bytes.Contains(body, []byte(id)) {
		panic("id not found")
	}
	return body
}
