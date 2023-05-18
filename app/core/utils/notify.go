package utils

import (
	"context"
	"errors"
	"os"
	"pos/notify"
)

func NotifyMassage(massage string) (*notify.Response, error) {
	token := os.Getenv("LINE_TOKEN")
	if token == "" {
		err := errors.New("line token empty")
		return nil, err
	}
	c := notify.NewClient()
	res, err := c.Notify(context.Background(), token, massage)
	if err != nil {
		return res, err
	}
	return res, nil
}
