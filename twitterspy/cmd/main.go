package main

import (
	"fmt"

	"github.com/gametaverse/twitterspy/cmd/api"
	"github.com/gametaverse/twitterspy/cmd/spider"
	"github.com/spf13/cobra"
)

var (
	_version bool
)

var rootCMD = &cobra.Command{
	Use:   "tttspy",
	Short: "start an tttspy",
	Long:  `start an tttspy`,
	Run:   _main,
}

func init() {
	rootCMD.PersistentFlags().BoolVarP(&_version, "version", "v", false, "print version of tttspy")

	rootCMD.AddCommand(spider.CMD)
	rootCMD.AddCommand(api.CMD)
	// rootCMD.AddCommand(pancakeswap.CMD)
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

}
