package digger

import (
	"fmt"

	"github.com/cz-theng/czkit-go/log"
	"github.com/gametaverse/twitterspy/digger"
	"github.com/spf13/cobra"
)

var (
	_rpcAddr      string
	_twitterCount uint32
)

var CMD = &cobra.Command{
	Use:   "digger",
	Short: "start an digger",
	Long:  `start an digger`,
	Run:   _main,
}

func init() {
	CMD.PersistentFlags().StringVarP(&_rpcAddr, "rpc", "r", "", "rpc addr of vname")
	CMD.PersistentFlags().Uint32VarP(&_twitterCount, "count", "c", 100, "count of twitter to digg")
}

func _main(cmd *cobra.Command, args []string) {

	if len(_rpcAddr) == 0 {
		cmd.Usage()
		return
	}

	logFile := "digger.log"
	logPath := "./"

	logNameOpt := log.WithLogName(logFile)
	logPathOpt := log.WithLogPath(logPath)
	log.Init(logNameOpt, logPathOpt)

	err := digger.Init(_rpcAddr, _twitterCount)
	if err != nil {
		fmt.Printf("init error:%s", err.Error())
		return
	}
	digger.Start()
}
