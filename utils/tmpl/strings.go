package tmpl

import (
	"fmt"
	"hermes/models"
	"strings"
	"time"
)

// template function that truncates a string to a given length and adds an ellipsis
func Truncate(s string, length int) string {
	if len(s) > length {
		return s[:length] + "..."
	}
	return s
}

func Substring(s string, start int, length int) string {
	maxLength := start
	if start+length-1 < len(s) {
		maxLength = start + length - 1
	} else {
		maxLength = len(s)
	}
	return s[start:maxLength]
}

func ToString(s any) string {
	return fmt.Sprint(s)
}

func ToUpperCase(s string) string {
	return strings.ToUpper(s)
}

func ToTitleCase(s string) string {
	return strings.Title(s)
}

func FormatTimestamp(timestamp time.Time) string {
	loc, _ := time.LoadLocation("Asia/Kolkata")
	formattedStr := timestamp.In(loc).Format("Monday, 02 Jan")
	return formattedStr
}

func GetRandomizerStringtoControlCaching() string {
	return fmt.Sprintln(time.Now().UnixMilli())[:8]
}

func MakeGenderChoiceString(obj models.GenderChoice) string {
	tempArr := []string{}
	if obj.Male {
		tempArr = append(tempArr, "Male")
	}
	if obj.Female {
		tempArr = append(tempArr, "Female")
	}
	if obj.Others {
		tempArr = append(tempArr, "Others")
	}
	return strings.Join(tempArr, ",")
}
