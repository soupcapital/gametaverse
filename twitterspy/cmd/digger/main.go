package digger

import (
	"fmt"

	"github.com/cz-theng/czkit-go/log"
	"github.com/gametaverse/twitterspy/digger"
	"github.com/spf13/cobra"
)

var (
	_configFile string
)

var CMD = &cobra.Command{
	Use:   "digger",
	Short: "start an digger",
	Long:  `start an digger`,
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

	logFile := "digger.log"
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

	err := digger.Init(config.DBURI, config.TokenRPC)
	if err != nil {
		fmt.Printf("init error:%s", err.Error())
		return
	}
	digger.Start()
}
