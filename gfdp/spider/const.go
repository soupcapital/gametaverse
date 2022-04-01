package spider

type ChainName string

const (
	ChainNameWax     = "wax"
	ChainNamePolygon = "polygon"
	ChainNameEth     = "eth"
	ChainNameBsc     = "bsc"
	ChainNameAvax    = "avax"
	ChainNameSolana  = "solana"
	ChainNameKardia  = "kardia"
)

func ValiedChainName(name string) bool {
	switch name {
	case ChainNameWax,
		ChainNamePolygon,
		ChainNameEth,
		ChainNameBsc,
		ChainNameAvax,
		ChainNameSolana,
		ChainNameKardia:
		return true
	default:
		return false
	}
}
