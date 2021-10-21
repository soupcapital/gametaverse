package cti

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/cz-theng/czkit-go/log"
)

const (
	UNISwapAPI = "https://api.thegraph.com/subgraphs/name/uniswap/uniswap-v2"
)

type UniswapToken struct {
	Symbol    string
	ID        string
	Name      string
	Liquidity float64
}

type UniswapPair struct {
	ID string
}

var (
	_getTokensFmt = `{"operationName":"tokens","variables":{"value":"%s","id":"ETH"},"query":"query tokens($value: String, $id: String) {\n  asSymbol: tokens(where: {symbol: $value}, orderBy: totalLiquidity, orderDirection: desc) {\n    id\n    symbol\n    name\n    totalLiquidity\n    __typename\n  }\n  }\n"}`
	_getPairsFmt  = `{"operationName":"pairs","variables":{"tokens":[%s],"id":"ET"},"query":"query pairs($tokens: [Bytes]!, $id: String) {\n  as0: pairs(where: {token0_in: $tokens}) {\n    id\n token0Price\n token1Price\n   token0 {\n      id\n      symbol\n      name\n      __typename\n    }\n    token1 {\n      id\n      symbol\n      name\n      __typename\n    }\n    __typename\n  }\n  as1: pairs(where: {token1_in: $tokens}) {\n    id\n token0Price\n token1Price\n    token0 {\n      id\n      symbol\n      name\n      __typename\n    }\n    token1 {\n      id\n      symbol\n      name\n      __typename\n    }\n    __typename\n  }\n  }\n"}`
)

func GetTokens(symbol string) ([]UniswapToken, error) {
	reqBodyBuf := []byte(fmt.Sprintf(_getTokensFmt, symbol))
	cli := &http.Client{}

	req, err := http.NewRequest("POST", UNISwapAPI, bytes.NewReader(reqBodyBuf))
	if err != nil {
		return nil, err
	}
	req.ContentLength = int64(len(reqBodyBuf))
	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	if r, err := rand.Int(rand.Reader, big.NewInt(int64(len(userAgents)))); err == nil {
		userAgent := userAgents[r.Int64()]
		req.Header.Add("User-Agent", userAgent)
	}
	// do request
	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New(resp.Status)
	}
	rspDecoder := json.NewDecoder(resp.Body)
	rspJSON := &struct {
		Data struct {
			AsSymbol []struct {
				Typename       string `json:"__typename"`
				ID             string `json:"id"`
				Name           string `json:"name"`
				Symbol         string `json:"symbol"`
				TotalLiquidity string `json:"totalLiquidity"`
			} `json:"asSymbol"`
		} `json:"data"`
	}{}
	if err = rspDecoder.Decode(rspJSON); err != nil {
		log.Error("request error:%s", err.Error())
		return nil, err
	}

	var tokens []UniswapToken
	for _, s := range rspJSON.Data.AsSymbol {
		l, _ := strconv.ParseFloat(s.TotalLiquidity, 64)
		t := UniswapToken{
			Symbol:    s.Symbol,
			Name:      s.Name,
			ID:        s.ID,
			Liquidity: l,
		}
		tokens = append(tokens, t)
	}

	return tokens, nil
}

func GetParis(tokens []string) ([]UniswapPair, error) {
	ts := ""
	split := ""
	for _, t := range tokens {
		ts += split + `"` + t + `"`
		split = ","
	}
	reqBodyBuf := []byte(fmt.Sprintf(_getPairsFmt, ts))
	cli := &http.Client{}

	req, err := http.NewRequest("POST", UNISwapAPI, bytes.NewReader(reqBodyBuf))
	if err != nil {
		return nil, err
	}
	req.ContentLength = int64(len(reqBodyBuf))
	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	if r, err := rand.Int(rand.Reader, big.NewInt(int64(len(userAgents)))); err == nil {
		userAgent := userAgents[r.Int64()]
		req.Header.Add("User-Agent", userAgent)
	}
	// do request
	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New(resp.Status)
	}
	rspDecoder := json.NewDecoder(resp.Body)
	rspJSON := &struct {
		Data struct {
			As0 []struct {
				Typename string `json:"__typename"`
				ID       string `json:"id"`
				Token0   struct {
					Typename string `json:"__typename"`
					ID       string `json:"id"`
					Name     string `json:"name"`
					Symbol   string `json:"symbol"`
				} `json:"token0"`
				Token0Price string `json:"token0Price"`
				Token1      struct {
					Typename string `json:"__typename"`
					ID       string `json:"id"`
					Name     string `json:"name"`
					Symbol   string `json:"symbol"`
				} `json:"token1"`
				Token1Price string `json:"token1Price"`
			} `json:"as0"`
			As1 []struct {
				Typename string `json:"__typename"`
				ID       string `json:"id"`
				Token0   struct {
					Typename string `json:"__typename"`
					ID       string `json:"id"`
					Name     string `json:"name"`
					Symbol   string `json:"symbol"`
				} `json:"token0"`
				Token0Price string `json:"token0Price"`
				Token1      struct {
					Typename string `json:"__typename"`
					ID       string `json:"id"`
					Name     string `json:"name"`
					Symbol   string `json:"symbol"`
				} `json:"token1"`
				Token1Price string `json:"token1Price"`
			} `json:"as1"`
		} `json:"data"`
	}{}
	if err = rspDecoder.Decode(rspJSON); err != nil {
		log.Error("request error:%s", err.Error())
		return nil, err
	}

	var pairs []UniswapPair
	for _, s := range rspJSON.Data.As0 {
		t := UniswapPair{
			ID: s.ID,
		}
		pairs = append(pairs, t)
	}

	for _, s := range rspJSON.Data.As1 {
		t := UniswapPair{
			ID: s.ID,
		}
		pairs = append(pairs, t)
	}

	return pairs, nil
}
