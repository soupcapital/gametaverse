package main

import (
	"fmt"

	"github.com/gametaverse/twitterspy/cmd/api"
	"github.com/gametaverse/twitterspy/cmd/digger"
	"github.com/gametaverse/twitterspy/cmd/spider"
	"github.com/gametaverse/twitterspy/cmd/token"
	"github.com/spf13/cobra"
)

var (
	_version bool
)

var rootCMD = &cobra.Command{
	Use:   "twitterspy",
	Short: "start an twitterspy",
	Long:  `start an twitterspy`,
	Run:   _main,
}

func init() {
	rootCMD.PersistentFlags().BoolVarP(&_version, "version", "v", false, "print version of twitterspy")

	rootCMD.AddCommand(spider.CMD)
	rootCMD.AddCommand(api.CMD)
	rootCMD.AddCommand(digger.CMD)
	rootCMD.AddCommand(token.CMD)
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
