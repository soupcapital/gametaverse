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

func ParsePBChain(chain string) pb.Chain {
	c := pb.Chain_UNKNOWN
	switch chain {
	case "bsc":
		c = pb.Chain_BSC
	case "polygon":
		c = pb.Chain_POLYGON
	case "eth":
		c = pb.Chain_ETH
	}
	return c
}
