package configs

import "os"

var Secret = []byte(os.Getenv("SECRET"))

const Userkey = "user"
const Influencerkey = "influencer"
const AdminHeader = "admin"
const SellerHeader = "seller"
const AddressKey = "address"
const MobileNumberKey = "mobileNumber"
const CountryCodeKey = "countryCode"
const OrderKey = "orderId"
const UserAgentIdentifier = "userAgentIdentifier"
const UserAgent = "userAgent"
const IP = "IPAddress"
const ProductKey = "productID"
const CouponKey = "couponCode"
const Cart = "cartID"
const Internal = "internal"
const Referral = "referral"
const UTM = "utm"
