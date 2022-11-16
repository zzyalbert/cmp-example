/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"encoding/hex"
	"log"
	"time"

	"github.com/0x1be20/cmp-example/src/common"
	"github.com/0x1be20/cmp-example/src/communication"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/taurusgroup/multi-party-sig/pkg/party"
	"github.com/taurusgroup/multi-party-sig/pkg/protocol"
)

// apiCmd represents the api command
var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("run as a node")

		stop := make(chan int, 1)

		n := communication.NewWSeNetwork(sessionId, string(id), func(m *common.Message) {
			log.Printf("got message type %s", m.Type)
			if m.Type == common.MesgTypeSignResult {
				signature := &common.Signature{}
				err := signature.Unmarshal(m.ExtraData)
				if err != nil {
					log.Printf("extra signature fail %+v", err)
				}

				log.Printf("got sign result %s %s", hex.EncodeToString(signature.S), hex.EncodeToString(signature.X))
				stop <- 1

			}
		})
		n.Init("localhost:8080", "/ws")

		time.Sleep(time.Second * 5)

		mesg := &common.Message{
			Type:      common.MesgTypeReqSign,
			SessionId: sessionId,
			MesgId:    0,
			RequestId: uuid.NewString(),
			ExtraData: []byte("helloworld"),
			Data:      &protocol.Message{},
		}
		n.Send(mesg)
		<-stop
		n.Done(party.ID(id))

	},
}

func init() {
	rootCmd.AddCommand(apiCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// apiCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// apiCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
