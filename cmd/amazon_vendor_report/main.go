package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
	"wm-func/common/account"
	"wm-func/common/db/platform_db"
	t_pool "wm-func/common/pool"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func main() {
	log.Println("starting...")
	runCode()

	//ticker := time.NewTicker(10 * time.Minute)
	//for range ticker.C {
	//	runCode()
	//}
}

func runCode() {
	platform := "amazonVendorPartner"
	accounts := account.GetAccountsWithPlatform(platform)

	log.Println("本次有", len(accounts), "个用户在跑")

	pool := t_pool.NewWorkerPool(5)
	pool.Run()
	for _, account := range accounts {
		if account.TenantId == 150155 {
			continue
		}

		fmt.Printf("Task %d: started\n", account.TenantId)
		execute(account.TenantId, account.AccountId, account.RefreshToken, platform)
		fmt.Printf("Task %d: completed\n", account.TenantId)
		//model.SaveSyncInfo(account.TenantId, account.AccountId, platform)
	}
}

type StreamSlice struct {
	Start time.Time
	End   time.Time
}

func execute(tenantId int64, accountId, refreshToken string, platform string) {
	reportDate := GetSyncInfo(tenantId, accountId, platform)

	streamSlice := getStartTime(reportDate)
	for _, slice := range streamSlice {
		client := NewAmazonVendorReportRequestClient(refreshToken)
		requestId := client.createQuery("2025-08-07", "2025-08-10")
		//requestId := client.createQuery(slice.Start.Format("2006-01-02"), slice.End.Format("2006-01-02"))

		//requestId := client.createQuery(slice.Start.Format("2006-01-02"), slice.End.Format("2006-01-02"))
		log.Printf("Created query id: %s", requestId)
		if requestId == "" {
			return
		}

		log.Printf("Polling document id for query %s", requestId)
		documentId := getDocumentId(client, requestId)
		log.Printf("Obtained document id: %s", documentId)
		if documentId == "" {
			return
		}

		docRes := client.queryDocument(documentId)
		log.Printf("Document URL: %s", docRes.DocumentUrl)

		results := getDocumentResult(docRes.DocumentUrl)
		log.Printf("Parsed document, dates: %d", len(results))

		// 将解析结果转换为 DailyReport
		dailyReports := convertToDailyReports(results, tenantId)
		log.Printf("Converted to DailyReport count: %d", len(dailyReports))

		// 批量写入数据库（1000 条一批）
		if len(dailyReports) > 0 {
			insertDailyReports(platform_db.GetDB(), dailyReports, 1000)
		}

		break
		SaveSyncInfoWithTime(tenantId, accountId, platform, slice.End)

	}
}

func getDocumentId(client *AmazonVendorReportRequestClient, requestId string) string {
	status := client.queryReportStatus(requestId)
	log.Printf("Query %s status: %s", requestId, status.ProcessingStatus)
	if status.ProcessingStatus == "DONE" {
		return status.DataDocumentId
	} else if status.ProcessingStatus == "FATAL" {
		log.Println("ERROR status is fatal")
		return ""
	}
	time.Sleep(10 * time.Second)
	return getDocumentId(client, requestId)
}

func getDocumentResult(url string) map[string][]DocumentResponse {
	log.Printf("Downloading document from %s", url)
	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	reader := bufio.NewReader(res.Body)
	result := make(map[string][]DocumentResponse)
	lineCount := 0

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lineCount++
		var item DocumentResponse
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			log.Printf("failed to unmarshal line %d: %v", lineCount, err)
			continue
		}
		result[item.StartDate] = append(result[item.StartDate], item)
	}

	log.Printf("Finished parsing document. Total lines: %d, date keys: %d", lineCount, len(result))
	return result
}

// convert map[string][]DocumentResponse -> []*DailyReport
func convertToDailyReports(results map[string][]DocumentResponse, tenantId int64) []*DailyReport {
	var reports []*DailyReport
	for _, docs := range results {
		for _, doc := range docs {
			for _, metric := range doc.Metrics {
				ents := []string{"DAILY", "ZIP", doc.StartDate, metric.GroupByKey.Asin, metric.GroupByKey.ShipToCountryCode, metric.GroupByKey.ShipToStateOrProvince, metric.GroupByKey.ShipToCity, metric.GroupByKey.ShipToZipCode}
				reports = append(reports, &DailyReport{
					TenantId:              tenantId,
					Asin:                  metric.GroupByKey.Asin,
					EntityId:              strings.Join(ents, "|"),
					ShipToZipCode:         metric.GroupByKey.ShipToZipCode,
					ShipToCountryCode:     metric.GroupByKey.ShipToCountryCode,
					ShipToCity:            metric.GroupByKey.ShipToCity,
					ShipToStateOrProvince: metric.GroupByKey.ShipToStateOrProvince,
					StatDate:              doc.StartDate,
					ShippedRevenue:        metric.Metrics.ShippedOrders.ShippedUnitsWithRevenue.Value.Amount,
					ShippedUnits:          int64(metric.Metrics.ShippedOrders.ShippedUnitsWithRevenue.Units),
					SalesDiscount:         metric.Metrics.Costs.SalesDiscount.Amount,
					ShippedCogs:           metric.Metrics.Costs.ShippedCogs.Amount,
					ContraCogs:            metric.Metrics.Costs.ContraCogs.Amount,
				})
			}
		}
	}
	return reports
}

// 批量插入DailyReport
func insertDailyReports(db *gorm.DB, reports []*DailyReport, batchSize int) {
	log.Printf("Start inserting DailyReport into DB, batch size: %d", batchSize)
	for i := 0; i < len(reports); i += batchSize {
		end := i + batchSize
		if end > len(reports) {
			end = len(reports)
		}
		batch := reports[i:end]
		if err := db.Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(batch, len(batch)).Error; err != nil {
			log.Printf("failed to insert batch [%d:%d]: %v", i, end, err)
		} else {
			log.Printf("successfully inserted batch [%d:%d]", i, end)
		}
	}
}

func getStartTime(startStr string) []StreamSlice {
	start := time.Now()
	if startStr == "" {
		start = time.Now().Add(-180 * 24 * time.Hour)
	}

	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		panic(err)
	}

	var streamSlice []StreamSlice

	if start.Add(time.Hour * 24 * 30).After(time.Now()) {
		start = start.Add(time.Hour * 24 * 30 * -1)
	}
	now := time.Now()
	start = now.Add(time.Hour * 24 * 3 * -1)
	for start.Before(now) && now.Sub(start).Hours() > 1 {
		if start.Add(time.Hour * 23).After(now) {
			break
		}

		end := start.Add(time.Hour * 24 * 2)
		if end.After(now) {
			end = now
		}

		streamSlice = append(streamSlice, StreamSlice{
			Start: start,
			End:   end,
		})

		if end.Equal(now) {
			break
		}
		start = end
	}

	return streamSlice
}
