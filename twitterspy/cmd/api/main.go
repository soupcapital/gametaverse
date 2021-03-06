package api

import (
	"fmt"

	"github.com/cz-theng/czkit-go/log"
	"github.com/gametaverse/twitterspy/api"
	"github.com/spf13/cobra"
)

var (
	_configFile string
)

var CMD = &cobra.Command{
	Use:   "api",
	Short: "start an api",
	Long:  `start an api`,
	Run:   _main,
}

func init() {
	CMD.PersistentFlags().StringVarP(&_configFile, "config", "c", "", "config file path for test")
}

func _main(cmd *cobra.Command, args []string) {
	if len(_configFile) > 0 {
		if err := loadConfig(_configFile); err != nil {
			return
		}
	} else {
		cmd.Usage()
		return
	}

	logFile := "api.log"
	logPath := "./"
	if len(config.LogPath) > 0 {
		logPath = config.LogPath
	}
	if len(config.LogFile) > 0 {
		logFile = config.LogFile
	}

	logNameOpt := log.WithLogName(logFile)
	logPathOpt := log.WithLogPath(logPath)
	log.Init(logNameOpt, logPathOpt)

	mongoURIOpt := api.WithMongoURI(config.DBURI)
	listenAddrOpt := api.WithListenAddr(config.ListenAddr)
	tokenRPCOpt := api.WithTokenRPC(config.TokenRPC)
	apiApp := api.NewServer()
	err := apiApp.Init(
		mongoURIOpt,
		tokenRPCOpt,
		listenAddrOpt)
	if err != nil {
		fmt.Printf("Init error:%s \n", err.Error())
		return
	}

	apiApp.Run()
}
