package spider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cz-theng/czkit-go/log"
	"github.com/gametaverse/gfdp/db"
	"github.com/gametaverse/gfdp/utils"
)

const (
	_getSoltFmt  = `{"jsonrpc":"2.0","id":1, "method":"getSlot"}`
	_getBlockFmt = `{"jsonrpc": "2.0","id":1,"method":"getBlock","params":[%d, {"encoding": "json","transactionDetails":"full","rewards":false}]}`
)

type TransactionInfo struct {
	Message struct {
		AccountKeys []string `json:"accountKeys"`
		Header      struct {
			NumReadonlySignedAccounts   int `json:"numReadonlySignedAccounts"`
			NumReadonlyUnsignedAccounts int `json:"numReadonlyUnsignedAccounts"`
			NumRequiredSignatures       int `json:"numRequiredSignatures"`
		} `json:"header"`
		Instructions []struct {
			Accounts       []int  `json:"accounts"`
			Data           string `json:"data"`
			ProgramIDIndex int    `json:"programIdIndex"`
		} `json:"instructions"`
		RecentBlockhash string `json:"recentBlockhash"`
	} `json:"message"`
	Signatures []string `json:"signatures"`
}

type BlockInfo struct {
	BlockHeight       int    `json:"blockHeight"`
	BlockTime         int    `json:"blockTime"`
	Blockhash         string `json:"blockhash"`
	ParentSlot        int    `json:"parentSlot"`
	PreviousBlockhash string `json:"previousBlockhash"`
	Transactions      []struct {
		Meta struct {
			Err               interface{}   `json:"err"`
			Fee               int           `json:"fee"`
			InnerInstructions []interface{} `json:"innerInstructions"`
			LogMessages       []string      `json:"logMessages"`
			PostBalances      []interface{} `json:"postBalances"`
			PostTokenBalances []interface{} `json:"postTokenBalances"`
			PreBalances       []interface{} `json:"preBalances"`
			PreTokenBalances  []interface{} `json:"preTokenBalances"`
			Rewards           []interface{} `json:"rewards"`
			Status            struct {
				Ok interface{} `json:"Ok"`
			} `json:"status"`
		} `json:"meta"`
		Transaction TransactionInfo `json:"transaction"`
	} `json:"transactions"`
}

type SolAntenna struct {
	cli     *http.Client
	rpcAddr string
	chainID int
}

func NewSolanaAntenna() *SolAntenna {
	antenna := &SolAntenna{}
	return antenna
}

func (ata *SolAntenna) Init(rpc string, chainID int) (err error) {
	log.Info("SolAntenna init")
	ata.cli = &http.Client{}
	ata.rpcAddr = rpc
	return
}

func (ata *SolAntenna) GetBlockHeight(ctx context.Context) (height uint64, err error) {
	URL := ata.rpcAddr
	reqBodyBuf := []byte(_getSoltFmt)
	req, err := http.NewRequest("POST", URL, bytes.NewReader(reqBodyBuf))
	if err != nil {
		log.Info("new request error:%s", err.Error())
		return
	}

	req.ContentLength = int64(len(reqBodyBuf))
	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")

	resp, err := ata.cli.Do(req)
	if err != nil {
		log.Error("request error:%s", err.Error())
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
		return
	}

	respJSON := struct {
		Jsonrpc string `json:"jsonrpc"`
		Result  int    `json:"result"`
		ID      int    `json:"id"`
	}{}

	bodyDecoder := json.NewDecoder(resp.Body)
	if err = bodyDecoder.Decode(&respJSON); err != nil {
		log.Error("request decode error:%s", err.Error())
		return
	}

	height = uint64(respJSON.Result)
	err = nil
	return
}

func (ata *SolAntenna) BlockByNumber(ctx context.Context, num uint64) (block *BlockInfo, err error) {

	URL := ata.rpcAddr
	reqBodyBuf := []byte(fmt.Sprintf(_getBlockFmt, num))
	req, err := http.NewRequest("POST", URL, bytes.NewReader(reqBodyBuf))
	if err != nil {
		log.Info("new request error:%s", err.Error())
		return
	}

	req.ContentLength = int64(len(reqBodyBuf))
	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	resp, err := ata.cli.Do(req)
	if err != nil {
		log.Error("request error:%s", err.Error())
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
		return
	}

	respJSON := struct {
		Jsonrpc string    `json:"jsonrpc"`
		Result  BlockInfo `json:"result"`
		ID      int       `json:"id"`
	}{}

	bodyDecoder := json.NewDecoder(resp.Body)
	if err = bodyDecoder.Decode(&respJSON); err != nil {
		log.Error("request decode error:%s", err.Error())
		return
	}
	block = &respJSON.Result
	//log.Info("%v ts: %v", num, block.BlockTime)
	err = nil
	return
}

func (ata *SolAntenna) GetTrxByNum(ctx context.Context, num uint64) (trxes []*Transaction, err error) {
	blk, err := ata.BlockByNumber(ctx, num)
	if err != nil {
		if strings.Contains(err.Error(), "non-empty transaction list but block header indicates no transactions") {
			return nil, nil
		}
		return nil, err
	}
	for i, trx := range blk.Transactions {
		trxInfo := trx.Transaction
		trxes = append(trxes, &Transaction{
			timestamp: utils.StartSecForDay(uint64(blk.BlockTime)),
			raw:       &trxInfo,
			block:     num,
			index:     uint16(i),
		})
	}
	return

}

func (ata *SolAntenna) DealTrx(rawtrx *Transaction) (txes []*db.Transaction, err error) {
	trx, ok := rawtrx.raw.(*TransactionInfo)
	if !ok {
		err = ErrUnknownTrx
		return
	}

	for i, instr := range trx.Message.Instructions {
		if i > 256 {
			continue
		}

		if len(instr.Accounts) == 0 {
			continue
		}
		if instr.ProgramIDIndex >= len(trx.Message.AccountKeys) || instr.Accounts[0] >= len(trx.Message.AccountKeys) {
			continue
		}
		progID := trx.Message.AccountKeys[instr.ProgramIDIndex]
		from := trx.Message.AccountKeys[instr.Accounts[0]]
		tx := &db.Transaction{
			Block:     rawtrx.block,
			Index:     (rawtrx.index << 8) | uint16(i),
			Timestamp: time.Unix(int64(rawtrx.timestamp), 0),
			From:      strings.ToLower(from),
			To:        strings.ToLower(progID),
		}
		txes = append(txes, tx)
	}

	err = nil
	return
}
