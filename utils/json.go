package utils

import (
	"fmt"
	"os"
)

func ReadJSON(path string) (json []byte, err error) {
	json, err = os.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}
