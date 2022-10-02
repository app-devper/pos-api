package utils

import (
	"context"
	"errors"
	"os"
	"pos/app/core/notify"
	"time"
)

func InitContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	return ctx, cancel
}

func NotifyMassage(massage string) (*notify.Response, error) {
	token := os.Getenv("LINE_TOKEN")
	if token == "" {
		err := errors.New("line token empty")
		return nil, err
	}
	c := notify.NewClient()
	res, err := c.NotifyMessage(context.Background(), token, massage)
	if err != nil {
		return res, err
	}
	return res, nil
}

func ToFormat(date time.Time) string {
	location, _ := time.LoadLocation("Asia/Bangkok")
	format := "02 Jan 2006 15:04"
	return date.In(location).Format(format)
}
