package daily

import (
	"log"
	"time"

	"github.com/gametaverse/gamefidata/daily"
	"github.com/gametaverse/gamefidata/utils"
	"github.com/spf13/cobra"
)

var (
	_mongoURL string
	_rpcURL   string
	_day      string
)

var CMD = &cobra.Command{
	Use:   "daily",
	Short: "daily project",
	Long:  `daily project`,
	Run:   _main,
}

func init() {
	CMD.PersistentFlags().StringVarP(&_mongoURL, "mongo", "m", "", "mongdb url")
	CMD.PersistentFlags().StringVarP(&_rpcURL, "rpc", "r", "", "rpc url")
	CMD.PersistentFlags().StringVarP(&_day, "day", "d", "", "day to query ")
}

func _main(cmd *cobra.Command, args []string) {
	if len(_day) == 0 ||
		len(_mongoURL) == 0 ||
		len(_rpcURL) == 0 {
		cmd.Usage()
		return
	}
	task := daily.NewDailyTask()
	if err := task.Init(_mongoURL, _rpcURL); err != nil {
		log.Printf("init error:%v", err)
		return
	}
	dayTs, err := time.Parse(utils.DateFormat, _day)
	if err != nil {
		log.Printf("parse day error")
		return
	}
	err = task.QueryHau(dayTs)
	if err != nil {
		log.Printf("query hau error:%s", err.Error())
		return
	}
}
