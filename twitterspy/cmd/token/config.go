package token

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type configInfo struct {
	RPCAddr  string `toml:"RPCAddr"`
	LogFile  string `toml:"LogFile"`
	LogPath  string `toml:"LogPath"`
	LogLevel string `toml:"LogLevel"`
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
