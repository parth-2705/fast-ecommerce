package payments

import (
	"os"

	"github.com/stripe/stripe-go/v74"
)

func Init() {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
}
