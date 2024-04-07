package GCS

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func DownloadImageFromBucket(bucketLink string, localFileName string) error {
	f, err := os.Create(localFileName)
	if err != nil {
		return fmt.Errorf("os.Create: %w", err)
	}
	defer f.Close()

	request, err := http.NewRequest("GET", bucketLink, nil)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return err
	}

	httpClient := &http.Client{}
	response, err := httpClient.Do(request)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return err
	}
	defer response.Body.Close()
	if _, err := io.Copy(f, response.Body); err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}

	return nil

}
