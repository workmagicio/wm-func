package main

import (
	"bufio"
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"log"
	"strings"
	"wm-func/common/cache"
)

type Applovin struct {
	cache     *cache.S3Cache
	TenantId  int64
	Bucket    string
	Prefix    string
	GcsClient *storage.Client
}

func NewApplovin(tenantId int64) *Applovin {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsJSON(credentials))
	if err != nil {
		panic(err)
	}

	return &Applovin{
		cache:     cache.LoadS3Cache(tenantId),
		TenantId:  tenantId,
		Bucket:    bucket,
		Prefix:    tenantAccountMap[tenantId],
		GcsClient: client,
	}
}

func (a *Applovin) Sync() {
	ctx := context.Background()
	query := &storage.Query{Prefix: a.Prefix}
	it := a.GcsClient.Bucket(bucket).Objects(ctx, query)
	log.Printf("获取文件夹 %s/%s 中的文件:", bucket, a.Prefix)

	insert := []OrderJoinSource{}
	insertKey := []cache.Cache{}
	for {
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}

		if err != nil {
			fmt.Errorf("遍历文件失败: %w", err)
		}
		log.Printf("文件: %s, 大小: %d bytes", attrs.Name, attrs.Size)

		key := fmt.Sprintf("%d-%s", a.TenantId, attrs.Name)

		if cache.IsNeedUpdate(a.cache, a.TenantId, key, attrs.Updated) {
			fmt.Println("need load")
			insert = append(insert, a.Download(attrs.Name)...)
			insertKey = append(insertKey, cache.Cache{
				Key:          key,
				LastModified: attrs.Updated,
			})
		} else {
			log.Printf("跳过下载 %s (未修改)", key)
		}

		if len(insert) > 500 {
			a.Save(insert, insertKey)
			insert = []OrderJoinSource{}
			insertKey = []cache.Cache{}
		}
	}

	if len(insert) > 0 {
		a.Save(insert, insertKey)
	}
	return
}

func (a *Applovin) Save(data []OrderJoinSource, keys []cache.Cache) {
	InsertOrderJoinSource(data)
	cache.SaveS3CacheWithArr(a.cache, a.TenantId, keys)
}

func (a *Applovin) Download(prefix string) []OrderJoinSource {
	if !strings.Contains(prefix, "json") {
		log.Printf("skip %s", prefix)
		return nil
	}
	ctx := context.Background()

	reader, err := a.GcsClient.Bucket(a.Bucket).Object(prefix).NewReader(ctx)
	if err != nil {
		panic(err)
	}
	defer reader.Close()

	// 一行一行读取
	scanner := bufio.NewScanner(reader)
	var insertData = []OrderJoinSource{}
	for scanner.Scan() {
		var applovinData ResponseData

		if err = json.Unmarshal(scanner.Bytes(), &applovinData); err != nil {
			panic(err)
		}

		if applovinData.EventType == "Purchase" {
			continue
		}

		tp := fmt.Sprintf("applovin_log_%s", applovinData.EventType)
		insertData = append(insertData, OrderJoinSource{
			TenantId:      a.TenantId,
			ImportingType: tp,
			OrderId:       applovinData.OrderId,
			SrcEntityType: tp,
			SrcEntityId:   fmt.Sprintf("%d|%s｜%s|%s|%s", a.TenantId, applovinData.AdsetId, applovinData.CampaignId, applovinData.OrderId, applovinData.EventTime),
			SrcEventTime:  applovinData.EventTime,
			SrcChannel:    "ads",
			SrcSource:     "applovin",
			SrcAdId:       applovinData.AdsetId,
			SrcAdsetId:    "-",
			SrcCampaignId: applovinData.CampaignId,
			MetaData:      scanner.Text(),
		})
	}

	if len(insertData) == 0 {
		return nil
	}
	return insertData
	//InsertOrderJoinSource(insertData)
}

type ResponseData struct {
	OrderId    string `json:"order_id"`
	EventId    string `json:"event_id"`
	EventTime  string `json:"event_time"`
	EventType  string `json:"event_type"`
	AdId       int    `json:"ad_id"`
	AdsetId    string `json:"adset_id"`
	CampaignId string `json:"campaign_id"`
}
