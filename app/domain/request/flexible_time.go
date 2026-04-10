package request

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type FlexibleTime struct {
	time.Time
}

func NewFlexibleTime(t time.Time) FlexibleTime {
	return FlexibleTime{Time: t}
}

func parseFlexibleTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("empty time value")
	}

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.999",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}

	for _, layout := range layouts {
		var (
			t   time.Time
			err error
		)
		switch layout {
		case "2006-01-02T15:04:05.999", "2006-01-02T15:04:05", "2006-01-02":
			t, err = time.ParseInLocation(layout, value, time.Local)
		default:
			t, err = time.Parse(layout, value)
		}
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported time format: %s", value)
}

func (ft *FlexibleTime) UnmarshalText(text []byte) error {
	value := strings.TrimSpace(string(text))
	if value == "" {
		ft.Time = time.Time{}
		return nil
	}

	parsed, err := parseFlexibleTime(value)
	if err != nil {
		return err
	}
	ft.Time = parsed
	return nil
}

func (ft *FlexibleTime) UnmarshalParam(value string) error {
	return ft.UnmarshalText([]byte(value))
}

func (ft *FlexibleTime) UnmarshalJSON(data []byte) error {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "null" || trimmed == "" {
		ft.Time = time.Time{}
		return nil
	}

	var value string
	if err := json.Unmarshal(data, &value); err == nil {
		return ft.UnmarshalText([]byte(value))
	}

	return ft.UnmarshalText([]byte(trimmed))
}

func (ft FlexibleTime) MarshalJSON() ([]byte, error) {
	if ft.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(ft.Time.Format(time.RFC3339Nano))
}
