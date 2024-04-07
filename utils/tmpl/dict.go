package tmpl

import (
	"errors"
	"fmt"
	"reflect"
	"time"
)

func MakeMap(values ...interface{}) (map[string]interface{}, error) {
	if len(values) == 0 {
		return nil, errors.New("invalid dict call")
	}

	dict := make(map[string]interface{})

	for i := 0; i < len(values); i++ {
		key, isset := values[i].(string)
		if !isset {
			if reflect.TypeOf(values[i]).Kind() == reflect.Map {
				m := values[i].(map[string]interface{})
				for i, v := range m {
					dict[i] = v
				}
			} else {
				return nil, errors.New("dict values must be maps")
			}
		} else {
			i++
			if i == len(values) {
				return nil, errors.New("specify the key for non array values")
			}
			dict[key] = values[i]
		}

	}
	return dict, nil
}

func TimeLeft(endsAt time.Time) string {

	startAt := time.Now()
	fmt.Println("endsAt:", endsAt, startAt)
	days := endsAt.Sub(startAt)
	hours := endsAt.Sub(startAt.Add(time.Duration(int64(days.Hours()/24)) * time.Hour * 24))
	minutes := endsAt.Sub(startAt.Add(time.Duration(int64(days.Hours()/24))*time.Hour*24 + time.Duration(int64(hours.Hours()))*time.Hour))
	seconds := endsAt.Sub(startAt.Add(time.Duration(int64(days.Hours()/24))*time.Hour*24 + time.Duration(int64(hours.Hours()))*time.Hour + time.Duration(int64(minutes.Minutes()))*time.Minute))
	timeLeft := fmt.Sprintf("%02d:%02d:%02d:%02d", int64(days.Hours()/24), int64(hours.Hours()), int64(minutes.Minutes()), int64(seconds.Seconds()))
	return timeLeft
}

func TimeLeftInHours(startedAt time.Time) string {
	timeNow := time.Now()
	span := timeNow.Sub(startedAt)
	timeLeft := fmt.Sprintf("%02d hrs ago", int64(span.Hours()))
	return timeLeft
}
