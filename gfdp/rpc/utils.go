package rpc

import "github.com/gametaverse/gfdp/rpc/pb"

func getTableName(chain pb.Chain) (table string, err error) {
	switch chain {
	case pb.Chain_BSC:
		table = "t_tx_bsc"
	case pb.Chain_ETH:
		table = "t_tx_eth"
	case pb.Chain_POLYGON:
		table = "t_tx_polygon"
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
