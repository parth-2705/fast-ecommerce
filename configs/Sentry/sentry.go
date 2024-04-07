package Sentry

import (
	"os"

	"github.com/getsentry/sentry-go"
)

func Init() (err error) {

	var prodEnv bool = true
	if os.Getenv("ENVIRONMENT") != "prod" {
		prodEnv = false
	}

	err = sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("SENTRY_DSN"),
		AttachStacktrace: true,
		Debug:            !prodEnv,
		TracesSampleRate: 1.0,
	})

	return
}
