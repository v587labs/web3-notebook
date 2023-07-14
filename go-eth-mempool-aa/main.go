package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"flag"
	"fmt"
	log2 "github.com/cometbft/cometbft/libs/log"
	bftos "github.com/cometbft/cometbft/libs/os"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	"math/big"
	"net/url"
	"os"
	"runtime"
	"strings"

	_ "embed"
)

const HandleOpsSign = "1fad948c"

var wsRpcVar = "wss://localhost:8080"
var originVar = ""
var privateKeyVar = ""

//go:embed abi.json
var abiJsonString []byte
var jsonABI abi.ABI
var gasVar uint64 = 10
var gasPriceVar uint64 = 10

func init() {
	flag.StringVar(&wsRpcVar, "ws", wsRpcVar, "-ws "+wsRpcVar)
	flag.StringVar(&originVar, "origin", originVar, "-origin "+originVar)
	flag.StringVar(&privateKeyVar, "key", privateKeyVar, "-key "+privateKeyVar)
	flag.Uint64Var(&gasVar, "gas", gasVar, fmt.Sprintf("-gas %d", gasVar))
	flag.Uint64Var(&gasPriceVar, "price", gasPriceVar, fmt.Sprintf("-price %d", gasPriceVar))
	log.Root().SetHandler(
		log.LvlFilterHandler(log.LvlDebug, log.StdoutHandler),
	)

	jsonABI, _ = abi.JSON(bytes.NewReader(abiJsonString))

}

var pendingTransactions chan common.Hash
var privateKey *ecdsa.PrivateKey
var fromAddress common.Address

func main() {
	flag.Parse()
	log.Debug("start", "ws prc", wsRpcVar)
	parse, err := url.Parse(wsRpcVar)
	if err != nil {
		log.Crit("ws url parse error", "err", err)
		return
	}

	ctx := context.Background()
	if len(originVar) < 1 {
		originVar = parse.Host
	}
	ws, err := rpc.DialWebsocket(ctx, wsRpcVar, originVar)
	if err != nil {
		log.Crit("dial error", "err", err)
		return
	}
	//load key

	privateKey, err = crypto.HexToECDSA(privateKeyVar)
	if err != nil {
		log.Crit("key load error", "err", err)
		return
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Crit("error casting public key to ECDSA")
		return
	}

	fromAddress = crypto.PubkeyToAddress(*publicKeyECDSA)
	// new geth
	gethClient := gethclient.New(ws)
	pendingTransactions = make(chan common.Hash, 100)
	subscribePendingTransactions, err := gethClient.SubscribePendingTransactions(ctx, pendingTransactions)
	if err != nil {
		log.Crit("subscribePendingTransactions error", "err", err)
		return
	}
	go ReadPendingTransaction(ws)
	bftos.TrapSignal(log2.NewTMLogger(os.Stdout), func() {
		close(pendingTransactions)
		subscribePendingTransactions.Unsubscribe()
		ws.Close()
	})
	select {}
}

func ReadPendingTransaction(ws *rpc.Client) {
	numCPU := runtime.NumCPU()
	for i := 0; i < numCPU; i++ {
		go _readPendingTransaction(ws)
	}
}

func _readPendingTransaction(ws *rpc.Client) {

	ethClient := ethclient.NewClient(ws)
	chainID, err := ethClient.ChainID(context.Background())
	if err != nil {
		return
	}
	for tx := range pendingTransactions {
		transaction, _, err := ethClient.TransactionByHash(context.Background(), tx)
		if err != nil {
			continue
		}

		signer := types.LatestSignerForChainID(chainID)
		sender, err := types.Sender(signer, transaction)

		if sender == fromAddress {
			continue
		}

		log.Debug("find transaction", "hash", tx.Hex())
		data := transaction.Data()
		if len(data) > 4 {
			head := strings.ToLower(hex.EncodeToString(data[:4]))
			if head == HandleOpsSign {

				method, err := jsonABI.MethodById(data[0:4])
				if err != nil {
					log.Error("method error", "err", err)
					continue
				}

				unpack, err := method.Inputs.Unpack(data[4:])
				if err != nil {
					log.Error("Unpack error", "err", err)
					continue
				}

				gasPrice := transaction.GasPrice()
				gasPrice = gasPrice.Add(gasPrice, big.NewInt(int64(gasPriceVar)))

				unpack[1] = fromAddress
				//fmt.Println(unpack)
				dataPack, err := method.Inputs.Pack(unpack...)
				if err != nil {
					log.Error("pack error", "err", err)
					continue
				}

				nonce, err := ethClient.PendingNonceAt(context.Background(), fromAddress)
				if err != nil {
					log.Error("get nonce error", "err", err)
					continue
				}
				tx := &types.LegacyTx{
					Nonce:    nonce,
					GasPrice: gasPrice,
					Gas:      transaction.Gas() + gasVar,
					To:       transaction.To(),
					Value:    big.NewInt(0),
					Data:     append(method.ID, dataPack...),
				}
				newTx := types.NewTx(tx)
				newTx.GasPrice()

				signTx, err := types.SignTx(newTx, types.NewEIP2930Signer(chainID), privateKey)
				if err != nil {
					log.Error("sign tx error", "err", err)
					continue
				}

				err = ethClient.SendTransaction(context.Background(), signTx)
				if err != nil {
					log.Error("send tx error", "err", err)
					continue
				}
				log.Info("send success", "hash", signTx.Hash().Hex())
			}

		}

	}
}
