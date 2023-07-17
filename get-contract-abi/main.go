package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	"math/big"
	"os"
)

var rpcVar = ""
var blockNumberVar = "latest"
var addressVar = "latest"
var chainID *big.Int

func init() {
	flag.StringVar(&rpcVar, "rpc", rpcVar, "-rpc "+rpcVar)
	flag.StringVar(&blockNumberVar, "block", blockNumberVar, "-block "+blockNumberVar)
	flag.StringVar(&addressVar, "address", addressVar, "-address "+addressVar)
	log.Root().SetHandler(
		log.LvlFilterHandler(log.LvlDebug, log.StdoutHandler),
	)
}
func main() {
	ctx := context.Background()
	flag.Parse()

	if !common.IsHexAddress(addressVar) {
		log.Error("请设置正确的地址")
		return
	}

	rpcClient, err := rpc.Dial(rpcVar)
	if err != nil {
		log.Error("dial error", "err", err)
		return
	}
	ethClient := ethclient.NewClient(rpcClient)
	chainID, err = ethClient.ChainID(ctx)
	if err != nil {
		log.Error("chain id get error", "err", err)
		return
	}
	log.Debug("start get code", "chainId", chainID.String())
	//var blockNumber *big.Int
	//if blockNumberVar == "latest" {
	//	_blockNumber, err := ethClient.BlockNumber(ctx)
	//	if err != nil {
	//		log.Error("BlockNumber get error", "err", err)
	//		return
	//	}
	//	log.Debug("get BlockNumber success", "blockNumber", _blockNumber)
	//	blockNumber = big.NewInt(int64(_blockNumber))
	//
	//} else {
	//	_blockNumber, err := hexutil.DecodeUint64(blockNumberVar)
	//	if err != nil {
	//		log.Debug("parse BlockNumber success", "blockNumber", _blockNumber)
	//		return
	//	}
	//	blockNumber = big.NewInt(int64(_blockNumber))
	//}
	bytes, err := ethClient.CodeAt(ctx, common.HexToAddress(addressVar), nil)
	if err != nil {
		log.Error("get account code error", "err", err)
		return
	}

	contract := vm.Contract{Code: bytes}
	file, err := os.OpenFile("contract.bin", os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Error("gopen error", "err", err)
		return
	}
	defer file.Close()
	write, err := file.Write(bytes)
	if err != nil {
		log.Error("write error", "err", err, "write", write)
		return
	}

	fmt.Println(contract)
}
