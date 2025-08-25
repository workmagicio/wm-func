package main

import (
	"fmt"
	"testing"
	"time"
	"wm-func/wm_account"
)

func TestRunning(t *testing.T) {
	accounts := wm_account.GetAccountsWithPlatform(Platform)
	for _, account := range accounts {
		res, err := RefreshToken(account)
		if err != nil {
			panic(err)
		}
		fmt.Println(res)

		state, err := GetState(account, "question")
		if err != nil {
			panic(err)
		}

		SaveState(account, state)
	}

	fmt.Println(accounts)
}

func TestStreamSlice(t *testing.T) {
	// --- 示例 1: 完整分割 ---
	fmt.Println("--- 示例 1: 可以被完整分割的情况 ---")
	dateFormat := "2006-01-02"
	//dateFormat := "2006-01-02 15:04:05"
	startTime := time.Now().Add(time.Hour * 24 * 50 * -1)
	endTime := time.Now()
	duration := Day * 7

	streamSlice := NewStreamSlice(startTime, endTime, dateFormat, duration)
	timeSlices := streamSlice.GetSlice()

	fmt.Println("生成的时间片:")
	for i, slice := range timeSlices {
		fmt.Printf("时间片 %d: Start: %s, End: %s\n", i+1, slice.Start, slice.End)
	}

	// --- 示例 2: 最后一个时间片不足一个完整 Duration ---
	fmt.Println("\n--- 示例 2: 最后一个时间片不完整的情况 ---")
	endTime2 := time.Now().Add(time.Hour * 24 * 29 * -1)
	streamSlice2 := NewStreamSlice(startTime, endTime2, dateFormat, duration)
	timeSlices2 := streamSlice2.GetSlice()

	fmt.Println("生成的时间片:")
	for i, slice := range timeSlices2 {
		fmt.Printf("时间片 %d: Start: %s, End: %s\n", i+1, slice.Start, slice.End)
	}
}
