package domains

type HealthCheck struct {
	AppStatus          string `json:"app_status,omitempty"`
	AppMsg             string `json:"app_msg,omitempty"`
	DbStatus           string `json:"db_status,omitempty"`
	DbMsg              string `json:"db_msg,omitempty"`
	RdbStatus          string `json:"rdb_status,omitempty"`
	RdbMsg             string `json:"rdb_msg,omitempty"`
	DrachmaStatus      string `json:"drachma_status,omitempty"`
	DrachmaMsg         string `json:"drachma_msg,omitempty"`
	MedjatStatus       string `json:"medjat_status,omitempty"`
	MedjatMsg          string `json:"medjat_msg,omitempty"`
	ExchangeRateStatus string `json:"exchange_rate_status,omitempty"`
	ExchangeRateMsg    string `json:"exchange_rate_msg,omitempty"`
}
