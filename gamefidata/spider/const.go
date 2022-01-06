package spider

/*
	case "polygon", "eth", "bsc":
		s.antenna = NewETHAntenna()
	case "avax":
		s.antenna = NewAvaxAntenna()
	case "wax":
		s.antenna = NewEOSAntenna()
	case "solana":
		s.antenna = NewSolanaAntenna()
*/

type ChainName string

const (
	ChainNameWax     = "wax"
	ChainNamePolygon = "polygon"
	ChainNameEth     = "eth"
	ChainNameBsc     = "bsc"
	ChainNameAvax    = "avax"
	ChainNameSolana  = "solana"
)

func ValiedChainName(name string) bool {
	switch name {
	case ChainNameWax,
		ChainNamePolygon,
		ChainNameEth,
		ChainNameBsc,
		ChainNameAvax,
		ChainNameSolana:
		return true
	default:
		return false
	}
}
