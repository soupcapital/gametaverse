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
		pb.Chain_KARDIA,
	}
)

func UnparsePBChain(chain pb.Chain) string {
	c := "unknown"
	switch chain {
	case pb.Chain_AVAX:
		c = "avax"
	case pb.Chain_BSC:
		c = "bsc"
	case pb.Chain_ETH:
		c = "eth"
	case pb.Chain_KARDIA:
		c = "kardia"
	case pb.Chain_POLYGON:
		c = "polygon"
	case pb.Chain_SOLANA:
		c = "solana"
	case pb.Chain_WAX:
		c = "wax"
	}
	return c
}

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
