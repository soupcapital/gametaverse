package main

import (
	"fmt"

	"github.com/cz-theng/czkit-go/log"
	"github.com/gametaverse/gametaversebot"
	"github.com/spf13/cobra"
)

var (
	_version    bool
	_configFile string
)

var rootCMD = &cobra.Command{
	Use:   "gametaversebot",
	Short: "start an gametaversebot",
	Long:  `start an gametaversebot`,
	Run:   _main,
}

func init() {
	rootCMD.PersistentFlags().BoolVarP(&_version, "version", "v", false, "print version of test")
	rootCMD.PersistentFlags().StringVarP(&_configFile, "config", "c", "", "config file path for test")
}

func main() {
	rootCMD.Execute()
}

func dumpVersion() {
	fmt.Printf("%s\n", "0.1.0")
}

func _main(cmd *cobra.Command, args []string) {
	if _version {
		dumpVersion()
		return
	}
	if len(_configFile) > 0 {
		if err := loadConfig(_configFile); err != nil {
			return
		}
	} else {
		cmd.Usage()
		return
	}

	logFile := "test.log"
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

	tokenOpt := gametaversebot.WithTGBotToken(config.TelegramToken)
	robotOpt := gametaversebot.WithRobot(config.Robot)
	groupsOpt := gametaversebot.WithGroups(config.Groups)
	rpcAddrOpt := gametaversebot.WithRPCAddr(config.RPCAddr)
	dbURIOpt := gametaversebot.WithMongoURI(config.DBURI)
	gametaversebot.Init(
		robotOpt,
		tokenOpt,
		groupsOpt,
		rpcAddrOpt,
		dbURIOpt,
	)
	gametaversebot.StartService()

}
