package fb

import (
	"encoding/json"
	"fmt"
	"hermes/models/Logs"
	"hermes/services/Sentry"
	"hermes/utils/data"
	env "hermes/utils/tmpl/Env"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tryamigo/themis"
)

type conversionEvent struct {
	EventName      string               `json:"event_name"`
	EventID        string               `json:"event_id"`
	EventTime      int64                `json:"event_time"`
	EventSourceUrl string               `json:"event_source_url"`
	ActionSource   string               `json:"action_source"`
	UserData       conversionUserData   `json:"user_data"`
	CustomData     conversionCustomData `json:"custom_data"`
}

type conversionUserData struct {
	ClientUserAgent string `json:"client_user_agent"`
	ClientIPAddress string `json:"client_ip_address"`
	FBP             string `json:"fbp,omitempty"`
	FBC             string `json:"fbc,omitempty"`
	Phone           string `json:"ph,omitempty"`
	Name            string `json:"fn,omitempty"`
	Pincode         string `json:"zp,omitempty"`
	Country         string `json:"country,omitempty"`
}

type conversionCustomData struct {
	ContentIDs  []string `json:"content_ids"`
	ContentType string   `json:"content_type"`
	Currency    string   `json:"currency,omitempty"`
	Value       float64  `json:"value,omitempty"`
}

var customDataStub = conversionCustomData{
	ContentIDs:  []string{"x06utj12zt", "o4f0ai6fiy", "q6zw6c5wio", "q6zw6c5wio", "q6zw6c5wio", "tlsgoejbie", "h6ym09yt8y", "wp4tfr678e", "dt5lb1zljz", "yxyx3a5vcf", "swb28fdjqa", "1yi9c25ogh", "gakaywf0cz", "i8rv10ngvo", "lrbgg7ad5p", "m4wz8cyzwc", "lb9pohckh1", "ys158vk99a", "ndnlppsait", "bml6psxk5a", "grkjlan3im", "odab3co7us", "3miey9bjsw", "kgqak31ypa", "wdxsydbagl", "el22z3x4i1", "10bm0gcp3c", "ttb7ufp84b", "1s9t51v6j1", "tikklmo495", "k4l32u4myf", "w9gyd23iuy", "2txrggfm0h", "u8xduthdsy", "la4amg1lsf", "ohqsios9y5", "kl4wax272w", "4zzj20i7m6", "jkqwmk28ca", "mjj33kxr0y", "nj4lam2gwg", "rsa27oasvk", "z8mt45c3xa", "zalgezjbyq", "q7i8xdelt1", "eiitho9ha5", "5mrtkz1kqz", "6fabfe1ld0", "hq9dmpqr8b", "73q60e4ei5", "hwu7xm06mh", "yy7hsqw4ku", "b7kr6t02nx", "jcz0b28dfh", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "27", "28", "29", "30", "31", "32", "33", "34", "35", "36", "37", "38", "39", "40", "41", "42", "43", "44", "45", "46", "47", "48", "49", "50", "51", "52", "53", "54", "55", "56", "57", "58", "59", "60", "61", "62", "63", "64", "65", "66", "67", "68", "69", "70", "71", "72", "73", "74", "75", "76", "77", "78", "79", "80", "81", "82", "83", "84", "85", "86", "87", "88", "89", "90", "91", "92", "93", "94", "95", "96", "97", "98", "99", "100", "101", "102", "103", "104", "105", "106", "107", "108", "109", "110", "111", "112", "113", "114", "115", "116", "117", "118", "119", "120", "121", "122", "123", "124", "125", "126", "127", "128", "129", "130", "131", "132", "133", "134", "135", "136", "137", "138", "139", "140", "141", "142", "143", "144", "145", "146", "147", "148", "149", "150", "151", "152", "153", "154", "155", "156", "157", "158", "159", "160", "161", "162", "163", "164", "165", "166", "167", "168", "169", "170", "171", "172", "173", "174", "175", "176", "177", "178", "179", "180", "181", "182", "183", "184", "185", "186", "187", "188", "189", "190", "191", "192", "193", "194", "195", "196", "197", "198", "199", "200", "201", "202", "203", "204", "205", "206", "207", "208", "209", "210", "211", "212", "213", "214", "215", "216", "217", "218", "219", "220", "221", "222", "223", "224", "225", "226", "227", "228", "229", "230", "231", "232", "233", "234", "235", "236", "237", "238", "239", "240", "241", "242", "243", "244", "245", "246", "247", "248", "249", "250", "251", "252", "253", "254", "255", "256", "257", "258", "259", "260", "261", "262", "263", "264", "265", "266", "267", "268", "269", "270", "271", "272", "273", "274", "275", "276", "277", "278", "279", "280", "281", "282", "283", "284", "285", "286", "287", "288", "289", "290", "291", "292", "293", "294", "295", "296", "297", "298", "299", "300", "301", "302", "303", "304", "305", "306", "307", "308", "309", "310", "311", "312", "313", "314", "315", "316", "317", "318", "319", "320", "321", "322", "323", "324", "325", "326", "327", "328", "329", "330", "331", "332", "333", "334", "335", "336", "337", "338", "339", "340", "341", "342", "343", "344", "345", "346", "347", "348", "349", "350", "351", "352", "353", "354", "355", "356", "357", "358", "359", "360", "361", "362", "363", "364", "365", "366", "367", "368", "369", "370", "371", "372", "373", "374", "375", "376", "377", "378", "379", "380", "381", "382", "383", "384", "385", "386", "387", "388", "389", "390", "391", "392", "393", "394", "395", "396", "397", "398", "399", "400", "401", "402", "403", "404", "405", "406", "407", "408", "409", "410", "411", "412", "413", "414", "415", "416", "417", "418", "419", "420", "421", "422", "423", "424", "425", "426", "427", "428", "429", "430", "431", "432", "433", "434", "435", "436", "437", "438", "439", "440", "441", "442", "443", "444", "445", "446", "447", "448", "449", "450", "451", "452", "453", "454", "455", "456", "457", "458", "459", "460", "461", "462", "463", "464", "465", "466", "467", "468", "469", "470", "471", "472", "473", "474", "475", "476", "477", "478", "479", "480", "481", "482", "483", "484", "485", "486", "487", "488", "489", "490", "491", "492", "493", "494", "495", "496", "497", "498", "499", "500", "501", "502", "503", "504", "505", "506", "507", "508", "509", "510", "511", "512", "513", "514", "515", "516", "517", "518", "519", "520", "521", "522", "523", "524", "525", "526", "527", "528", "529", "530", "531", "532", "533", "534", "535", "536", "537", "538", "539", "540", "541", "542", "543", "544", "545", "546", "547", "548", "549", "550", "551", "552", "553", "554", "555", "556", "557", "558", "559", "560", "561", "562", "563", "564", "565", "566", "567", "568", "569", "570", "571", "572", "573", "574", "575", "576", "577", "578", "579", "580", "581", "582", "583", "584", "585", "586", "587", "588", "589", "590", "591", "592", "593", "594", "595", "596", "597", "598", "599", "600", "601", "602", "603", "604", "605", "606", "607", "608", "609", "610", "611", "612", "613", "614", "615", "616", "617", "618", "619", "620", "621", "622", "623", "624", "625", "626", "627", "628", "629", "630", "631", "632", "633", "634", "635", "636", "637", "638", "639", "640", "641", "642", "643", "644", "645", "646", "647", "648", "649", "650", "651", "652", "653", "654", "655", "656", "657", "658", "659", "660", "661", "662", "663", "664", "665", "666", "667", "668", "669", "670", "671", "672", "673", "674", "675", "676", "677", "678", "679", "680", "681", "682", "683", "684", "685", "686", "687", "688", "689", "690", "691", "692", "693", "694", "695", "696", "697", "698", "699", "700", "701", "702", "703", "704", "705", "706", "707", "708", "709", "710", "711", "712", "713", "714", "715", "716", "717", "718", "719", "720", "721", "722", "723", "724", "725", "726", "727", "728", "729", "730", "731", "732", "733", "734", "735", "736", "737", "738", "739", "740", "741", "742", "743", "744", "745", "746", "747", "748", "749", "750", "751", "752", "753", "754", "755", "756", "757", "758", "759", "760", "761", "762", "763", "764", "765", "766", "767", "768", "769", "770", "771", "772", "773", "774", "775", "776", "777", "778", "779", "780", "781", "782", "783", "784", "785", "786", "787", "788", "789", "790", "791", "792", "793", "794", "795", "796", "797", "798", "799", "800", "801", "802", "803", "804", "805", "806", "807", "808", "809", "810", "811", "812", "813", "814", "815", "816", "817", "818", "819", "820", "821", "822", "823", "824", "825", "826", "827", "828", "829", "830", "831", "832", "833", "834", "835", "836", "837", "838", "839", "840", "841", "842", "843", "844", "845", "846", "847", "848", "849", "850", "851", "852", "853", "854", "855", "856", "857", "858", "859", "860", "861", "862", "863", "864", "865", "866", "867", "868", "869", "870", "871", "872", "873", "874", "875", "876", "877", "878", "879", "880", "881", "882", "883", "884", "885", "886", "887", "888", "889", "890", "891", "892", "893", "894", "895", "896", "897", "898", "899", "900", "901", "902", "903", "904", "905", "906", "907", "908", "909", "910", "911", "912", "913", "914", "915", "916", "917", "918", "919", "920", "921", "922", "923", "924", "925", "926", "927", "928", "929", "930", "931", "932", "933", "934", "935", "936", "937", "938", "939", "940", "941", "942", "943", "944", "945", "946", "947", "948", "949", "950", "951", "952", "953", "954", "955", "956", "957", "958", "959", "960", "961", "962", "963", "964", "965", "966", "967", "968", "969", "970", "971", "972", "973", "974", "975", "976", "977", "978", "979", "980", "981", "982", "983", "984", "985", "986", "987", "988", "989", "990", "991", "992", "993", "994", "995", "996", "997", "998", "999", "1000", "1001", "1002", "1003", "1004", "1005", "1006", "1007", "1008", "1009", "1010", "1011", "1012", "1013", "1014", "1015", "1016", "1017", "1018", "1019", "1020", "1021", "1022", "1023", "1024", "1025", "1026", "1027", "1028", "1029", "1030", "1031", "1032", "1033", "1034", "1035", "1036", "1037", "1038", "1039", "1040", "1041", "1042", "1043", "1044", "1045", "1yi9c25ogh", "i8rv10ngvo", "6fabfe1ld0", "dt5lb1zljz", "eiitho9ha5", "gakaywf0cz", "h6ym09yt8y", "hq9dmpqr8b", "5mrtkz1kqz", "m4wz8cyzwc", "lrbgg7ad5p", "q7i8xdelt1", "swb28fdjqa", "tlsgoejbie", "wp4tfr678e", "yxyx3a5vcf", "z8mt45c3xa", "zalgezjbyq", "8900002134814", "8900002134319", "8900002134678", "PXS-005", "PXS-004", "PXS-003", "109", "122", "189", "190", "298", "299", "BWH-BB-0000150", "BWH-CB-0000200", "B-PSCW-W240200", "B-PSHZ-0000150", "NS001", "NS002", "NS003", "NS004", "NS005", "109", "122", "189", "190", "298", "299", "qravhi8q9u", "x5oo0d2fiu", "gvzfwq6987", "x4ui6hbjj4", "BRW-FS-0000250", "CRAN1", "B-PSSS-0000250", "8900002134678", "B-PSPS-0000250", "SP001", "8900002134814", "CASHEW1", "B-CHR-0200", "TOSLIM100", "HAZEL1", "BLUE1", "SP002", "B-STR-0250", "TOMSLA100", "TOCHAMFL50", "SP003", "SP004", "SP005", "PRIME_1", "8900002134616", "8900002134319", "723803345122", "8906069010306", "MANGO123","BIOTIN12", "BWH-CB-0000200-1", "B-PSCW-W240200-1", "B-PSHZ-0000150-1", "BWH-BB-0000150-1", "PRS400ML", "PFM250G", "PCM250G","BIOTIN12", "BWH-CB-0000200-1", "B-PSCW-W240200-1", "B-PSHZ-0000150-1", "BWH-BB-0000150-1", "PRS400ML", "PFM250G", "PCM250G","8900002134401", "8900002134104", "8900002134302", "8900002134661", "APPLE1234","APPLE123","MIX201A","A2CowGhee", "THCMuseli"},
	ContentType: "product",
}

type conversionsAPIPayload struct {
	Data     []conversionEvent `json:"data"`
	TestCode string            `json:"test_event_code,omitempty"`
}

func SendViewContentEvent(c *gin.Context, productID string) error {

	event := conversionEvent{
		EventName:      VIEWCONTENT,
		EventID:        productID,
		EventTime:      time.Now().Unix(),
		EventSourceUrl: "https://roovo.in/product",
		ActionSource:   "website",
		UserData: conversionUserData{
			ClientUserAgent: data.GetUserAgentFromSession(c),
			ClientIPAddress: data.GetIPAddressFromSession(c),
			FBP:             data.GetFBPCookie(c),
			FBC:             data.GetFBCCookie(c),
		},
		CustomData: customDataStub,
	}

	err := SendFBAdsConversionAPIRequest(event)
	return err
}

func SendAddToCartEvent(cartID string, clientUserAgent, clientIPAddress, fbp, fbc string) error {

	event := conversionEvent{
		EventName:      ADDTOCARTEVENT,
		EventID:        cartID,
		EventTime:      time.Now().Unix(),
		ActionSource:   "website",
		EventSourceUrl: "https://roovo.in/cart/create",
		UserData: conversionUserData{
			ClientUserAgent: clientUserAgent,
			ClientIPAddress: clientIPAddress,
			FBP:             fbp,
			FBC:             fbc,
		},
		CustomData: customDataStub,
	}

	err := SendFBAdsConversionAPIRequest(event)
	return err
}

func SendChatPurchaseEvent(orderValue float64, orderID string, phone string, pincode string) error {

	event := conversionEvent{
		EventName:      PURCHASE,
		EventID:        orderID,
		EventTime:      time.Now().Unix(),
		EventSourceUrl: "https://api-roovo.in/order",
		ActionSource:   "chat",
		CustomData: conversionCustomData{
			ContentIDs:  customDataStub.ContentIDs,
			ContentType: customDataStub.ContentType,
			Currency:    "INR",
			Value:       orderValue,
		},
		UserData: conversionUserData{
			Phone:   data.NormalizePhoneAndHash(phone),
			Pincode: data.SHA256Hash(pincode),
		},
	}

	err := SendFBAdsConversionAPIRequest(event)

	return err

}

func SendPurchaseEvent(orderValue float64, orderID string, userAgent string, userIP string, fbc string, fbp string, phone string, pincode string) error {

	event := conversionEvent{
		EventName:      PURCHASE,
		EventID:        orderID,
		EventTime:      time.Now().Unix(),
		EventSourceUrl: "https://roovo.in/order/create",
		ActionSource:   "website",
		UserData: conversionUserData{
			ClientUserAgent: userAgent,
			ClientIPAddress: userIP,
			FBP:             fbp,
			FBC:             fbc,
			Phone:           data.NormalizePhoneAndHash(phone),
			Pincode:         data.SHA256Hash(pincode),
		},
		CustomData: conversionCustomData{
			ContentIDs:  customDataStub.ContentIDs,
			ContentType: customDataStub.ContentType,
			Currency:    "INR",
			Value:       orderValue,
		},
	}

	err := SendFBAdsConversionAPIRequest(event)

	return err
}

func SendFBAdsConversionAPIRequest(event conversionEvent) error {

	if !env.IsProd() {
		return nil
	}

	userParamsJson, _ := json.Marshal(event.UserData)

	log, _ := Logs.CreateFBAdsConversionLog(event.EventName, event.ActionSource, userParamsJson)

	event.UserData.Country = "582967534d0f909d196b97f9e6921342777aea87b46fa52df165389db1fb8ccf" // Add SHA256 hashed County code for India (in)

	payload := conversionsAPIPayload{
		Data: []conversionEvent{event},
	}

	requestBody, _ := json.Marshal(payload)

	var CONVERSIONSAPIVERSION = os.Getenv("CONVERSIONS_API_VERSION")
	var METAPIXEL = os.Getenv("META_PIXEL_ID")
	var CONVERSIONSAPIACCESSTOKEN = os.Getenv("CONVERSIONS_API_ACCESSTOKEN")

	conversionsAPIEndpoint := fmt.Sprintf("https://graph.facebook.com/%s/%s/events?access_token=%s", CONVERSIONSAPIVERSION, METAPIXEL, CONVERSIONSAPIACCESSTOKEN)

	statusCode, responseBody, err := themis.HitAPIEndpoint2(conversionsAPIEndpoint, "POST", requestBody, nil, nil)
	if err != nil {
		Sentry.SentryCaptureException(err)
		return err
	}

	log.UpdateResponseStatus(statusCode, string(responseBody))

	if statusCode >= 400 {
		err = fmt.Errorf("%s", string(responseBody))
		fmt.Printf("err: %v\n", err)
		Sentry.SentryCaptureException(err)
		return err
	}

	return nil
}
