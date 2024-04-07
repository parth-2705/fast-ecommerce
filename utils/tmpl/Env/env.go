package env

import "os"

func IsProd() bool {
	env := os.Getenv("ENVIRONMENT")
	return env == "prod"
}
