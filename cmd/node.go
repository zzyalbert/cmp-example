/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"encoding/hex"
	"log"
	"time"

	"github.com/0x1be20/cmp-example/src/client"
	"github.com/0x1be20/cmp-example/src/common"
	"github.com/0x1be20/cmp-example/src/communication"
	"github.com/spf13/cobra"
	"github.com/taurusgroup/multi-party-sig/pkg/protocol"
)

var mainNode bool

// nodeCmd represents the node command
var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "start as a node",
	Long:  `start as a node`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("run as a node")
		c := &client.Client{}
		c.Load(wallet)

		id := c.ID
		n := communication.NewWSeNetwork(sessionId, string(id), func(m *common.Message) {
			if m.Type == common.MesgTypeReqSign {
				// 签名
				go func(m *common.Message) {
					message := m.ExtraData
					signature, err := c.Sign(message, m.RequestId)
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
					// signatureStr := xStr + "|" + sStr
					// log.Printf("signature string %s", signatureStr)
					// signatureHex := []byte(signatureStr)

					if mainNode {
						mesg := common.Message{
							Type:      common.MesgTypeSignResult,
							SessionId: sessionId,
							RequestId: m.RequestId,
							MesgId:    0,
							ExtraData: signatureHex,
							Data:      &protocol.Message{},
						}
						c.Network().Send(&mesg)
					}
				}(m)

			}
		})
		n.Init("localhost:8080", "/ws")
		c.AddNetwork(n)
		c.Register(sessionId)

		// 应该注册处理函数，如果有什么消息过来，等待处理
		for {
			time.Sleep(time.Second)
		}
	},
}

func init() {
	rootCmd.AddCommand(nodeCmd)
	nodeCmd.PersistentFlags().BoolVarP(&mainNode, "main", "m", false, "main node") //(&dist, "dist", "d", "", "dist address")

	// nodeCmd.PersistentFlags().StringVarP(&session, "session", "", "0x921B004dc386ba15604bB97205Bb20988192DEDf", "dist address")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// nodeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// nodeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
