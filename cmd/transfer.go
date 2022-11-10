/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"time"

	"github.com/0x1be20/cmp-example/src/core"
	unit "github.com/DeOne4eg/eth-unit-converter"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/spf13/cobra"
	"github.com/taurusgroup/multi-party-sig/pkg/math/curve"
	"github.com/taurusgroup/multi-party-sig/pkg/pool"
	"github.com/taurusgroup/multi-party-sig/protocols/cmp/config"
)

var dist string
var send bool

// transferCmd represents the transfer command
var transferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "a transfer sample",
	Long:  `transfer 0.01 ether to dist address`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("dist address %s", dist)
		log.Printf("wallet %s", wallet)
		bytes, err := ioutil.ReadFile(wallet)
		keygenConfig := config.EmptyConfig(curve.Secp256k1{})
		// keygenConfig.
		err = keygenConfig.UnmarshalBinary(bytes)
		if err != nil {
			log.Panic(err)
		}

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

		id := keygenConfig.ID
		parties := keygenConfig.PartyIDs()
		n := core.NewWSeNetwork(string(id))
		n.Init()
		time.Sleep(time.Second * 10)

		signers := parties[:]

		pl := pool.NewPool(0)
		defer pl.TearDown()

		// 构建signature
		// 验证签名
		client, err := ethclient.Dial("https://rpc.ankr.com/eth_goerli")

		var chainId int64 = 5
		txSigner := types.NewLondonSigner(big.NewInt(chainId))
		amount := unit.NewEther(big.NewFloat(0.01))
		gasLimit := uint64(22000)
		// gasPrice, err := client.SuggestGasPrice(context.Background())
		// log.Printf("gas price %d", gasPrice.Int64())
		// core.FailOnErr(err, "get gasprice fail")
		gasPrice := big.NewInt(66083460)
		gasPrice = gasPrice.Mul(gasPrice, big.NewInt(5)).Div(gasPrice, big.NewInt(4))
		core.FailOnErr(err, "fail to get gas price")
		nonce, err := client.PendingNonceAt(context.Background(), address)
		log.Printf("nonce %d", nonce)
		core.FailOnErr(err, "fail to get nonce")
		var data []byte = []byte("helloworld")
		to := common.HexToAddress(dist) //0x921B004dc386ba15604bB97205Bb20988192DEDf
		tx := types.NewTransaction(nonce, to, amount.Value, gasLimit, gasPrice, data)
		hash := txSigner.Hash(tx)
		message := hash.Bytes()
		log.Printf("message %s", hash.Hex())
		// 得到R S
		signature, err := core.CMPSign(keygenConfig, message, signers, n, pl)
		if err != nil {
			core.FailOnErr(err, "")
		}
		log.Printf("ID:%s CMPSign ok %+v", string(id), signature)

		xPoint := (signature.R).(*curve.Secp256k1Point)
		xBytes = xPoint.XBytes()
		sBytes, err := signature.S.MarshalBinary()

		var (
			secp256k1N, _  = new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
			secp256k1halfN = new(big.Int).Div(secp256k1N, big.NewInt(2))
		)

		s := &big.Int{}
		s.SetBytes(sBytes)
		if s.Cmp(secp256k1halfN) == 1 {
			s = s.Sub(s, secp256k1halfN)
			log.Printf("S>half N")
		}
		sBytes = s.Bytes()

		log.Printf("r hex:%s", hex.EncodeToString(xBytes))
		log.Printf("s hex:%s", hex.EncodeToString(sBytes))

		yBytes = signature.R.(*curve.Secp256k1Point).YBytes()
		var v int64 = 0
		// 得到V
		// log.Printf("v=%d", v)
		for _, v = range []int64{0, 1} {
			log.Printf("===============================\n\n\n")
			log.Printf("v=%d", v)
			rawSignature := fmt.Sprintf("%s%s", hex.EncodeToString(xBytes),
				hex.EncodeToString(sBytes),
			)
			if v == 0 {
				rawSignature += "00"
			} else {
				rawSignature += "01"
			}
			log.Printf("tx rawsignature %s", rawSignature)
			signatureHex, err := hex.DecodeString(rawSignature)
			core.FailOnErr(err, "signature hex failed")
			signedTx, err := tx.WithSignature(txSigner, signatureHex)
			core.FailOnErr(err, "sign tx failed")

			tv, tr, ts := signedTx.RawSignatureValues()
			log.Printf("\ntxv %s  \ntxr %s \ntxs %s", hex.EncodeToString(tv.Bytes()), hex.EncodeToString(tr.Bytes()), hex.EncodeToString(ts.Bytes()))

			rawTxBytes, _ := rlp.EncodeToBytes(signedTx)
			rawTxHex := hex.EncodeToString(rawTxBytes)
			fmt.Printf("raw tx hex  %s\n", rawTxHex)
			fmt.Printf("tx hash %s\n", signedTx.Hash().Hex())
			msg, err := signedTx.AsMessage(types.NewLondonSigner(signedTx.ChainId()), nil)
			if err != nil {
				log.Printf("tx as mesg fail %+v", err)
				continue
			}
			fmt.Printf("tx from:%s\n", msg.From().Hex())
			if msg.From().Hex() == address.Hex() {
				if send {
					err = client.SendTransaction(context.Background(), signedTx)
					core.FailOnErr(err, "send tx failed")
					fmt.Printf("tx sent:%s", signedTx.Hash().Hex())
					break
				}
			}
		}

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
