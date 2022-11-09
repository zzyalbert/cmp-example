package core

import (
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/taurusgroup/multi-party-sig/pkg/ecdsa"
	"github.com/taurusgroup/multi-party-sig/pkg/math/curve"
	"github.com/taurusgroup/multi-party-sig/pkg/party"
	"github.com/taurusgroup/multi-party-sig/pkg/pool"
	"github.com/taurusgroup/multi-party-sig/pkg/protocol"
	"github.com/taurusgroup/multi-party-sig/pkg/taproot"
	"github.com/taurusgroup/multi-party-sig/protocols/cmp"
	"github.com/taurusgroup/multi-party-sig/protocols/cmp/config"
	"github.com/taurusgroup/multi-party-sig/protocols/example"
	"github.com/taurusgroup/multi-party-sig/protocols/frost"
)

func XOR(id party.ID, ids party.IDSlice, n *WSNetwork) error {
	h, err := protocol.NewMultiHandler(example.StartXOR(id, ids), nil)
	if err != nil {
		return err
	}
	HandlerLoopByQueue(id, h, n)
	_, err = h.Result()
	if err != nil {
		return err
	}
	return nil
}

func CMPKeygen(id party.ID, ids party.IDSlice, threshold int, n *WSNetwork, pl *pool.Pool) (*cmp.Config, error) {
	h, err := protocol.NewMultiHandler(cmp.Keygen(curve.Secp256k1{}, id, ids, threshold, pl), nil)
	if err != nil {
		return nil, err
	}
	HandlerLoopByQueue(id, h, n)
	r, err := h.Result()
	if err != nil {
		return nil, err
	}

	return r.(*cmp.Config), nil
}

func CMPRefresh(c *cmp.Config, n *WSNetwork, pl *pool.Pool) (*cmp.Config, error) {
	hRefresh, err := protocol.NewMultiHandler(cmp.Refresh(c, pl), nil)
	if err != nil {
		return nil, err
	}
	HandlerLoopByQueue(c.ID, hRefresh, n)

	r, err := hRefresh.Result()
	if err != nil {
		return nil, err
	}

	return r.(*cmp.Config), nil
}

func CMPSign(c *cmp.Config, m []byte, signers party.IDSlice, n *WSNetwork, pl *pool.Pool) (*ecdsa.Signature, error) {
	h, err := protocol.NewMultiHandler(cmp.Sign(c, signers, m, pl), nil)
	if err != nil {
		return nil, err
	}
	HandlerLoopByQueue(c.ID, h, n)

	signResult, err := h.Result()
	if err != nil {
		return nil, err
	}
	signature := signResult.(*ecdsa.Signature)
	if !signature.Verify(c.PublicPoint(), m) {
		return nil, errors.New("failed to verify cmp signature")
	}
	return signature, nil
}

func CMPPreSign(c *cmp.Config, signers party.IDSlice, n *WSNetwork, pl *pool.Pool) (*ecdsa.PreSignature, error) {
	h, err := protocol.NewMultiHandler(cmp.Presign(c, signers, pl), nil)
	if err != nil {
		return nil, err
	}

	HandlerLoopByQueue(c.ID, h, n)

	signResult, err := h.Result()
	if err != nil {
		return nil, err
	}

	preSignature := signResult.(*ecdsa.PreSignature)
	if err = preSignature.Validate(); err != nil {
		return nil, errors.New("failed to verify cmp presignature")
	}
	return preSignature, nil
}

func CMPPreSignOnline(c *cmp.Config, preSignature *ecdsa.PreSignature, m []byte, n *WSNetwork, pl *pool.Pool) error {
	h, err := protocol.NewMultiHandler(cmp.PresignOnline(c, preSignature, m, pl), nil)
	if err != nil {
		return err
	}
	HandlerLoopByQueue(c.ID, h, n)

	signResult, err := h.Result()
	if err != nil {
		return err
	}
	signature := signResult.(*ecdsa.Signature)
	if !signature.Verify(c.PublicPoint(), m) {
		return errors.New("failed to verify cmp signature")
	}
	return nil
}

func FrostKeygen(id party.ID, ids party.IDSlice, threshold int, n *WSNetwork) (*frost.Config, error) {
	h, err := protocol.NewMultiHandler(frost.Keygen(curve.Secp256k1{}, id, ids, threshold), nil)
	if err != nil {
		return nil, err
	}
	HandlerLoopByQueue(id, h, n)
	r, err := h.Result()
	if err != nil {
		return nil, err
	}

	return r.(*frost.Config), nil
}

func FrostSign(c *frost.Config, id party.ID, m []byte, signers party.IDSlice, n *WSNetwork) error {
	h, err := protocol.NewMultiHandler(frost.Sign(c, signers, m), nil)
	if err != nil {
		return err
	}
	HandlerLoopByQueue(id, h, n)
	r, err := h.Result()
	if err != nil {
		return err
	}

	signature := r.(frost.Signature)
	if !signature.Verify(c.PublicKey, m) {
		return errors.New("failed to verify frost signature")
	}
	return nil
}

func FrostKeygenTaproot(id party.ID, ids party.IDSlice, threshold int, n *WSNetwork) (*frost.TaprootConfig, error) {
	h, err := protocol.NewMultiHandler(frost.KeygenTaproot(id, ids, threshold), nil)
	if err != nil {
		return nil, err
	}
	HandlerLoopByQueue(id, h, n)
	r, err := h.Result()
	if err != nil {
		return nil, err
	}

	return r.(*frost.TaprootConfig), nil
}
func FrostSignTaproot(c *frost.TaprootConfig, id party.ID, m []byte, signers party.IDSlice, n *WSNetwork) error {
	h, err := protocol.NewMultiHandler(frost.SignTaproot(c, signers, m), nil)
	if err != nil {
		return err
	}
	HandlerLoopByQueue(id, h, n)
	r, err := h.Result()
	if err != nil {
		return err
	}

	signature := r.(taproot.Signature)
	if !c.PublicKey.Verify(signature, m) {
		return errors.New("failed to verify frost signature")
	}
	return nil
}

func demo() {
	cmdType := flag.String("type", "server", "类型")
	idStr := flag.String("id", "1", "ID")
	ids := party.IDSlice{"1", "2", "3"}
	keyPath := flag.String("wallet", "", "wallet path")

	flag.Parse()
	id := party.ID(*idStr)
	n := NewWSeNetwork(string(id))

	pl := pool.NewPool(0)
	defer pl.TearDown()

	threshold := 2

	if *cmdType == "client" {
		log.Println("run as client")
		var keygenConfig *config.Config = nil
		if *keyPath != "" {
			bytes, err := ioutil.ReadFile(*keyPath)
			keygenConfig = config.EmptyConfig(curve.Secp256k1{})
			// keygenConfig.
			err = keygenConfig.UnmarshalBinary(bytes)
			if err != nil {
				log.Panic(err)
			}
			id = keygenConfig.ID
			n = NewWSeNetwork(string(id))
			n.Init()
			time.Sleep(time.Second * 10)
		} else {
			//key generation
			// 10s之后开始xor流程
			time.Sleep(time.Second * 10)
			n := NewWSeNetwork(string(id))
			n.Init()
			keygenConfig, err := CMPKeygen(id, ids, threshold, n, pl)

			bytes, err := keygenConfig.MarshalBinary()
			ioutil.WriteFile(*keyPath, bytes, 0777)
			if err != nil {
				FailOnErr(err, "")
			}
			log.Printf("ID%s CMPKeygen ok", string(id))
			return
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

	} else {
		log.Println("run as server")
		InitServer()
	}
}
