package intz

func Int64ToDefault(i *int64, def int64) int64 {
	if i == nil {
		return def
	}

	return *i
}
