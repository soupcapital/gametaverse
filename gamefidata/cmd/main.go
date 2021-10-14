package main

import (
	"fmt"

	"github.com/gametaverse/gamefidata/cmd/api"
	"github.com/gametaverse/gamefidata/cmd/eth"
	"github.com/gametaverse/gamefidata/cmd/wax"
	"github.com/spf13/cobra"
)

var (
	_version bool
)

var rootCMD = &cobra.Command{
	Use:   "gamefidata",
	Short: "start an gamefidata",
	Long:  `start an gamefidata`,
	Run:   _main,
}

func init() {
	rootCMD.PersistentFlags().BoolVarP(&_version, "version", "v", false, "print version of gamefidata")

	rootCMD.AddCommand(eth.CMD)
	rootCMD.AddCommand(api.CMD)
	rootCMD.AddCommand(wax.CMD)
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
