package main

var subTypeList = []string{SUBTYPE_QUESTION, SUBTYPE_RESPONSE, SUBTYPE_RESPONSE_COUNT, SUBTYPE_SURVEY}

var timeRageHourMap = map[string]int{
	SUBTYPE_RESPONSE:       1,
	SUBTYPE_SURVEY:         1,
	SUBTYPE_RESPONSE_COUNT: 12,
	SUBTYPE_QUESTION:       1,
}

var timeStepDailyMap = map[string]int{
	SUBTYPE_RESPONSE:       10,
	SUBTYPE_SURVEY:         -1,
	SUBTYPE_RESPONSE_COUNT: 1,
}

var preDays = 1
