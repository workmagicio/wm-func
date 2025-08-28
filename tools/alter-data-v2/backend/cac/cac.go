package cac

import (
	"time"
	"wm-func/common/config"
)

type Cac struct {
}

type DateSequence struct {
	Date    string `json:"date"`
	ApiData int64  `json:"api_data"`
	Data    int64  `json:"data"`
}

func GenerateDateSequence() []DateSequence {
	now := time.Now()
	start := now.Add(config.DateDay * -90)
	var res []DateSequence

	for start.Before(now) {
		res = append(res, DateSequence{
			Date:    start.Format("2006-01-02"),
			ApiData: 0,
			Data:    0,
		})
		start = start.Add(config.DateDay)
	}

	return res
}
