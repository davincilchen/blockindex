package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"example.com/models"
	"github.com/gin-gonic/gin"
)

//https://pkg.go.dev/github.com/bitly/go-simplejson

type SimpleBlock struct {
	Num        int64  `json:"block_num"`
	Hash       string `json:"block_hash"`
	Time       uint64 `json:"block_time"`
	ParentHash string `json:"parent_hash"`
	Stable     bool   `json:"stable"`
	ChainID    int64  `json:"chain_id"`
}

type SimpleTransaction struct {
	Hash string
}

type GetBlocksOutput struct {
	BlockCnt int           `json:"block_count"` //just demo for quickly check by human
	Blocks   []SimpleBlock `json:"blocks"`
}

type GetBlockOutput struct {
	SimpleBlock
	Transactions []string `json:"transactions"`
}

type GetTransactionOutput struct {
	Hash  string        `json:"hash"`
	From  string        `json:"from"`
	To    string        `json:"to"`
	Nonce uint64        `json:"nonce"`
	Data  string        `json:"data"`
	Value *int64        `json:"value"`
	Logs  []interface{} `json:"logs"`
}

func parseChainID(val string) *int64 {
	if val == "" {
		ret := int64(-1) //default
		return &ret
	}

	chainID, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return nil
	}
	return &chainID
}

func ToAPIBlockWithTransations(parm *models.Block) GetBlockOutput {
	blk := ToAPISimpleBlock(parm)

	ret := GetBlockOutput{
		SimpleBlock: blk,
	}
	ret.Transactions = make([]string, len(parm.Transactions))
	for i, v := range parm.Transactions {
		ret.Transactions[i] = v.Hash
	}

	return ret
}

func ToAPISimpleBlock(parm *models.Block) SimpleBlock {
	return SimpleBlock{
		Num:        parm.Num,
		Hash:       parm.Hash,
		Time:       parm.Time,
		ParentHash: parm.ParentHash,
		Stable:     parm.Stable,
		ChainID:    parm.ChainID,
	}
}
func GetBlocks(c *gin.Context) {

	tmp := c.Query("limit")

	if tmp == "" {
		tmp = "10"
	}

	limit, err := strconv.ParseInt(tmp, 10, 64)
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}
	if limit <= 0 {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	} else if limit > 10000 {
		limit = 10000
	}

	tmp = c.Query("chain_id")
	chainID := parseChainID(tmp)
	if chainID == nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	blocks, err := models.GetChainBlocks(limit, *chainID)
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	cnt := len(blocks)
	res := GetBlocksOutput{
		BlockCnt: cnt,
		Blocks:   make([]SimpleBlock, len(blocks)),
	}

	for i, v := range blocks {
		res.Blocks[i] = ToAPISimpleBlock(&v)
	}

	c.JSON(http.StatusOK, res)
}

func GetBlock(c *gin.Context) {
	numText := c.Param("num")
	num, err := strconv.ParseInt(numText, 10, 64)
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	tmp := c.Query("chain_id")
	chainID := parseChainID(tmp)
	if chainID == nil {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	block, err := models.GetChainBlock(num, *chainID)
	if err != nil {
		if models.ErrRecordNotFound(err) {
			c.Writer.WriteHeader(http.StatusNotFound)
			return
		}
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	res := ToAPIBlockWithTransations(block)
	c.JSON(http.StatusOK, res)
}

func GetTransaction(c *gin.Context) {
	txHash := c.Param("txHash")

	transaction, err := models.GetTransaction(txHash)
	if err != nil {
		if models.ErrRecordNotFound(err) {
			c.Writer.WriteHeader(http.StatusNotFound)
			return
		}
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	res := GetTransactionOutput{
		Hash:  transaction.Hash,
		From:  transaction.From,
		To:    transaction.To,
		Nonce: transaction.Nonce,
		Data:  transaction.Data,
		Value: transaction.Value,
	}

	if transaction.Logs != "" {
		err = json.Unmarshal([]byte(transaction.Logs), &res.Logs)
		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	c.JSON(http.StatusOK, res)
}
