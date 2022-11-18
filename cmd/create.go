/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"time"

	"github.com/0x1be20/cmp-example/src/common"
	"github.com/0x1be20/cmp-example/src/communication"
	"github.com/spf13/cobra"
	"github.com/taurusgroup/multi-party-sig/pkg/protocol"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new wallet",
	Long:  `create a new network`,
	Run: func(cmd *cobra.Command, args []string) {

		n := communication.NewWSeNetwork(sessionId, func(m *common.Message) {

		})

		n.Init("localhost:8080", "/ws")

		// m := &common.Message{
		// 	Type:      common.MesgTypeRegister,
		// 	NodeType:  common.NodeTypeUser,
		// 	SessionId: sessionId,
		// 	MesgId:    0,
		// 	RequestId: id,
		// 	ExtraData: []byte{},
		// 	Data:      &protocol.Message{},
		// }
		// n.Send(m)

		m := &common.Message{
			Type:      common.MesgTypeUserRegister,
			NodeType:  "",
			SessionId: sessionId,
			MesgId:    0,
			RequestId: "",
			ExtraData: []byte{},
			Data:      &protocol.Message{},
		}
		n.Send(m)

		for {
			time.Sleep(time.Second)
		}

		// c := client.NewClient(partyId, n)
		// c.AddNetwork(n)

		// //等待一会
		// time.Sleep(time.Second * 10)

		// keygenConfig, err := c.Keygen(parties, parties.Len()-1, "", sessionId)
		// if err != nil {
		// 	core.FailOnErr(err, "")
		// }
		// bytes, err := keygenConfig.MarshalBinary()
		// walletPath := fmt.Sprintf("%s_%s.kgc", sessionId, "user")
		// ioutil.WriteFile(walletPath, bytes, 0777)
		// if err != nil {
		// 	core.FailOnErr(err, "")
		// }
		// log.Printf("ID:%s CMPKeygen ok", string(id))

		// pb := keygenConfig.PublicPoint().(*curve.Secp256k1Point)
		// xBytes := pb.XBytes()
		// x := &big.Int{}
		// x.SetBytes(xBytes)
		// yBytes := pb.YBytes()
		// y := &big.Int{}
		// y.SetBytes(yBytes)

		// publicKey := &ecies.PublicKey{
		// 	X:     x,
		// 	Y:     y,
		// 	Curve: ecies.DefaultCurve,
		// }

		// address := crypto.PubkeyToAddress(*publicKey.ExportECDSA())
		// log.Printf("address:%s", address.Hex())
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
