package spider

import (
	"fmt"

	"github.com/cz-theng/czkit-go/log"
	"github.com/gametaverse/gfdp/spider"
	"github.com/spf13/cobra"
)

var (
	_configFile string
	_initDB     bool
)

var CMD = &cobra.Command{
	Use:   "spider",
	Short: "spider project",
	Long:  `spider project`,
	Run:   _main,
}

func init() {
	CMD.PersistentFlags().StringVarP(&_configFile, "config", "c", "", "config file path for spider")

	CMD.PersistentFlags().BoolVarP(&_initDB, "initdb", "d", false, "init database")

}

func _main(cmd *cobra.Command, args []string) {

	if len(_configFile) > 0 {
		if err := spider.LoadConfig(_configFile); err != nil {
			return
		}
	} else {
		cmd.Usage()
		return
	}

	logFile := "spider.log"
	logPath := "./"
	if len(spider.Config.LogPath) > 0 {
		logPath = spider.Config.LogPath
	}
	if len(spider.Config.LogFile) > 0 {
		logFile = spider.Config.LogFile
	}

	logNameOpt := log.WithLogName(logFile)
	logPathOpt := log.WithLogPath(logPath)
	log.Init(logNameOpt, logPathOpt)

	chainOpt := spider.WithChain(spider.Config.Chain)
	chainIDOpt := spider.WithChainID(spider.Config.ChainID)
	rpcAddrOpt := spider.WithRPCAddr(spider.Config.RPCAddr)
	bottomBlockOpt := spider.WithBottomBlock(spider.Config.BottomBlock)
	fintervalOpt := spider.WithForwardInterval(spider.Config.ForwardInterval)
	bintervalOpt := spider.WithBackwardInterval(spider.Config.BackwardInterval)
	fworksOpt := spider.WithForwardWorks(spider.Config.ForwardWorks)
	bworksOpt := spider.WithBackwardWorks(spider.Config.BackwardWorks)

	spiderApp := spider.New()
	err := spiderApp.Init(
		chainOpt,
		chainIDOpt,
		rpcAddrOpt,
		bottomBlockOpt,
		fintervalOpt,
		bintervalOpt,
		fworksOpt,
		bworksOpt,
	)
	if err != nil {
		fmt.Printf("Init error:%s \n", err.Error())
		return
	}

	spiderApp.Run()
}
