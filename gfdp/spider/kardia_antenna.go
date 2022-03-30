package spider

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/cz-theng/czkit-go/log"
	"github.com/gametaverse/gfdp/db"
	"github.com/kardiachain/go-kardia/rpc"
)

type KardiaLog struct {
	Address          string   `json:"address"`
	Topics           []string `json:"topics"`
	Data             string   `json:"data"`
	TransactionIndex int      `json:"transactionIndex"`
	LogIndex         int      `json:"logIndex"`
}

type KardiaReceipt struct {
	TransactionHash   string      `json:"transactionHash"`
	GasUsed           int         `json:"gasUsed"`
	CumulativeGasUsed int         `json:"cumulativeGasUsed"`
	ContractAddress   string      `json:"contractAddress"`
	Logs              []KardiaLog `json:"logs"`
	Status            int         `json:"status"`
}

type KardiaTransaction struct {
	BlockHash        string    `json:"blockHash"`
	BlockNumber      int       `json:"blockNumber"`
	Time             time.Time `json:"time"`
	From             string    `json:"from"`
	Gas              int       `json:"gas"`
	GasPrice         int64     `json:"gasPrice"`
	Hash             string    `json:"hash"`
	Input            string    `json:"input"`
	Nonce            int       `json:"nonce"`
	To               string    `json:"to"`
	TransactionIndex int       `json:"transactionIndex"`
	Value            string    `json:"value"`
}

type KardiaBlock struct {
	Hash     string              `json:"hash"`
	Height   int                 `json:"height"`
	Time     time.Time           `json:"time"`
	NumTxs   int                 `json:"numTxs"`
	GasLimit int                 `json:"gasLimit"`
	GasUsed  int                 `json:"gasUsed"`
	Txs      []KardiaTransaction `json:"txs"`
	Receipts []KardiaReceipt     `json:"receipts"`
}

type KardiaAntenna struct {
	cli     *rpc.Client
	chainID int
}

func NewKardiaAntennaAntenna() *KardiaAntenna {
	antenna := &KardiaAntenna{}
	return antenna
}

func (ata *KardiaAntenna) Init(rpcAddr string, chainID int) (err error) {
	log.Info("ETHAntenna init")
	ata.chainID = chainID
	ata.cli, err = rpc.Dial(rpcAddr)
	if err != nil {
		log.Error("Dial error:%s", err.Error())
		return err
	}
	return
}

func (ata *KardiaAntenna) GetBlockHeight(ctx context.Context) (height uint64, err error) {
	var blockHeight rpc.BlockHeight
	err = ata.cli.CallContext(ctx, &blockHeight, "kai_blockNumber")
	if err != nil {
		return
	}
	log.Info("block height:%v", blockHeight)
	height = blockHeight.Uint64()
	return
}

func (ata *KardiaAntenna) GetTrxByNum(ctx context.Context, num uint64) (trxes []*Transaction, err error) {
	var block KardiaBlock
	err = ata.cli.CallContext(ctx, &block, "kai_getBlockByNumber", num)
	if err != nil {
		if strings.Contains(err.Error(), "non-empty transaction list but block header indicates no transactions") {
			return nil, nil
		}
		return nil, err
	}
	if block.Height == 0 {
		return nil, errors.New("not found")
	}
	for i, trx := range block.Txs {
		trxes = append(trxes, &Transaction{
			timestamp: uint64(block.Time.Unix()),
			block:     num,
			index:     uint16(i),
			raw:       trx,
		})
	}
	return
}

func (ata *KardiaAntenna) DealTrx(rawtrx *Transaction) (txes []*db.Transaction, err error) {
	trx, ok := rawtrx.raw.(KardiaTransaction)
	if !ok {
		return nil, ErrUnknownTrx
	}
	if len(trx.To) == 0 {
		return // done with 0x0000...000
	}

	to := trx.To
	from := trx.From
	tx := &db.Transaction{
		Block:     rawtrx.block,
		Index:     rawtrx.index,
		Timestamp: time.Unix(int64(rawtrx.timestamp), 0),
		From:      strings.ToLower(from),
		To:        strings.ToLower(to),
	}
	txes = append(txes, tx)

	return
}
