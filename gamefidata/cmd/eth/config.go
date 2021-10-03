package eth

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/gametaverse/gamefidata/spider"
)

type configInfo struct {
	LogFile        string            `toml:"LogFile"`
	LogPath        string            `toml:"LogPath"`
	LogLevel       string            `toml:"LogLevel"`
	PrivKey        string            `toml:"PrivKey"`
	DBURI          string            `toml:"DBURI"`
	Games          []spider.GameInfo `toml:"Games"`
	Chain          string            `toml:"Chain"`
	ChainID        int               `toml:"ChainID"`
	RPCAddr        string            `toml:"RPCAddr"`
	BottomBlock    uint64            `toml:"BottomBlock"`
	Interval       int               `toml:"Interval"`
	BackwardFactor int               `toml:"BackwardFactor"`
}

var config configInfo

func loadConfig(fp string) (err error) {
	if _, err = toml.DecodeFile(fp, &config); err != nil {
		fmt.Printf("Decode Config Error:%s \n", err.Error())
		return
	}

	fmt.Printf("Load Config file %s Success \n", fp)
	err = nil
	return
}
