package api

import (
	"github.com/gametaverse/gfdp/rpc/pb"
)

const (
	cDateFormat  = "2006-01-02"
	cSecondofDay = 60 * 60 * 24
)
const (
	RPCAddr = "172.31.6.11:8081"
)

var (
	AllChain = []pb.Chain{
		pb.Chain_BSC,
		pb.Chain_ETH,
		pb.Chain_POLYGON,
		pb.Chain_AVAX,
		pb.Chain_WAX,
		pb.Chain_SOLANA,
		pb.Chain_KARDIA,
	}
)

func ParsePBChain(chain string) pb.Chain {
	c := pb.Chain_UNKNOWN
	switch chain {
	case "bsc":
		c = pb.Chain_BSC
	case "polygon":
		c = pb.Chain_POLYGON
	case "eth":
		c = pb.Chain_ETH
	case "avax":
		c = pb.Chain_AVAX
	case "wax":
		c = pb.Chain_WAX
	case "solana":
		c = pb.Chain_SOLANA
	case "kardia":
		c = pb.Chain_KARDIA
	}
	return c
}
