/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"io/ioutil"
	"log"
	"math/big"
	"time"

	"github.com/0x1be20/cmp-example/src/client"
	"github.com/0x1be20/cmp-example/src/common"
	"github.com/0x1be20/cmp-example/src/communication"
	"github.com/0x1be20/cmp-example/src/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/spf13/cobra"
	"github.com/taurusgroup/multi-party-sig/pkg/math/curve"
	"github.com/taurusgroup/multi-party-sig/pkg/party"
)

var id string
var ids []string
var threshold int64

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new wallet",
	Long:  `create a new network`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("wallet path:%s", wallet)
		log.Printf("wallet id:%s", id)
		log.Printf("%+v", ids)
		log.Printf("threshold:%d", threshold)

		partyId := party.ID(id)
		partyIds := []party.ID{}
		for _, p := range ids {
			partyIds = append(partyIds, party.ID(p))
		}
		parties := party.NewIDSlice(partyIds)

		n := communication.NewWSeNetwork(sessionId, string(partyId), func(m *common.Message) {})

		n.Init("localhost:8080", "/ws")

		c := client.NewClient(partyId, n)

		//等待一会
		time.Sleep(time.Second * 10)

		keygenConfig, err := c.Keygen(parties, int(threshold), "")
		if err != nil {
			core.FailOnErr(err, "")
		}
		bytes, err := keygenConfig.MarshalBinary()
		ioutil.WriteFile(wallet, bytes, 0777)
		if err != nil {
			core.FailOnErr(err, "")
		}
		log.Printf("ID:%s CMPKeygen ok", string(id))

		pb := keygenConfig.PublicPoint().(*curve.Secp256k1Point)
		xBytes := pb.XBytes()
		x := &big.Int{}
		x.SetBytes(xBytes)
		yBytes := pb.YBytes()
		y := &big.Int{}
		y.SetBytes(yBytes)

		publicKey := &ecies.PublicKey{
			X:     x,
			Y:     y,
			Curve: ecies.DefaultCurve,
		}

		address := crypto.PubkeyToAddress(*publicKey.ExportECDSA())
		log.Printf("address:%s", address.Hex())
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringSliceVarP(&ids, "ids", "s", []string{}, "")
	createCmd.Flags().Int64VarP(&threshold, "threshold", "t", 0, "threshold")
	rootCmd.PersistentFlags().StringVarP(&id, "id", "i", "", "wallet id")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
