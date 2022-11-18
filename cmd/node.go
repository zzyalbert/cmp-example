/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"encoding/hex"
	"fmt"
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
	"github.com/taurusgroup/multi-party-sig/pkg/protocol"
)

var mainNode bool
var role string

func getWallet(wallet string, n communication.ProtocolNetwork) *client.Client {
	c := &client.Client{}
	c.Load(wallet)
	c.AddNetwork(n)
	return c
}

func getRoleWallet(sessionId string, role common.NodeType, n communication.ProtocolNetwork) *client.Client {
	walletPath := fmt.Sprintf("%s_%s.kgc", sessionId, role)
	return getWallet(walletPath, n)
}

// nodeCmd represents the node command
var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "start as a node",
	Long:  `start as a node`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("run as a node")

		var n *communication.WSNetwork

		n = communication.NewWSeNetwork(sessionId, func(m *common.Message) {
			log.Printf("rec mesg,type %s nodetype %s sessionId %s", m.Type, m.NodeType, m.SessionId)
			if m.Type == common.MesgTypeReqSign {
				// 根据sessionId获取wallet
				c := getRoleWallet(m.SessionId, common.NodeType(role), n)

				// 签名
				go func(m *common.Message, c *client.Client) {
					message := m.ExtraData
					signature, err := c.Sign(message, m.RequestId, m.SessionId)
					if err != nil {
						log.Printf("sign fail,%+v", err)
						return
					}
					xBytes, _ := signature.R.XScalar().MarshalBinary()
					sBytes, _ := signature.S.MarshalBinary()
					xStr := hex.EncodeToString(xBytes)
					sStr := hex.EncodeToString(sBytes)
					log.Printf("%s %s", xStr, sStr)
					s := &common.Signature{
						S: sBytes,
						X: xBytes,
					}
					signatureHex, err := s.MarshalBinary()
					if err != nil {
						log.Printf("marshal signature fail,%+v", err)
						return
					}

					if role == string(common.NodeTypeUser) {
						mesg := common.Message{
							Type:      common.MesgTypeSignResult,
							NodeType:  common.NodeTypeUser,
							SessionId: m.SessionId,
							RequestId: m.RequestId,
							MesgId:    0,
							ExtraData: signatureHex,
							Data:      &protocol.Message{},
						}
						n.Send(&mesg)
					}
				}(m, c)

			} else if m.Type == common.MesgTypeUserRegister {
				go func(m *common.Message) {
					// 注册新的钱包
					c := client.NewClient(party.ID(role), n)
					c.AddNetwork(n)

					// if role == string(common.NodeTypeUser) {
					// 	c.Register(m.SessionId, common.NodeType(role))
					// }

					partyIds := []party.ID{}
					roles := []common.NodeType{
						common.NodeTypeAuditor,
						common.NodeTypeServer,
						common.NodeTypeUser,
					}
					for _, p := range roles {
						partyIds = append(partyIds, party.ID(p))
					}
					parties := party.NewIDSlice(partyIds)

					keygenConfig, err := c.Keygen(parties, parties.Len()-1, "", m.SessionId)
					if err != nil {
						core.FailOnErr(err, "")
					}
					bytes, err := keygenConfig.MarshalBinary()
					walletPath := fmt.Sprintf("%s_%s.kgc", m.SessionId, role)
					ioutil.WriteFile(walletPath, bytes, 0777)
					if err != nil {
						core.FailOnErr(err, "")
					}
					log.Printf("ID:%s CMPKeygen ok", string(role))

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
				}(m)

			} else if m.Type == common.MesgTypeWalletCreate {

			}
		})
		n.Init("localhost:8080", "/ws")

		m := common.Message{
			Type:      common.MesgTypeRegister,
			NodeType:  common.NodeType(role),
			SessionId: sessionId,
			MesgId:    0,
			RequestId: "",
			ExtraData: []byte{},
			Data:      &protocol.Message{},
		}
		n.Send(&m)

		// 应该注册处理函数，如果有什么消息过来，等待处理
		for {
			time.Sleep(time.Second)
		}
	},
}

func init() {
	rootCmd.AddCommand(nodeCmd)
	nodeCmd.PersistentFlags().BoolVarP(&mainNode, "main", "m", false, "main node") //(&dist, "dist", "d", "", "dist address")
	nodeCmd.PersistentFlags().StringVarP(&role, "role", "r", "user", "node role")  //(&dist, "dist", "d", "", "dist address")

	// nodeCmd.PersistentFlags().StringVarP(&session, "session", "", "0x921B004dc386ba15604bB97205Bb20988192DEDf", "dist address")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// nodeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// nodeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
