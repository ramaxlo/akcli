package cmd

import "github.com/spf13/cobra"

var yoctoCmd = &cobra.Command{
	Use:   "yocto",
	Short: "Yocto build environment commands",
}

func init() {
	rootCmd.AddCommand(yoctoCmd)
}
