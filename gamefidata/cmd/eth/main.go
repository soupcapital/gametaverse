package eth

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
	Use:   "eth",
	Short: "eth project",
	Long:  `eth project`,
	Run:   _main,
}

func init() {
	CMD.PersistentFlags().StringVarP(&_configFile, "config", "c", "", "config file path for matcha")

	CMD.PersistentFlags().BoolVarP(&_initDB, "initdb", "d", false, "init database")

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

	gamsOpt := spider.WithGames(config.Games)
	privKeyOpt := spider.WithPrivKey(config.PrivKey)
	mongoURIOpt := spider.WithMongoURI(config.DBURI)
	chainOpt := spider.WithChain(config.Chain)
	chainIDOpt := spider.WithChainID(config.ChainID)
	rpcAddrOpt := spider.WithRPCAddr(config.RPCAddr)
	bottomBlockOpt := spider.WithBottomBlock(config.BottomBlock)
	fintervalOpt := spider.WithForwardInterval(config.ForwardInterval)
	bintervalOpt := spider.WithBackwardInterval(config.BackwardInterval)
	fworksOpt := spider.WithForwardWorks(config.ForwardWorks)
	bworksOpt := spider.WithBackwardWorks(config.BackwardWorks)

	if _initDB {
		err := db.CreateAndInitDB(config.DBURI)
		if err != nil {
			log.Error("DB error:%s", err.Error())
		}
		return
	}

	spiderApp := spider.New()
	err := spiderApp.Init(privKeyOpt,
		gamsOpt,
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
