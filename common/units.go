package common

func KBtoBytes(kb int64) int64 {
	return 1024 * kb
}

func MBtoBytes(mb int64) int64 {
	return 1024 * 1024 * mb
}

func GBtoBytes(gb int64) int64 {
	return 1024 * 1024 * 1024 * gb
}
