/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// recoverCmd represents the recover command
var recoverCmd = &cobra.Command{
	Use:   "recover",
	Short: "cover master private key from shared keys",
	Long:  `cover master private key from shared keys`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("recover called")
		// keygenConfig := config.EmptyConfig(curve.Secp256k1{})
		// // bytes, err := keygenConfig.ECDSA.MarshalBinary()

		// core.FailOnErr(err, "keygen fail")
	},
}

func init() {
	rootCmd.AddCommand(recoverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// recoverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// recoverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
