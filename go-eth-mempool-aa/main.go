package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	log2 "github.com/cometbft/cometbft/libs/log"
	bftos "github.com/cometbft/cometbft/libs/os"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	"net/url"
	"os"
	"runtime"
	"strings"
)

const ABIString = `[{"inputs":[{"internalType":"uint256","name":"preOpGas","type":"uint256"},{"internalType":"uint256","name":"paid","type":"uint256"},{"internalType":"uint48","name":"validAfter","type":"uint48"},{"internalType":"uint48","name":"validUntil","type":"uint48"},{"internalType":"bool","name":"targetSuccess","type":"bool"},{"internalType":"bytes","name":"targetResult","type":"bytes"}],"name":"ExecutionResult","type":"error"},{"inputs":[{"internalType":"uint256","name":"opIndex","type":"uint256"},{"internalType":"string","name":"reason","type":"string"}],"name":"FailedOp","type":"error"},{"inputs":[{"internalType":"address","name":"sender","type":"address"}],"name":"SenderAddressResult","type":"error"},{"inputs":[{"internalType":"address","name":"aggregator","type":"address"}],"name":"SignatureValidationFailed","type":"error"},{"inputs":[{"components":[{"internalType":"uint256","name":"preOpGas","type":"uint256"},{"internalType":"uint256","name":"prefund","type":"uint256"},{"internalType":"bool","name":"sigFailed","type":"bool"},{"internalType":"uint48","name":"validAfter","type":"uint48"},{"internalType":"uint48","name":"validUntil","type":"uint48"},{"internalType":"bytes","name":"paymasterContext","type":"bytes"}],"internalType":"struct IEntryPoint.ReturnInfo","name":"returnInfo","type":"tuple"},{"components":[{"internalType":"uint256","name":"stake","type":"uint256"},{"internalType":"uint256","name":"unstakeDelaySec","type":"uint256"}],"internalType":"struct IStakeManager.StakeInfo","name":"senderInfo","type":"tuple"},{"components":[{"internalType":"uint256","name":"stake","type":"uint256"},{"internalType":"uint256","name":"unstakeDelaySec","type":"uint256"}],"internalType":"struct IStakeManager.StakeInfo","name":"factoryInfo","type":"tuple"},{"components":[{"internalType":"uint256","name":"stake","type":"uint256"},{"internalType":"uint256","name":"unstakeDelaySec","type":"uint256"}],"internalType":"struct IStakeManager.StakeInfo","name":"paymasterInfo","type":"tuple"}],"name":"ValidationResult","type":"error"},{"inputs":[{"components":[{"internalType":"uint256","name":"preOpGas","type":"uint256"},{"internalType":"uint256","name":"prefund","type":"uint256"},{"internalType":"bool","name":"sigFailed","type":"bool"},{"internalType":"uint48","name":"validAfter","type":"uint48"},{"internalType":"uint48","name":"validUntil","type":"uint48"},{"internalType":"bytes","name":"paymasterContext","type":"bytes"}],"internalType":"struct IEntryPoint.ReturnInfo","name":"returnInfo","type":"tuple"},{"components":[{"internalType":"uint256","name":"stake","type":"uint256"},{"internalType":"uint256","name":"unstakeDelaySec","type":"uint256"}],"internalType":"struct IStakeManager.StakeInfo","name":"senderInfo","type":"tuple"},{"components":[{"internalType":"uint256","name":"stake","type":"uint256"},{"internalType":"uint256","name":"unstakeDelaySec","type":"uint256"}],"internalType":"struct IStakeManager.StakeInfo","name":"factoryInfo","type":"tuple"},{"components":[{"internalType":"uint256","name":"stake","type":"uint256"},{"internalType":"uint256","name":"unstakeDelaySec","type":"uint256"}],"internalType":"struct IStakeManager.StakeInfo","name":"paymasterInfo","type":"tuple"},{"components":[{"internalType":"address","name":"aggregator","type":"address"},{"components":[{"internalType":"uint256","name":"stake","type":"uint256"},{"internalType":"uint256","name":"unstakeDelaySec","type":"uint256"}],"internalType":"struct IStakeManager.StakeInfo","name":"stakeInfo","type":"tuple"}],"internalType":"struct IEntryPoint.AggregatorStakeInfo","name":"aggregatorInfo","type":"tuple"}],"name":"ValidationResultWithAggregation","type":"error"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"bytes32","name":"userOpHash","type":"bytes32"},{"indexed":true,"internalType":"address","name":"sender","type":"address"},{"indexed":false,"internalType":"address","name":"factory","type":"address"},{"indexed":false,"internalType":"address","name":"paymaster","type":"address"}],"name":"AccountDeployed","type":"event"},{"anonymous":false,"inputs":[],"name":"BeforeExecution","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"account","type":"address"},{"indexed":false,"internalType":"uint256","name":"totalDeposit","type":"uint256"}],"name":"Deposited","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"aggregator","type":"address"}],"name":"SignatureAggregatorChanged","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"account","type":"address"},{"indexed":false,"internalType":"uint256","name":"totalStaked","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"unstakeDelaySec","type":"uint256"}],"name":"StakeLocked","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"account","type":"address"},{"indexed":false,"internalType":"uint256","name":"withdrawTime","type":"uint256"}],"name":"StakeUnlocked","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"account","type":"address"},{"indexed":false,"internalType":"address","name":"withdrawAddress","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"}],"name":"StakeWithdrawn","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"bytes32","name":"userOpHash","type":"bytes32"},{"indexed":true,"internalType":"address","name":"sender","type":"address"},{"indexed":true,"internalType":"address","name":"paymaster","type":"address"},{"indexed":false,"internalType":"uint256","name":"nonce","type":"uint256"},{"indexed":false,"internalType":"bool","name":"success","type":"bool"},{"indexed":false,"internalType":"uint256","name":"actualGasCost","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"actualGasUsed","type":"uint256"}],"name":"UserOperationEvent","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"bytes32","name":"userOpHash","type":"bytes32"},{"indexed":true,"internalType":"address","name":"sender","type":"address"},{"indexed":false,"internalType":"uint256","name":"nonce","type":"uint256"},{"indexed":false,"internalType":"bytes","name":"revertReason","type":"bytes"}],"name":"UserOperationRevertReason","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"account","type":"address"},{"indexed":false,"internalType":"address","name":"withdrawAddress","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"}],"name":"Withdrawn","type":"event"},{"inputs":[],"name":"SIG_VALIDATION_FAILED","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"bytes","name":"initCode","type":"bytes"},{"internalType":"address","name":"sender","type":"address"},{"internalType":"bytes","name":"paymasterAndData","type":"bytes"}],"name":"_validateSenderAndPaymaster","outputs":[],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint32","name":"unstakeDelaySec","type":"uint32"}],"name":"addStake","outputs":[],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"address","name":"account","type":"address"}],"name":"balanceOf","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"account","type":"address"}],"name":"depositTo","outputs":[],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"address","name":"","type":"address"}],"name":"deposits","outputs":[{"internalType":"uint112","name":"deposit","type":"uint112"},{"internalType":"bool","name":"staked","type":"bool"},{"internalType":"uint112","name":"stake","type":"uint112"},{"internalType":"uint32","name":"unstakeDelaySec","type":"uint32"},{"internalType":"uint48","name":"withdrawTime","type":"uint48"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"account","type":"address"}],"name":"getDepositInfo","outputs":[{"components":[{"internalType":"uint112","name":"deposit","type":"uint112"},{"internalType":"bool","name":"staked","type":"bool"},{"internalType":"uint112","name":"stake","type":"uint112"},{"internalType":"uint32","name":"unstakeDelaySec","type":"uint32"},{"internalType":"uint48","name":"withdrawTime","type":"uint48"}],"internalType":"struct IStakeManager.DepositInfo","name":"info","type":"tuple"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"sender","type":"address"},{"internalType":"uint192","name":"key","type":"uint192"}],"name":"getNonce","outputs":[{"internalType":"uint256","name":"nonce","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"bytes","name":"initCode","type":"bytes"}],"name":"getSenderAddress","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"components":[{"internalType":"address","name":"sender","type":"address"},{"internalType":"uint256","name":"nonce","type":"uint256"},{"internalType":"bytes","name":"initCode","type":"bytes"},{"internalType":"bytes","name":"callData","type":"bytes"},{"internalType":"uint256","name":"callGasLimit","type":"uint256"},{"internalType":"uint256","name":"verificationGasLimit","type":"uint256"},{"internalType":"uint256","name":"preVerificationGas","type":"uint256"},{"internalType":"uint256","name":"maxFeePerGas","type":"uint256"},{"internalType":"uint256","name":"maxPriorityFeePerGas","type":"uint256"},{"internalType":"bytes","name":"paymasterAndData","type":"bytes"},{"internalType":"bytes","name":"signature","type":"bytes"}],"internalType":"struct UserOperation","name":"userOp","type":"tuple"}],"name":"getUserOpHash","outputs":[{"internalType":"bytes32","name":"","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[{"components":[{"components":[{"internalType":"address","name":"sender","type":"address"},{"internalType":"uint256","name":"nonce","type":"uint256"},{"internalType":"bytes","name":"initCode","type":"bytes"},{"internalType":"bytes","name":"callData","type":"bytes"},{"internalType":"uint256","name":"callGasLimit","type":"uint256"},{"internalType":"uint256","name":"verificationGasLimit","type":"uint256"},{"internalType":"uint256","name":"preVerificationGas","type":"uint256"},{"internalType":"uint256","name":"maxFeePerGas","type":"uint256"},{"internalType":"uint256","name":"maxPriorityFeePerGas","type":"uint256"},{"internalType":"bytes","name":"paymasterAndData","type":"bytes"},{"internalType":"bytes","name":"signature","type":"bytes"}],"internalType":"struct UserOperation[]","name":"userOps","type":"tuple[]"},{"internalType":"contract IAggregator","name":"aggregator","type":"address"},{"internalType":"bytes","name":"signature","type":"bytes"}],"internalType":"struct IEntryPoint.UserOpsPerAggregator[]","name":"opsPerAggregator","type":"tuple[]"},{"internalType":"address payable","name":"beneficiary","type":"address"}],"name":"handleAggregatedOps","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"components":[{"internalType":"address","name":"sender","type":"address"},{"internalType":"uint256","name":"nonce","type":"uint256"},{"internalType":"bytes","name":"initCode","type":"bytes"},{"internalType":"bytes","name":"callData","type":"bytes"},{"internalType":"uint256","name":"callGasLimit","type":"uint256"},{"internalType":"uint256","name":"verificationGasLimit","type":"uint256"},{"internalType":"uint256","name":"preVerificationGas","type":"uint256"},{"internalType":"uint256","name":"maxFeePerGas","type":"uint256"},{"internalType":"uint256","name":"maxPriorityFeePerGas","type":"uint256"},{"internalType":"bytes","name":"paymasterAndData","type":"bytes"},{"internalType":"bytes","name":"signature","type":"bytes"}],"internalType":"struct UserOperation[]","name":"ops","type":"tuple[]"},{"internalType":"address payable","name":"beneficiary","type":"address"}],"name":"handleOps","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint192","name":"key","type":"uint192"}],"name":"incrementNonce","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"bytes","name":"callData","type":"bytes"},{"components":[{"components":[{"internalType":"address","name":"sender","type":"address"},{"internalType":"uint256","name":"nonce","type":"uint256"},{"internalType":"uint256","name":"callGasLimit","type":"uint256"},{"internalType":"uint256","name":"verificationGasLimit","type":"uint256"},{"internalType":"uint256","name":"preVerificationGas","type":"uint256"},{"internalType":"address","name":"paymaster","type":"address"},{"internalType":"uint256","name":"maxFeePerGas","type":"uint256"},{"internalType":"uint256","name":"maxPriorityFeePerGas","type":"uint256"}],"internalType":"struct EntryPoint.MemoryUserOp","name":"mUserOp","type":"tuple"},{"internalType":"bytes32","name":"userOpHash","type":"bytes32"},{"internalType":"uint256","name":"prefund","type":"uint256"},{"internalType":"uint256","name":"contextOffset","type":"uint256"},{"internalType":"uint256","name":"preOpGas","type":"uint256"}],"internalType":"struct EntryPoint.UserOpInfo","name":"opInfo","type":"tuple"},{"internalType":"bytes","name":"context","type":"bytes"}],"name":"innerHandleOp","outputs":[{"internalType":"uint256","name":"actualGasCost","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"","type":"address"},{"internalType":"uint192","name":"","type":"uint192"}],"name":"nonceSequenceNumber","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"components":[{"internalType":"address","name":"sender","type":"address"},{"internalType":"uint256","name":"nonce","type":"uint256"},{"internalType":"bytes","name":"initCode","type":"bytes"},{"internalType":"bytes","name":"callData","type":"bytes"},{"internalType":"uint256","name":"callGasLimit","type":"uint256"},{"internalType":"uint256","name":"verificationGasLimit","type":"uint256"},{"internalType":"uint256","name":"preVerificationGas","type":"uint256"},{"internalType":"uint256","name":"maxFeePerGas","type":"uint256"},{"internalType":"uint256","name":"maxPriorityFeePerGas","type":"uint256"},{"internalType":"bytes","name":"paymasterAndData","type":"bytes"},{"internalType":"bytes","name":"signature","type":"bytes"}],"internalType":"struct UserOperation","name":"op","type":"tuple"},{"internalType":"address","name":"target","type":"address"},{"internalType":"bytes","name":"targetCallData","type":"bytes"}],"name":"simulateHandleOp","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"components":[{"internalType":"address","name":"sender","type":"address"},{"internalType":"uint256","name":"nonce","type":"uint256"},{"internalType":"bytes","name":"initCode","type":"bytes"},{"internalType":"bytes","name":"callData","type":"bytes"},{"internalType":"uint256","name":"callGasLimit","type":"uint256"},{"internalType":"uint256","name":"verificationGasLimit","type":"uint256"},{"internalType":"uint256","name":"preVerificationGas","type":"uint256"},{"internalType":"uint256","name":"maxFeePerGas","type":"uint256"},{"internalType":"uint256","name":"maxPriorityFeePerGas","type":"uint256"},{"internalType":"bytes","name":"paymasterAndData","type":"bytes"},{"internalType":"bytes","name":"signature","type":"bytes"}],"internalType":"struct UserOperation","name":"userOp","type":"tuple"}],"name":"simulateValidation","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"unlockStake","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address payable","name":"withdrawAddress","type":"address"}],"name":"withdrawStake","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address payable","name":"withdrawAddress","type":"address"},{"internalType":"uint256","name":"withdrawAmount","type":"uint256"}],"name":"withdrawTo","outputs":[],"stateMutability":"nonpayable","type":"function"},{"stateMutability":"payable","type":"receive"}]`
const HandleOpsSign = "1fad948c"

var wsRpcVar = "wss://localhost:8080"
var originVar = ""

var jsonABI abi.ABI

func init() {
	flag.StringVar(&wsRpcVar, "ws", wsRpcVar, "-ws "+wsRpcVar)
	flag.StringVar(&originVar, "origin", originVar, "-origin "+originVar)
	log.Root().SetHandler(
		log.LvlFilterHandler(log.LvlDebug, log.StdoutHandler),
	)

	reader := bytes.NewReader([]byte(ABIString))
	jsonABI, _ = abi.JSON(reader)
}

var pendingTransactions chan common.Hash

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
	for tx := range pendingTransactions {
		transaction, _, err := ethClient.TransactionByHash(context.Background(), tx)
		if err != nil {
			continue
		}

		data := transaction.Data()
		if len(data) > 4 {
			head := strings.ToLower(hex.EncodeToString(data[:4]))
			if head == HandleOpsSign {

				fmt.Println(head, tx, head == HandleOpsSign)
				//fmt.Println(hex.EncodeToString(transaction.Data()), isPending)

				// recover Method from signature and ABI
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
				fmt.Println(unpack)
			}

		}

	}
}
