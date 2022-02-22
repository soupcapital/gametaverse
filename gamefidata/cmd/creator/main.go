package creator

import (
	"github.com/spf13/cobra"
)

var (
	_addr  string
	_chain string
	_proxy string
)

var CMD = &cobra.Command{
	Use:   "creator",
	Short: "creator project",
	Long:  `creator project`,
	Run:   _main,
}

func init() {
	CMD.PersistentFlags().StringVarP(&_addr, "addr", "a", "", "the address of creator")
	CMD.PersistentFlags().StringVarP(&_chain, "chain", "c", "", "the name of chain should be 'bsc' 'polygon' or 'eth' ")
	CMD.PersistentFlags().StringVarP(&_proxy, "proxy", "p", "", "http proxy address ")

	//CMD.PersistentFlags().BoolVarP(&_initDB, "initdb", "d", false, "init database")

}

func _main(cmd *cobra.Command, args []string) {
	if len(_addr) == 0 ||
		len(_chain) == 0 {
		cmd.Usage()
		return
	}
	process(_addr, _chain, _proxy)
}
