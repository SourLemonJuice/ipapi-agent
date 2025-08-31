package respstruct

type Query struct {
	Status      string `json:"status"`
	Message     string `json:"message,omitempty"`
	DataSource  string `json:"dataSource"`
	Country     string `json:"country"`
	CountryCode string `json:"countryCode"`
	Region      string `json:"region"`
	Timezone    string `json:"timezone"`
	UTCOffset   int    `json:"utcOffset"`
	ISP         string `json:"isp"`
	Org         string `json:"org"`
	ASN         string `json:"asn"`
}
