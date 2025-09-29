package tags

var default_filter_tenants = []int64{
	133822, 133849, 134531, 150076, 150075, 150078, 150079, 150080, 150081, 150082, 150083,
}

func GetDefaultTags() map[int64]string {
	res := map[int64]string{}
	for i := 0; i < len(default_filter_tenants); i++ {
		res[default_filter_tenants[i]] = "code_filter_region"
	}
	return res
}
