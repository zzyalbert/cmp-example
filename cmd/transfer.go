/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	"github.com/0x1be20/cmp-example/src/client"
	cmpcommon "github.com/0x1be20/cmp-example/src/common"
	"github.com/0x1be20/cmp-example/src/communication"
	"github.com/0x1be20/cmp-example/src/core"
	unit "github.com/DeOne4eg/eth-unit-converter"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/taurusgroup/multi-party-sig/pkg/protocol"
)

var dist string
var send bool

func txSigner() types.Signer {
	var chainId int64 = 5
	txSigner := types.NewLondonSigner(big.NewInt(chainId))
	return txSigner
}

func txClient() *ethclient.Client {
	client, _ := ethclient.Dial("https://rpc.ankr.com/eth_goerli")
	return client
}

func buildUnsignedTx(address common.Address) (*types.Transaction, []byte) {
	client := txClient()
	signer := txSigner()
	// amount,gaslimit,gasprice
	amount := unit.NewEther(big.NewFloat(0.01))
	gasLimit := uint64(22000)
	gasPrice := big.NewInt(66083460)
	gasPrice = gasPrice.Mul(gasPrice, big.NewInt(5)).Div(gasPrice, big.NewInt(4))

	// nonce
	nonce, err := client.PendingNonceAt(context.Background(), address)
	core.FailOnErr(err, "fail to get nonce")

	// message
	var data []byte = []byte("helloworld")
	to := common.HexToAddress(dist) //0x921B004dc386ba15604bB97205Bb20988192DEDf
	tx := types.NewTransaction(nonce, to, amount.Value, gasLimit, gasPrice, data)
	hash := signer.Hash(tx)
	message := hash.Bytes()
	log.Printf("message %s", hash.Hex())

	return tx, message
}

func buildSignedTx(xBytes []byte, sBytes []byte, tx *types.Transaction, address common.Address) *types.Transaction {

	signer := txSigner()

	// var sBytes []byte
	// var xPoint *curve.Secp256k1Point
	s := &big.Int{}

	// xPoint = (signature.R).(*curve.Secp256k1Point)
	// xBytes := xPoint.XBytes()
	// sBytes, _ = signature.S.MarshalBinary()

	var (
		secp256k1N, _  = new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
		secp256k1halfN = new(big.Int).Div(secp256k1N, big.NewInt(2))
	)

	s.SetBytes(sBytes)
	if s.Cmp(secp256k1halfN) == 1 {
		s = secp256k1N.Sub(secp256k1N, s)
	}

	sBytes = s.Bytes()

	log.Printf("r hex:%s", hex.EncodeToString(xBytes))
	log.Printf("s hex:%s", hex.EncodeToString(sBytes))

	var v int64 = 0
	// 得到V
	for _, v = range []int64{0, 1} {
		rawSignature := fmt.Sprintf("%s%s", hex.EncodeToString(xBytes),
			hex.EncodeToString(sBytes),
		)
		if v == 0 {
			rawSignature += "00"
		} else {
			rawSignature += "01"
		}
		signatureHex, err := hex.DecodeString(rawSignature)
		core.FailOnErr(err, "signature hex failed")
		signedTx, err := tx.WithSignature(signer, signatureHex)
		core.FailOnErr(err, "sign tx failed")

		msg, err := signedTx.AsMessage(types.NewLondonSigner(signedTx.ChainId()), nil)
		if err != nil {
			log.Printf("tx as mesg fail %+v", err)
			continue
		}
		if msg.From().Hex() == address.Hex() {
			fmt.Printf("tx hash %s\n", signedTx.Hash().Hex())
			fmt.Printf("tx from:%s\n", msg.From().Hex())
			return signedTx
		}

	}
	return nil

}

// transferCmd represents the transfer command
var transferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "a transfer sample",
	Long:  `transfer 0.01 ether to dist address`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("dist address %s", dist)
		log.Printf("wallet %s", wallet)

		signatureChan := make(chan *cmpcommon.Signature, 1)

		ethClient := txClient()

		c := &client.Client{}
		c.Load(wallet)

		id := c.ID
		n := communication.NewWSeNetwork(sessionId, string(id), func(m *cmpcommon.Message) {
			log.Printf("got message type %s", m.Type)
			if m.Type == cmpcommon.MesgTypeSignResult {
				signature := &cmpcommon.Signature{}
				err := signature.Unmarshal(m.ExtraData)
				if err != nil {
					log.Printf("extra signature fail %+v", err)
				}

				log.Printf("got sign result %s %s", hex.EncodeToString(signature.S), hex.EncodeToString(signature.X))
				signatureChan <- signature

			}
		})
		n.Init("localhost:8080", "/ws")
		c.AddNetwork(n)

		// c.Register(sessionId)

		// time.Sleep(time.Second * 10)

		tx, message := buildUnsignedTx(c.Address)

		mesg := &cmpcommon.Message{
			Type:      cmpcommon.MesgTypeReqSign,
			SessionId: sessionId,
			MesgId:    0,
			RequestId: uuid.NewString(),
			ExtraData: message,
			Data:      &protocol.Message{},
		}
		n.Send(mesg)

		signature := <-signatureChan

		// signature, err := c.Sign(message, "")
		// if err != nil {
		// 	core.FailOnErr(err, "")
		// }
		// log.Printf("ID:%s CMPSign ok %+v", string(id), signature)

		signedTx := buildSignedTx(signature.X, signature.S, tx, c.Address)
		err := ethClient.SendTransaction(context.Background(), signedTx)
		core.FailOnErr(err, "send tx failed")
		fmt.Printf("tx sent:%s", signedTx.Hash().Hex())
	},
}

func init() {
	rootCmd.AddCommand(transferCmd)
	transferCmd.PersistentFlags().StringVarP(&dist, "dist", "d", "0x921B004dc386ba15604bB97205Bb20988192DEDf", "dist address")
	transferCmd.PersistentFlags().BoolVarP(&send, "send", "", false, "") //(&dist, "dist", "d", "", "dist address")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// transferCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// transferCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
