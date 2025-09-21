package main

import (
	"log"
	"time"
	t_pool "wm-func/common/pool"
	"wm-func/wm_account"
)

func main() {
	accounts := wm_account.GetAccountsWithPlatformNotNull("pinterest")

	pool := t_pool.NewWorkerPool(1)
	pool.Run()

	for _, account := range accounts {
		if account.TenantId != 150219 {
			continue
		}
		ac := account
		pool.AddTask(func() {
			pinterest := NewPinterest(ac)
			traceId := pinterest.getTraceId()

			log.Printf("[%s] 开始处理Pinterest账户", traceId)

			// 初始化Campaign IDs
			if err := pinterest.InitCampaignIds(); err != nil {
				log.Printf("[%s] 初始化Campaign IDs失败: %v", traceId, err)
				return
			}

			// 拉取并保存Campaign数据
			if err := pinterest.PullOdsCampaignsAndSave(); err != nil {
				log.Printf("[%s] 处理Campaign数据失败: %v", traceId, err)
				return
			}

			// 拉取AdGroup数据
			if err := pinterest.PullOdsAdGroups(); err != nil {
				log.Printf("[%s] 处理AdGroup数据失败: %v", traceId, err)
				// 不返回，继续完成其他处理
			}

			// 拉取Ad数据
			if err := pinterest.PullOdsAds(); err != nil {
				log.Printf("[%s] 处理Ad数据失败: %v", traceId, err)
				// 不返回，继续完成其他处理
			}

			//拉取AdGroup Analytics数据
			if err := pinterest.PullAdGroupAnalytics(); err != nil {
				log.Printf("[%s] 处理AdGroup Analytics数据失败: %v", traceId, err)
				// 不返回，继续完成其他处理
			}

			// 拉取Ad Analytics数据
			if err := pinterest.PullAdAnalytics(); err != nil {
				log.Printf("[%s] 处理Ad Analytics数据失败: %v", traceId, err)
				// 不返回，继续完成其他处理
			}

			log.Printf("[%s] Pinterest账户处理完成", traceId)
		})
	}

	pool.Wait()
}

func (p *Pinterest) InitCampaignIds() error {
	traceId := p.getTraceId()
	log.Printf("[%s] 开始初始化Campaign IDs", traceId)

	ids, err := p.ProcessReportData()
	if err != nil {
		log.Printf("[%s] 获取报告数据失败: %v", traceId, err)
		return err
	}

	p.IdForCampaigns = ids
	log.Printf("[%s] 成功初始化Campaign IDs，共%d个", traceId, len(ids))
	return nil
}

// PullOdsCampaignsAndSave 拉取并保存Campaign数据
func (p *Pinterest) PullOdsCampaignsAndSave() error {
	traceId := p.getTraceId()
	log.Printf("[%s] 开始处理Campaign数据", traceId)

	// 拉取Campaign数据
	data, err := p.PullOdsCampaigns(p.IdForCampaigns)
	if err != nil {
		log.Printf("[%s] 拉取Campaign数据失败: %v", traceId, err)
		return err
	}

	// 保存 Campaign 数据到 Airbyte
	if len(data) > 0 {
		if err := SaveCampaignsToAirbyte(data, p.Account.TenantId); err != nil {
			log.Printf("[%s] 保存Campaign数据到Airbyte失败: %v", traceId, err)
			return err
		}
		log.Printf("[%s] 成功保存%d条Campaign数据到Airbyte", traceId, len(data))
	} else {
		log.Printf("[%s] 没有Campaign数据需要保存", traceId)
	}

	log.Printf("[%s] Campaign数据处理完成", traceId)
	return nil
}

func (p *Pinterest) PullOdsCampaigns(ids []string) ([]Campaign, error) {
	traceId := p.getTraceId()
	log.Printf("[%s] 开始拉取Campaign数据，共%d个IDs", traceId, len(ids))

	if len(ids) == 0 {
		log.Printf("[%s] 没有Campaign IDs需要处理", traceId)
		return nil, nil
	}

	// 每批处理50个ID
	batchSize := 50
	totalBatches := (len(ids) + batchSize - 1) / batchSize

	data := []Campaign{}

	log.Printf("[%s] 将分%d批处理，每批最多%d个IDs", traceId, totalBatches, batchSize)

	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}

		batch := ids[i:end]
		batchNum := (i / batchSize) + 1

		log.Printf("[%s] 处理第%d/%d批，包含%d个IDs", traceId, batchNum, totalBatches, len(batch))

		// 调用ListCampaigns获取Campaign详情
		campaigns, err := p.ListCampaigns(batch)
		if err != nil {
			log.Printf("[%s] 第%d批Campaign数据获取失败: %v", traceId, batchNum, err)
			// 继续处理下一批，不返回错误
			continue
		}

		if campaigns == nil || len(campaigns.Items) == 0 {
			log.Printf("[%s] 第%d批未获取到Campaign数据", traceId, batchNum)
			continue
		}

		log.Printf("[%s] 第%d批成功获取%d个Campaign", traceId, batchNum, len(campaigns.Items))

		data = append(data, campaigns.Items...)

		// 在批次之间添加短暂延迟，避免API限流
		if batchNum < totalBatches {
			log.Printf("[%s] 等待1秒后处理下一批", traceId)
			time.Sleep(1 * time.Second)
		}
	}

	log.Printf("[%s] Campaign数据拉取完成", traceId)
	return data, nil
}

// PullOdsAdGroups 拉取并保存AdGroup数据
func (p *Pinterest) PullOdsAdGroups() error {
	traceId := p.getTraceId()
	log.Printf("[%s] 开始处理AdGroup数据", traceId)

	if len(p.IdForCampaigns) == 0 {
		log.Printf("[%s] 没有Campaign IDs，跳过AdGroup数据处理", traceId)
		return nil
	}

	// 使用优化后的批量获取方法
	if err := p.PullAllAdGroupsAndSave(); err != nil {
		log.Printf("[%s] 拉取AdGroup数据失败: %v", traceId, err)
		return err
	}

	log.Printf("[%s] AdGroup数据处理完成", traceId)
	return nil
}

// PullOdsAds 拉取并保存Ad数据
func (p *Pinterest) PullOdsAds() error {
	traceId := p.getTraceId()
	log.Printf("[%s] 开始处理Ad数据", traceId)

	if len(p.IdForCampaigns) == 0 {
		log.Printf("[%s] 没有Campaign IDs，跳过Ad数据处理", traceId)
		return nil
	}

	// 使用优化后的批量获取方法
	if err := p.PullAllAdsAndSave(); err != nil {
		log.Printf("[%s] 拉取Ad数据失败: %v", traceId, err)
		return err
	}

	log.Printf("[%s] Ad数据处理完成", traceId)
	return nil
}
