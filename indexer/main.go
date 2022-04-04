package main

//32 bit && 64 bit

//https://github.com/ethereum/go-ethereum/issues/2602
import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"example2.com/ethprocess"
	"example2.com/models"

	"github.com/caarlos0/env"
	"github.com/ethereum/go-ethereum/core/types"
	_ "github.com/joho/godotenv/autoload" //support .env && autoload
)

var currentChainID *big.Int

const (
	NormalWorkerCnt            = 5
	UnstableReservedNum int64  = 20 //TODO
	ChainEndPoint       string = "https://data-seed-prebsc-1-s1.binance.org:8545/"
)

//增加log確認 保證資料完整性 與 錯誤校正
//set to release mode when production

func getDBConfig() *models.DBConfig {
	cfg := &models.DBConfig{}
	env.Parse(cfg)
	return cfg
}

func withContextFunc(ctx context.Context, f func()) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(c)

		select {
		case <-ctx.Done():
			fmt.Println("withContextFunc <-ctx.Done():")
		case <-c:
			cancel()
			f()
		}
	}()

	return ctx
}

func workerUnstable(ctx context.Context, wg *sync.WaitGroup) {
	const procInterval = time.Millisecond * 15
	const queryInterval = time.Millisecond * 15000
	check := int(queryInterval / procInterval)
	waitCnt := 0
	var unstableBlks []models.Block
	defer func() {
		wg.Done()
		fmt.Println("worker: workerUnstable leave")
	}()
	for {
		select {
		case <-ctx.Done():
			fmt.Println("worker: workerUnstable <-ctx.Done():")
			return
		default:
			time.Sleep(procInterval)
			size := len(unstableBlks)
			if size > 0 {
				block := unstableBlks[size-1]
				allDone, err := processUnstable(ctx, block)
				if err != nil {
					continue
				}
				if allDone {
					unstableBlks = unstableBlks[:0]
				} else {
					unstableBlks = unstableBlks[:size-1]
				}
			} else {
				if waitCnt == 0 {
					latest := ethprocess.GetLatestBlockNumber(ctx)
					if latest != -1 {
						unstableBlks = getUnstableBlock(latest - UnstableReservedNum)
					}
				}
				waitCnt++
				waitCnt = waitCnt % check
			}
		}
	}
}
func workerGetFailed(ctx context.Context, wg *sync.WaitGroup) {
	const procInterval = time.Millisecond * 15
	const queryInterval = time.Millisecond * 15000
	check := int(queryInterval / procInterval)
	waitCnt := 0

	var blks []models.NonGetBlock
	defer func() {
		wg.Done()
		fmt.Println("worker: workerGetFailed leave")
	}()
	for {
		select {
		case <-ctx.Done():
			fmt.Println("worker: workerGetFailed <-ctx.Done():")
			return
		default:
			time.Sleep(procInterval)
			size := len(blks)
			if size > 0 {
				block := blks[size-1]
				err := processBlock(ctx, block.Num, block.Num+UnstableReservedNum*2, false) //only process stable
				if err != nil {
					continue
				}

				blks = blks[:size-1]
				models.DeleteNonGetBlockNum(block.Num, block.ChainID)

			} else {
				if waitCnt == 0 {
					latest := ethprocess.GetLatestBlockNumber(ctx)
					if latest != -1 {
						blks = models.NonGetBlockNum(latest-UnstableReservedNum, currentChainID.Int64())
					}
				}
				waitCnt++
				waitCnt = waitCnt % check
			}
		}
	}
}

type ProcessBlock struct {
	num    int64
	latest int64
}

func worker(ctx context.Context, wk int, wg *sync.WaitGroup, ch <-chan ProcessBlock) {
	defer func() {
		wg.Done()
		fmt.Println("worker: ", wk, " leave")
	}()
	for {
		select {
		case <-ctx.Done():
			fmt.Println("worker: ", wk, " <-ctx.Done():")
			return
		default:
			select {
			case info := <-ch:
				processBlock(ctx, info.num, info.latest, true)
			default:

			}
		}
	}
}

func workerManager(ctx context.Context, wg *sync.WaitGroup) {
	const procInterval = time.Millisecond * 30
	const checkInterval = time.Millisecond * 5000
	wgWorker := &sync.WaitGroup{}
	workerCnt := NormalWorkerCnt + 2 //normal + unstable + workerGetFailed
	wgWorker.Add(workerCnt)
	ch := make(chan ProcessBlock, NormalWorkerCnt)

	latestNumber := ethprocess.GetLatestBlockNumber(ctx)
	start := latestNumber - UnstableReservedNum
	if start < 0 {
		start = 0
	}
	numInDB := models.GetLatestBlockNum(currentChainID.Int64())
	numInDB++
	if start < numInDB {
		start = numInDB
	}
	go workerUnstable(ctx, wgWorker)
	go workerGetFailed(ctx, wgWorker)
	for i := 0; i < NormalWorkerCnt; i++ {
		go worker(ctx, i, wgWorker, ch)
	}

	defer func() {
		wg.Done()
		close(ch)
	}()
	for {

		select {
		case <-ctx.Done():
			fmt.Println("workerManager <-ctx.Done():  before wait")
			wgWorker.Wait()
			fmt.Println("workerManager <-ctx.Done():  after wait")
			return
		default:
			if start >= latestNumber {
				latestNumber = ethprocess.GetLatestBlockNumber(ctx)
				time.Sleep(checkInterval)
			} else {
				tmp := ProcessBlock{
					num:    start,
					latest: latestNumber,
				}
				select {
				case ch <- tmp: //write
					start++
				default:
				}

			}
			time.Sleep(procInterval)
		}
	}

}

func main() {
	finished := make(chan bool)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	ctx := withContextFunc(context.Background(), func() {
		log.Println("cancel application")
		wg.Wait()
		close(finished)
	})

	chainID, err := ethprocess.Connect(ctx, ChainEndPoint)
	if err != nil {
		log.Fatal(err)
	}
	currentChainID = chainID

	cfg := getDBConfig()
	err = models.Init(cfg)

	if err != nil {
		log.Fatal(err)
	}
	models.Migration() //remove for production

	go workerManager(ctx, wg)

	<-finished
	log.Println("Finished && close indexer.")
}

func getUnstableBlock(beforeNumber int64) []models.Block {

	data, err := models.GetUnstableBlock(currentChainID.Int64(), beforeNumber)
	if err != nil {
		return []models.Block{}
	}
	return data

}

func processUnstable(ctx context.Context, blk models.Block) (
	bool, error) { //all stable, error

	chainID := currentChainID.Int64()

	cur, err := ethprocess.BlockByNumber(ctx, blk.Num)
	if err != nil {
		return false, err
	}

	if cur.Hash().String() == blk.Hash {
		//parent will not change if hash dont' change
		//update all block to stable before blk.Num
		models.UpdateBlockStable(blk.Num, chainID)
		return true, nil
	}

	//update single block
	desBlock, err := wrapEthBlockAndTransactinos(ctx, cur, chainID, true)
	if err != nil {
		return false, err
	}

	desBlock.ID = blk.ID
	for i := 0; i < len(desBlock.Transactions); i++ {
		desBlock.Transactions[i].BlockID = blk.ID
	}
	models.ReplaceBlock(desBlock)
	return false, nil
}

func wrapEthBlockAndTransactinos(ctx context.Context, block *types.Block,
	chainID int64, stable bool) (*models.Block, error) {

	desBlock := models.Block{
		Num:        block.Number().Int64(),
		Hash:       block.Hash().String(),
		Time:       block.Header().Time,
		ParentHash: block.ParentHash().String(),
		Stable:     stable,
		ChainID:    chainID,
	}

	transactions := block.Transactions()
	cnt := transactions.Len()
	desBlock.Transactions = make([]models.Transaction, cnt)
	for i := 0; i < cnt; i++ {
		msg, err := transactions[i].AsMessage(types.NewEIP155Signer(currentChainID), nil)
		if err != nil {
			return nil, err
		}
		receipt, err := ethprocess.TransactionReceipt(ctx, transactions[i].Hash())
		if err != nil {
			return nil, err
		}
		logs, err := json.Marshal(receipt.Logs) //TODO: decide how to process error
		if err != nil {
			return nil, err
		}
		to := msg.To()
		toString := ""
		if to != nil {
			toString = to.String()
		}

		desBlock.Transactions[i] = models.Transaction{
			Hash:  transactions[i].Hash().String(),
			From:  msg.From().String(),
			To:    toString,
			Nonce: msg.Nonce(),
			Data:  hex.EncodeToString(msg.Data()),
			Logs:  string(logs),
		}
		if msg.Value() != nil {
			value := msg.Value().Int64()
			desBlock.Transactions[i].Value = &value
		}
	}

	return &desBlock, nil
}

func processBlock(ctx context.Context, num, latestInChain int64,
	saveNonGet bool) (theError error) {
	stable := true
	if num > latestInChain-UnstableReservedNum {
		stable = false
	}
	chainID := currentChainID.Int64()
	defer func() {
		if theError == nil {
			return
		}
		if !saveNonGet {
			return
		}
		nonGet := &models.NonGetBlock{
			Num:     num,
			ChainID: chainID,
			Error:   theError.Error(),
		}
		//紀錄未取得區塊
		//可以再用worker去取得
		models.CreateNonGetBlock(nonGet)
		//如果未寫入資料庫,則寫入輔助儲存裝置或log
		//以利後續更正處理
	}()
	block, err := ethprocess.BlockByNumber(ctx, num)
	if block == nil {
		theError = err
		return
	}

	desBlock, err := wrapEthBlockAndTransactinos(ctx, block, chainID, stable)
	if err != nil {
		theError = err
		return
	}

	models.CreateBlock(desBlock)
	//如果有錯誤或未寫入資料庫,則寫入輔助儲存裝置或log
	//以利後續更正處理
	return
}
