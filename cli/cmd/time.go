package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var timeCmd = &cobra.Command{
	Use:   "time",
	Short: "Print current timestamp in user's local timezone",
	Long:  "Outputs the current date/time formatted for the user's locale and timezone. Used by skills and plans for consistent timestamps.",
	Run: func(cmd *cobra.Command, args []string) {
		now := time.Now()
		fmt.Println(now.Format("02/01/2006 15:04"))
	},
}
