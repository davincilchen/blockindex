package models

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gomodule/redigo/redis"
	"gorm.io/gorm"
)

var db *gorm.DB

type BlockBase struct {
	gorm.Model
	Num  int64
	Hash string
}

type Block struct {
	BlockBase
	Time         uint64
	ParentHash   string
	Stable       bool
	ChainID      int64
	Transactions []Transaction
}

type Transaction struct {
	gorm.Model
	BlockID uint
	Block   *Block
	Hash    string
	From    string
	To      string
	Nonce   uint64
	Data    string
	Value   *int64
	Logs    string
}

const redisExpireSec = int64(60)

func RedisKeyForBlock(blockNum, chainID int64) string {
	return fmt.Sprintf("block_%d", blockNum) //TODO improve chainID
}

func GetBlockInRedis(blockNum, chainID int64) (*Block, error) {
	key := RedisKeyForBlock(blockNum, chainID)
	data, err := RedisGet(key)
	if err != nil {
		return nil, err
	}

	ss, err := redis.String(data, err)
	if err != nil {
		return nil, err
	}

	out := Block{}
	err = json.Unmarshal([]byte(ss), &out)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

func SaveBlockToRedis(block *Block) error {
	// if block == nil {
	// 	return nil
	// }
	key := RedisKeyForBlock(block.Num, block.ChainID)
	bytes, err := json.Marshal(block)
	if err != nil {
		return err
	}

	_, err = RedisSetWithExpire(redisExpireSec, key, bytes)

	return err
}
func Init(_db *gorm.DB) {
	db = _db
}

func ErrRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func GetChainBlocks(limit, chainID int64) ([]Block, error) {
	if chainID < 0 {
		return getBlocks(limit)
	}
	return getChainBlocks(limit, chainID)
}

func getBlocks(limit int64) ([]Block, error) {
	var blocks []Block
	dbc := db.Limit(int(limit)).
		Order("num desc").Find(&blocks)

	return blocks, dbc.Error
}

func getChainBlocks(limit, chainID int64) ([]Block, error) {

	var blocks []Block
	dbc := db.Limit(int(limit)).Where("chain_id = ?", chainID).
		Order("num desc").Find(&blocks)

	return blocks, dbc.Error
}

// .. //

func GetChainBlock(blockNum, chainID int64) (*Block, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered in GetChainBlock for : %v\n", r)
		}
	}()
	block, err := GetBlockInRedis(blockNum, chainID)
	if err == nil {
		return block, nil
	}
	var blk *Block
	defer func() {
		SaveBlockToRedis(blk)
	}()

	if chainID < 0 {
		block, err2 := getBlock(blockNum)
		if err2 != nil {
			return nil, err2
		}
		blk = block
		return blk, nil
	}
	block, err = getChainBlock(blockNum, chainID)
	if err != nil {
		return nil, err
	}
	blk = block
	return blk, nil
}

func getBlock(blockNum int64) (*Block, error) {
	var block Block
	dbc := db.Preload("Transactions").First(&block, "num = ?", blockNum)
	if dbc.Error != nil {
		return nil, dbc.Error
	}
	return &block, nil
}

func getChainBlock(blockNum, chainID int64) (*Block, error) {
	var block Block
	dbc := db.Preload("Transactions").
		First(&block, "num = ? and chain_id = ?", blockNum, chainID)

	if dbc.Error != nil {
		return nil, dbc.Error
	}
	return &block, nil
}

// .. //
func GetTransaction(hash string) (*Transaction, error) {
	var transaction Transaction
	dbc := db.First(&transaction, "hash = ?", hash)
	if dbc.Error != nil {
		return nil, dbc.Error
	}
	return &transaction, nil
}
