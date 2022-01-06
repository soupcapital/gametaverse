package spider

import (
	"github.com/cz-theng/czkit-go/log"
	"github.com/gametaverse/twitterspy"
	"github.com/spf13/cobra"
)

var (
	_configFile string
)

var CMD = &cobra.Command{
	Use:   "spider",
	Short: "start an spider",
	Long:  `start an spider`,
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

	logFile := "spider.log"
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

	gropsOpt := twitterspy.WithGroups(config.Groups)
	vsOpt := twitterspy.WithVs(config.Vs)
	tokenOpt := twitterspy.WithTGBotToken(config.TelegramToken)
	internalOpt := twitterspy.WithTwitterInternal(config.TwitterInterval)
	countOpt := twitterspy.WithTwitterCount(config.TwitterCount)
	wordsOpt := twitterspy.WithKeyWords(config.KeyWords)
	mongoUrlOpt := twitterspy.WithMongoURI(config.MongoURI)
	tokenRPCOpt := twitterspy.WithTokenRPC(config.TokenRPC)

	twitterspy.Init(gropsOpt,
		vsOpt,
		tokenOpt,
		internalOpt,
		countOpt,
		mongoUrlOpt,
		tokenRPCOpt,
		wordsOpt)
	twitterspy.StartService()

}
