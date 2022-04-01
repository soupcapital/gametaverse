package rpc

import (
	"fmt"

	"github.com/gametaverse/gfdp/rpc/pb"
)

func getTableName(chain pb.Chain) (table string, err error) {
	chainName, err := chainName(chain)
	if err != nil {
		return
	}
	table = fmt.Sprintf("t_tx_%s", chainName)
	return
}

func chainName(chain pb.Chain) (table string, err error) {
	switch chain {
	case pb.Chain_BSC:
		table = "bsc"
	case pb.Chain_ETH:
		table = "eth"
	case pb.Chain_POLYGON:
		table = "polygon"
	case pb.Chain_AVAX:
		table = "avax"
	case pb.Chain_WAX:
		table = "wax"
	case pb.Chain_SOLANA:
		table = "solana"
	case pb.Chain_KARDIA:
		table = "kardia"
	default:
		err = ErrUnknownChain
	}
	return
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
