package client

import (
	"errors"
	"io/ioutil"
	"log"
	"math/big"

	"github.com/0x1be20/cmp-example/src/communication"
	"github.com/0x1be20/cmp-example/src/sdk"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/taurusgroup/multi-party-sig/pkg/ecdsa"
	"github.com/taurusgroup/multi-party-sig/pkg/math/curve"
	"github.com/taurusgroup/multi-party-sig/pkg/party"
	"github.com/taurusgroup/multi-party-sig/pkg/pool"
	"github.com/taurusgroup/multi-party-sig/protocols/cmp"
	"github.com/taurusgroup/multi-party-sig/protocols/cmp/config"
)

type Client struct {
	N       communication.ProtocolNetwork
	ID      party.ID
	config  *cmp.Config
	Address common.Address
}

func NewClient(partyId party.ID, network communication.ProtocolNetwork) *Client {
	return &Client{
		N:  network,
		ID: partyId,
	}
}

func (c *Client) Load(wallet string) {
	bytes, err := ioutil.ReadFile(wallet)
	keygenConfig := config.EmptyConfig(curve.Secp256k1{})

	err = keygenConfig.UnmarshalBinary(bytes)
	if err != nil {
		log.Panic(err)
	}

	c.config = keygenConfig
	c.ID = c.config.ID

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
	c.Address = address

	log.Printf("address:%s", address.Hex())
}

func (c *Client) Keygen(ids party.IDSlice, threshold int) (*cmp.Config, error) {
	pl := pool.NewPool(0)
	defer pl.TearDown()
	result, _, err := sdk.CMPKeygen(c.ID, ids, threshold, c.N, pl)
	if err != nil {
		log.Printf("keygen failed,%+v", err)
		return nil, err
	}
	r := <-result
	config := r.(*cmp.Config)
	return config, nil
}

func (c *Client) Sign(m []byte) (*ecdsa.Signature, error) {
	pl := pool.NewPool(0)
	defer pl.TearDown()
	signers := c.config.PartyIDs()
	result, _, err := sdk.CMPSign(c.config, m, signers, c.N, pl)
	if err != nil {
		log.Printf("sign failed,%+v", err)
		return nil, err
	}
	r := <-result
	if r != nil {
		signature := r.(*ecdsa.Signature)
		return signature, nil
	}
	return nil, errors.New("signature failed")

}
