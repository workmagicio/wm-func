package data

type DataSlice struct {
	Date    string
	APiData int64
	WMData  int64
}

type Analytics struct {
	TenantId     int64
	IsNewTenant  bool
	Data         []DataSlice
	RegisterTime string
	Tags         []string
	avg          int64
	zeroData     int64
	HaveApiData  bool
	LastSyncDate string
}

func (a Analytics) GetAvg() int64 {

	if a.avg > 1 {
		return a.avg
	}

	lens := int64(len(a.Data))
	var apiTotal, wmTotal int64

	for i := 0; i < len(a.Data); i++ {
		apiTotal += a.Data[i].APiData
		wmTotal += a.Data[i].WMData
	}

	if apiTotal > 10 {
		a.avg = apiTotal / lens
	} else {
		a.avg = wmTotal / lens
	}

	return a.avg
}

func (a Analytics) ZeroData() int64 {
	if a.zeroData > 0 {
		return a.zeroData
	}
	a.zeroData = a.GetAvg() / 10
	return a.zeroData
}
