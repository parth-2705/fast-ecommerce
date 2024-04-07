package configs

import (
	"fmt"
	"os"
	"time"

	"github.com/Flagsmith/flagsmith-go-client"
)

var fs *flagsmith.Client

func InitializeFlagsmith2() error {
	apiKey := os.Getenv("FLAGSMITH_API_KEY")
	fs = flagsmith.NewClient(apiKey, flagsmith.Config{
		BaseURI: os.Getenv("FLAGSMITH_BASE_URL"),
		Timeout: 5 * time.Second,
	})

	t, err := fs.GetFeatures()
	fmt.Printf("t: %v\n", t)
	return err
}

func GetFeature(feature string) bool {
	f, err := fs.FeatureEnabled(feature)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return false
	}
	return f
}
