package ethprocess

//32 bit && 64 bit
//https://github.com/ethereum/go-ethereum/issues/2602

import (
	"context"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	Retry     = 1
	RetryWait = 1 //second
)

var ethClient *ethclient.Client

func Connect(ctx context.Context, rawurl string) (*big.Int, error) {
	client, err := ethclient.Dial(rawurl)
	if err != nil {
		return nil, err
	}
	ethClient = client
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, err
	}
	return chainID, nil

}

func GetLatestBlockNumber(ctx context.Context) int64 {

	latestNumber := int64(-1)
	try := Retry
	for {
		header, err := ethClient.HeaderByNumber(ctx, nil)
		if err != nil {
			try--
			if try <= 0 {
				break
			}
			time.Sleep(RetryWait * time.Second)
			continue
		}
		num := header.Number.String()
		latestNumber, _ = strconv.ParseInt(num, 10, 64)
		break
	}

	return latestNumber
}

func BlockByNumber(ctx context.Context, index int64) (*types.Block, error) {
	var theErr error
	var block *types.Block
	try := Retry
	for {
		tmp, err := ethClient.BlockByNumber(ctx, big.NewInt(index))
		if err != nil {
			theErr = err
			try--
			if try <= 0 {
				break
			}
			time.Sleep(RetryWait * time.Second)
			continue
		}
		block = tmp
		break
	}

	return block, theErr
}

func TransactionReceipt(ctx context.Context, txHash common.Hash) (
	*types.Receipt, error) {
	var theErr error
	var receipt *types.Receipt
	try := Retry
	for {
		tmp, err := ethClient.TransactionReceipt(ctx, txHash)
		if err != nil {
			theErr = err
			try--
			if try <= 0 {
				break
			}
			time.Sleep(RetryWait * time.Second)
			continue
		}
		receipt = tmp
		break
	}

	return receipt, theErr
}
