package response

type Query struct {
	Status      string `json:"status"`
	Message     string `json:"message,omitempty"`
	DataSource  string `json:"dataSource"`
	Country     string `json:"country"`
	CountryCode string `json:"countryCode"`
	Region      string `json:"region"`
	Timezone    string `json:"timezone"`
	UTCOffset   int    `json:"utcOffset"`
	Org         string `json:"org"`
	ISP         string `json:"isp"` // when no ISP data available, set to empty string
	ASN         string `json:"asn"`
	Anycast     bool   `json:"anycast,omitempty"` // only ipinfo-free can provided anycast info
}
