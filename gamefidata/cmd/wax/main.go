package wax

import (
	"fmt"

	"github.com/cz-theng/czkit-go/log"
	"github.com/gametaverse/gamefidata/db"
	"github.com/gametaverse/gamefidata/spider"
	"github.com/spf13/cobra"
)

var (
	_configFile string
	_initDB     bool
)

var CMD = &cobra.Command{
	Use:   "wax",
	Short: "wax project",
	Long:  `wax project`,
	Run:   _main,
}

func init() {
	CMD.PersistentFlags().StringVarP(&_configFile, "config", "c", "", "config file path for matcha")

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

	privKeyOpt := spider.WithPrivKey(spider.Config.PrivKey)
	mongoURIOpt := spider.WithMongoURI(spider.Config.DBURI)
	chainOpt := spider.WithChain(spider.Config.Chain)
	chainIDOpt := spider.WithChainID(spider.Config.ChainID)
	rpcAddrOpt := spider.WithRPCAddr(spider.Config.RPCAddr)
	bottomBlockOpt := spider.WithBottomBlock(spider.Config.BottomBlock)
	fintervalOpt := spider.WithForwardInterval(spider.Config.ForwardInterval)
	bintervalOpt := spider.WithBackwardInterval(spider.Config.BackwardInterval)
	fworksOpt := spider.WithForwardWorks(spider.Config.ForwardWorks)
	bworksOpt := spider.WithBackwardWorks(spider.Config.BackwardWorks)

	if _initDB {
		err := db.CreateAndInitDB(spider.Config.DBURI)
		if err != nil {
			log.Error("DB error:%s", err.Error())
		}
		return
	}

	spiderApp := spider.New()
	err := spiderApp.Init(privKeyOpt,
		chainOpt,
		chainIDOpt,
		rpcAddrOpt,
		bottomBlockOpt,
		mongoURIOpt,
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
