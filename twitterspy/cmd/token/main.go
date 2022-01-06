package token

import (
	"github.com/cz-theng/czkit-go/log"
	"github.com/gametaverse/twitterspy/token"
	"github.com/spf13/cobra"
)

var (
	_configFile string
)

var CMD = &cobra.Command{
	Use:   "token",
	Short: "start an token",
	Long:  `start an token`,
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

	logFile := "token.log"
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

	rpcAddrOpt := token.WithListenAddr(config.RPCAddr)
	svr := token.NewServer()
	if err := svr.Init(rpcAddrOpt); err != nil {
		log.Error("Init server error:%s", err.Error())
		return
	}
	if err := svr.Run(); err != nil {
		log.Error("run server error:%s", err.Error())
		return
	}

}
