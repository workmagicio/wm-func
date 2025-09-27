package main

import (
	"encoding/json"
	"fmt"
	"time"
)

func RequestQuestion(account FAccount) {
	res, err := callFairingQuestionsAPI(account)
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
}

func RequestResponse(account FAccount) {

	start := time.Now().UTC().Add(time.Hour * 24 * 30 * -1)
	//now := time.Now().UTC()

	//for start.Before(now) {
	data := []FairingUserResponse{}
	st := fmt.Sprintf("%sT00:00:00Z", start.Format("2006-01-02"))
	ed := fmt.Sprintf("%sT23:59:59Z", start.Format("2006-01-02"))

	res, err := callFairingResponsesAPI(account, st, ed, 100)
	if err != nil {
		panic(err)
	}

	data = append(data, res.Data...)

	var next = res.Next

	for next != nil {
		tmp, err2 := callFairingResponsesAPINext(account, *next)
		if err2 != nil {
			panic(err2)
		}

		data = append(data, tmp.Data...)
		next = tmp.Next
		time.Sleep(time.Second * 2)
	}

	var insertAirbyteData []AirbyteData

	for _, r := range data {
		var b []byte
		b, err = json.Marshal(r)
		if err != nil {
			panic(err)
		}

		now2 := time.Now().Format("2006-01-02 15:04:05")

		//fmt.Println(string(b))

		insertAirbyteData = append(insertAirbyteData, AirbyteData{
			TenantId:            account.TenantId,
			AirbyteRawId:        r.Id,
			AirbyteData:         b,
			AirbyteExtractedAt:  now2,
			AirbyteLoadedAt:     now2,
			AirbyteMeta:         `{}`,
			AirbyteGenerationId: 0,
			ItemType:            "-",
		})

	}

	SaveToAirbyte(account, insertAirbyteData, SubTypeResponse)
	start = start.Add(time.Hour * 24)

	//}

}
