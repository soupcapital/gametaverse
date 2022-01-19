package spider

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type ContractType string

const (
	ContractERC20  ContractType = "erc20"
	ContractERC721 ContractType = "erc721"
	ContractGame   ContractType = "game"
	ContractOther  ContractType = "other"
)

type GameInfo struct {
	Name      string   `toml:"Name"`
	ID        string   `toml:"ID"`
	Contracts []string `toml:"Contracts"`
}

type ConfigInfo struct {
	LogFile          string  `toml:"LogFile"`
	LogPath          string  `toml:"LogPath"`
	LogLevel         string  `toml:"LogLevel"`
	PrivKey          string  `toml:"PrivKey"`
	DBURI            string  `toml:"DBURI"`
	Chain            string  `toml:"Chain"`
	ChainID          int     `toml:"ChainID"`
	RPCAddr          string  `toml:"RPCAddr"`
	BottomBlock      uint64  `toml:"BottomBlock"`
	ForwardInterval  float32 `toml:"ForwardInterval"`
	BackwardInterval float32 `toml:"BackwardInterval"`
	ForwardWorks     int     `toml:"ForwardWorks"`
	BackwardWorks    int     `toml:"BackwardWorks"`
}

var Config ConfigInfo

func LoadConfig(fp string) (err error) {
	if _, err = toml.DecodeFile(fp, &Config); err != nil {
		fmt.Printf("Decode Config Error:%s \n", err.Error())
		return
	}

	fmt.Printf("Load Config file %s Success \n", fp)
	err = nil
	return
}
