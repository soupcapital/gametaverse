package main

import (
	"gitee.com/c_z/cti/db"
	"github.com/spf13/cobra"
)

var (
	_dbInitFlag bool
)

var dbCMD = &cobra.Command{
	Use:   "db",
	Short: "setup mongodb",
	Long:  `setup mongodb`,
	Run:   _dbMain,
}

func init() {
	dbCMD.PersistentFlags().BoolVarP(&_dbInitFlag, "init", "i", false, "create and init mongodb")

}

func _dbMain(cmd *cobra.Command, args []string) {
	if _dbInitFlag {
		db.CreateAndInitDB("cz", "Solong2020")
		return
	}
}
