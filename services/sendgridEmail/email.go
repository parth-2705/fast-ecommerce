package sendgridEmail

import (
	"context"
	"fmt"
	"strconv"

	// "hermes/models"

	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/tryamigo/themis"
)

var InternalUsers map[string]struct{} = map[string]struct{}{
	// "+918587957686": {}, // Sankalp
	"+918800561308": {}, // Aryan
	"+919910074373": {}, // Shashank
	"+918426003071": {}, // Prashant
	"+919571049258": {}, // Prashant
	"+919599723760": {}, // Muskan
	"+917703933820": {}, // Tanvi
	"+917042072821": {}, // Abhishek
	"+919990165720": {}, // Samrat
	"+918097338559": {}, // Parth
	"+919089749849": {}, // Alou
	"+919643099621": {}, // Kshitij
	"+918700732909": {}, // Kshitij
	"+919687029466": {}, // Shalin
	"+919910606373": {}, // Support
	"+919910608373": {}, // Support
}

func getListofEmailsFromVault() []*mail.Email {

	emailsList := make([]*mail.Email, 0)

	csEmails, err := themis.GetSecret("EMAILS_TO_NOTIFY")
	if err != nil {
		return emailsList
	}

	stringEmailsList := strings.Split(csEmails, ",")
	for _, email := range stringEmailsList {
		name := strings.Split(email, "@")[0]
		emailsList = append(emailsList, mail.NewEmail(name, email))
	}

	return emailsList
}

func PopulateInternalUsersList() (err error) {

	// Get Comma Seperated list of numbers from Vault
	csNumbers, err := themis.GetSecret("INTERNAL_USERS_NUMBERS")
	if err != nil {

		fmt.Println(err)

		return
	}

	// Iterate over the numbers and add them to the Internal Users List
	for _, number := range strings.Split(csNumbers, ",") {
		InternalUsers[number] = struct{}{}
	}

	return
}

type EmailNotifier struct {
	UserPhone      string
	ProductID      string
	ProductName    string
	ProductURL     string
	BrandID        string
	BrandName      string
	UserID         string
	OrderID        string
	PaymentMethod  string
	Price          string
	AddressID      string
	ContactName    string
	Address        string
	Quantity       int
	ProductPrice   float64
	CouponDiscount float64
	CouponID       string
	RPBalance      float64
	ReferredBy     string
}

func SendOrderPlacedEmail(ctx context.Context, info EmailNotifier) error {

	defer sentry.Recover()

	if _, ok := InternalUsers[info.UserPhone]; ok {
		fmt.Println("Internal Order")
		return nil
	}

	from := mail.NewEmail("Amigo Data", "alerts@tryamigo.com")

	m := mail.NewV3Mail()

	// emailHTML := `
	// <p>
	// 	ProductID: %productID%  ProductName: %productName%<br>
	// 	BrandID: %brandID% BrandName: %brandName% <br>
	// 	userID: %userID% userPhone %userPhone% <br>
	// 	OrderID: %orderID% Selling Price:%sellingPrice% <br>
	// 	AddressID: %addressID% ContactName: %contactName% <br>
	// 	Address: %address%
	// </p>
	// `

	emailHTML2 := `
	<table style="width: 100%;border: 1px solid;">
		<tr>
			<th>ProductID:</th>
			<td>%productID%</td>
		</tr>
		<tr>
			<th>ProductName:</th>
			<td>%productName%</td>
		</tr>
		<tr>
			<th>ProductURL:</th>
			<td>%productURL%</td>
		</tr>
		<tr>
			<th>BrandName:</th>
			<td>%brandName%</td>
		</tr>
		<tr>
			<th>userID:</th>
			<td>%userID%</td>
		</tr>
		<tr>
			<th>userPhone:</th>
			<td>%userPhone%</td>
		</tr>
		<tr>
			<th>Referred by:</th>
			<td>%referredBy%</td>
		</tr>
		<tr>
			<th>OrderID:</th>
			<td>%orderID%</td>
		</tr>
		<tr>
			<th>Payment Method:</th>
			<td>%paymentMethod%</td>
		</tr>
		<tr>
			<th>userPhone:</th>
			<td>%userPhone%</td>
		</tr>
		<tr>
			<th>Quantity:</th>
			<td>%quantity%</td>
		</tr>
		<tr>
			<th>ProductPrice :</th>
			<td>%productPrice%</td>
		</tr>
		<tr>
			<th>CouponDiscount :</th>
			<td>%couponDiscount%</td>
		</tr>
		<tr>
			<th>CouponID :</th>
			<td>%couponID%</td>
		</tr>
		<tr>
			<th>Total Amount:</th>
			<td>%sellingPrice%</td>
		</tr>
		<tr>
			<th>RP Balance:</th>
			<td>%rpBalance%</td>
		</tr>
		<tr>
			<th>AddressID:</th>
			<td>%addressID%</td>
		</tr>
		<tr>
			<th>ContactName:</th>
			<td>%contactName%</td>
		</tr>
		<tr>
			<th>Address:</th>
			<td>%address%</td>
		</tr>
	</table>
	`

	content := mail.NewContent("text/html", emailHTML2)

	m.SetFrom(from)
	m.AddContent(content)

	// create new *Personalization
	personalization := mail.NewPersonalization()

	// populate `personalization` with data
	to1 := mail.NewEmail("Shalin", "shalin@roovo.in")
	// to2 := mail.NewEmail("Sankalp", "sankalp@amigo.gg")
	to3 := mail.NewEmail("Shashank", "shashank@amigo.gg")
	// to5 := mail.NewEmail("Aryan", "aryan@amigo.gg")

	personalization.AddTos(to1, to3)

	emailsListFromVault := getListofEmailsFromVault()

	personalization.AddTos(emailsListFromVault...)

	personalization.SetSubstitution("%productID%", info.ProductID)
	personalization.SetSubstitution("%productName%", info.ProductName)
	personalization.SetSubstitution("%productURL%", info.ProductURL)
	personalization.SetSubstitution("%brandID%", info.BrandID)
	personalization.SetSubstitution("%brandName%", info.BrandName)
	personalization.SetSubstitution("%userID%", info.UserID)
	personalization.SetSubstitution("%userPhone%", info.UserPhone)
	personalization.SetSubstitution("%orderID%", info.OrderID)
	personalization.SetSubstitution("%paymentMethod%", info.PaymentMethod)
	personalization.SetSubstitution("%sellingPrice%", info.Price)
	personalization.SetSubstitution("%addressID%", info.AddressID)
	personalization.SetSubstitution("%contactName%", info.ContactName)
	personalization.SetSubstitution("%address%", info.Address)
	personalization.SetSubstitution("%quantity%", strconv.FormatInt(int64(info.Quantity), 10))
	personalization.SetSubstitution("%productPrice%", strconv.FormatFloat(info.ProductPrice, 'f', 0, 64))
	personalization.SetSubstitution("%couponDiscount%", strconv.FormatFloat(info.CouponDiscount, 'f', 0, 64))
	personalization.SetSubstitution("%couponID%", info.CouponID)
	personalization.SetSubstitution("%rpBalance%", strconv.FormatFloat(info.RPBalance, 'f', 0, 64))
	personalization.SetSubstitution("%referredBy%", info.ReferredBy)

	personalization.Subject = "New Order placed on Roovo!"

	// add `personalization` to `m`
	m.AddPersonalizations(personalization)

	sendgridAPIKey, err := themis.GetSecret("SENDGRID_API_KEY")
	if err != nil {
		fmt.Println(err)
		return err
	}

	request := sendgrid.GetRequest(sendgridAPIKey, "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	request.Body = mail.GetRequestBody(m)
	_, err = sendgrid.API(request)
	if err != nil {

		fmt.Println(err)
		return err
	}
	return nil
}
