package api

import (
	"fmt"

	"github.com/cz-theng/czkit-go/log"
	"github.com/gametaverse/gfdp/rpc"
	"github.com/spf13/cobra"
)

var (
	_configFile string
)

var CMD = &cobra.Command{
	Use:   "api",
	Short: "api project",
	Long:  `api project`,
	Run:   _main,
}

func init() {
	CMD.PersistentFlags().StringVarP(&_configFile, "config", "c", "", "config file path for matcha")

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
	listenAddrOpt := rpc.WithListenAddr(config.ListenAddr)
	dbAddrOpt := rpc.WithDbUrl(config.DbUrl)
	dbUserOpt := rpc.WithDbUser(config.DbUser)
	dbPasswdOpt := rpc.WithDbPasswd(config.DbPasswd)
	dbNameOpt := rpc.WithDbName(config.DbName)

	app := rpc.NewServer()
	err := app.Init(
		dbAddrOpt,
		listenAddrOpt,
		dbUserOpt,
		dbPasswdOpt,
		dbNameOpt,
	)
	if err != nil {
		fmt.Printf("Init error:%s \n", err.Error())
		return
	}

	app.Run()
}
