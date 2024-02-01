package utils

import "time"

func ToFormat(date time.Time) string {
	location := GetLocation()
	format := "02 Jan 2006 15:04"
	return date.In(location).Format(format)
}

func ToFormatDate(date time.Time) string {
	location := GetLocation()
	format := "02 Jan 2006"
	return date.In(location).Format(format)
}

func GetLocation() *time.Location {
	location, _ := time.LoadLocation("Asia/Bangkok")
	return location
}

func Bod(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}
