package main

import (
	"encoding/json"
	"time"
	"wm-func/common/state"
	"wm-func/wm_account"
)

type State struct {
	Name            string    `json:"name"`
	LastSync        time.Time `json:"last_sync"`
	TimeRange       int       `json:"time_range"` // 时间间隔
	NextRunningTime time.Time `json:"next_running_time"`
	LastRunningTime time.Time `json:"last_running_time"`
}

func GetState(account wm_account.Account, subType string) (*State, error) {
	syncInfo := state.GetSyncInfo(account.TenantId, account.AccountId, Platform, subType)

	if syncInfo == nil {
		lastYear := time.Now().Add(Day * -1 * time.Duration(preDays))
		return &State{
			Name:            subType,
			LastSync:        lastYear,
			TimeRange:       timeRageHourMap[subType],
			LastRunningTime: lastYear,
			NextRunningTime: lastYear,
		}, nil
	}

	res := State{}
	if err := json.Unmarshal(syncInfo, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func SaveState(account wm_account.Account, s *State) error {
	if s == nil {
		panic("nil state")
	}

	s.LastRunningTime = time.Now()
	s.TimeRange = timeRageHourMap[s.Name]

	if s.TimeRange == 0 {
		s.TimeRange = 1
	}

	s.NextRunningTime = time.Now().Add(time.Hour * time.Duration(s.TimeRange))

	var b []byte
	var err error
	if b, err = json.Marshal(s); err != nil {
		panic(err)
	}

	state.SaveSyncInfo(account.TenantId, account.AccountId, Platform, s.Name, b)
	return nil
}
