package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Result struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      Data   `json:"data"`
	RequestId string `json:"request_id"`
}
type Data struct {
	AdvertiserId          string        `json:"advertiser_id"`
	AffiliatePostsEnabled bool          `json:"affiliate_posts_enabled"`
	AgeGroups             []string      `json:"age_groups"`
	Budget                int           `json:"budget"`
	CampaignId            string        `json:"campaign_id"`
	CampaignName          string        `json:"campaign_name"`
	CustomAnchorVideoList []interface{} `json:"custom_anchor_video_list"`
	DeepBidType           string        `json:"deep_bid_type"`
	IdentityList          []struct {
		IdentityId   string `json:"identity_id"`
		IdentityType string `json:"identity_type"`
		StoreId      string `json:"store_id"`
	} `json:"identity_list"`
	ItemGroupIds             []string      `json:"item_group_ids"`
	ItemList                 []interface{} `json:"item_list"`
	LocationIds              []string      `json:"location_ids"`
	OperationStatus          string        `json:"operation_status"`
	OptimizationGoal         string        `json:"optimization_goal"`
	Placements               []string      `json:"placements"`
	ProductSpecificType      string        `json:"product_specific_type"`
	ProductVideoSpecificType string        `json:"product_video_specific_type"`
	RoasBid                  float64       `json:"roas_bid"`
	RoiProtectionEnabled     bool          `json:"roi_protection_enabled"`
	ScheduleEndTime          string        `json:"schedule_end_time"`
	ScheduleStartTime        string        `json:"schedule_start_time"`
	ScheduleType             string        `json:"schedule_type"`
	ShoppingAdsType          string        `json:"shopping_ads_type"`
	StoreAuthorizedBcId      string        `json:"store_authorized_bc_id"`
	StoreId                  string        `json:"store_id"`
}

func RequestById(advertiserId, campaignId, accessToken string) *Data {
	url := "https://business-api.tiktok.com/open_api/v1.3/campaign/gmv_max/info/?advertiser_id=%s&campaign_id=%s"
	url = fmt.Sprintf(url, advertiserId, campaignId)
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return nil
	}
	req.Header.Add("Access-Token", accessToken)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	var result Result
	if err = json.Unmarshal(body, &result); err != nil {
		panic(err)
	}
	return &result.Data
}
