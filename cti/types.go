package cti

type CoinInfo struct {
	ID                       string  `json:"id"`
	Symbol                   string  `json:"symbol"`
	Name                     string  `json:"name"`
	Price                    float32 `json:"price"`
	PriceChangePercentage24h float64 `json:"price_change_percentage_24h"`
	PriceChangePercentage7d  float64 `json:"price_change_percentage_7d"`
}

type CoinWindowItem struct {
	Coin  string
	Count int
}
