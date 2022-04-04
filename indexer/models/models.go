package models

import (
	"fmt"

	"gorm.io/gorm"
)

//hash example
//0xe5c7c4e8024b237d44b2e570e9858503913784178a20473ea7147d3f6b679602

type NonGetBlock struct {
	Num     int64 `gorm:"index;not null"`
	ChainID int64 `gorm:"not null"`
	Error   string
}

type BlockBase struct {
	gorm.Model
	Num  int64  `gorm:"index;not null"`
	Hash string `gorm:"type:char(66) ;not null"`
}

type Block struct {
	gorm.Model
	Num          int64  `gorm:"index;not null"`
	Hash         string `gorm:"type:char(66) ;not null"`
	Time         uint64 `gorm:"not null"`
	ParentHash   string `gorm:"type:char(66) ;not null"`
	Stable       bool   `gorm:"type:boolean ;not null"`
	ChainID      int64  `gorm:"not null"`
	Transactions []Transaction
}

type Transaction struct {
	gorm.Model
	BlockID uint `gorm:"index;not null"`
	Block   *Block
	Hash    string `gorm:"index;type:char(66);not null"`
	From    string `gorm:"type:char(66);not null"`
	To      string `gorm:"type:char(66)"`
	Nonce   uint64
	Data    string
	Value   *int64
	Logs    string
}

var db *gorm.DB

func Init(cfg *DBConfig) error {
	d, err := GormOpen(cfg)
	if err != nil {
		return err
	}
	db = d
	return nil
}

func Migration() { //remove for production
	fmt.Println("Run Migration --> Start")

	db.AutoMigrate(&Block{})
	db.AutoMigrate(&Transaction{})
	db.AutoMigrate(&NonGetBlock{})

	fmt.Println("Run Migration --> Done")

}

func GetLatestBlockNum(chainID int64) int64 {
	var blocks []Block
	db.Limit(1).Select("num").Where("chain_id = ?", chainID).
		Order("num desc").Find(&blocks)

	if len(blocks) == 0 {
		return 0
	}
	return blocks[0].Num
}

func DeleteNonGetBlockNum(num, chainID int64) {
	var block NonGetBlock
	db.Where("num = ? and chain_id = ?", num, chainID).Delete(&block)
}

func NonGetBlockNum(num, chainID int64) []NonGetBlock {
	var blocks []NonGetBlock
	db.Where("chain_id = ? and num < ?", chainID, num).
		Order("num ").Find(&blocks)
	return blocks
}

func GetUnstableBlock(chainID int64, beforeNumber int64) ([]Block, error) {
	dataOut := []Block{}

	dbc := db.Where("chain_id = ? and stable = false and num <= ?", chainID, beforeNumber).
		Order("num").
		Find(&dataOut)

	if dbc.Error != nil {
		return nil, dbc.Error
	}

	return dataOut, nil
}

func CreateNonGetBlock(block *NonGetBlock) error {
	tx := db.Create(block)
	return tx.Error
}

//TODO: write to log or some storage and
// assign some application/worker to update losed data to DB when write to DB failed
func CreateBlock(block *Block) error {


	dbc := db.Where("num = ? and chain_id = ?", block.Num, block.ChainID).
		FirstOrCreate(block)


	if dbc.Error != nil {
		return dbc.Error
	}
	return nil
}

func ReplaceBlock(block *Block) error {

	tx := db.Begin()
	defer tx.Rollback()

	dbc := tx.Where("block_id = ?", block.ID).Delete(&Transaction{})
	if dbc.Error != nil {
		return dbc.Error
	}

	dbc = tx.Create(&block.Transactions)
	if dbc.Error != nil {
		return dbc.Error
	}

	dbc = tx.Model(&block).Update("stable", "true")
	if dbc.Error != nil {

		return dbc.Error
	}

	_ = tx.Commit()
	return nil
}

func UpdateBlockStable(beforeNum int64, chainID int64) error {
	dbc := db.Model(Block{}).Where("chain_id = ? and num <= ?",
		chainID, beforeNum).Update("stable", "true")
	return dbc.Error

}

//https://ethereum.stackexchange.com/questions/268/ethereum-block-architecture
//https://medium.com/coinmonks/ethereum-under-the-hood-part-7-blocks-c8a5f57cc356
//https://github.com/ethereum/wiki/wiki/%5B%E4%B8%AD%E6%96%87%5D-%E4%BB%A5%E5%A4%AA%E5%9D%8A%E7%99%BD%E7%9A%AE%E4%B9%A6
//https://medium.com/@pswoo/%E4%B9%99%E5%A4%AA%E5%9D%8A%E9%BB%83%E7%9A%AE%E6%9B%B8%E4%B9%8B%E9%96%B1%E8%AE%80%E7%AD%86%E8%A8%98-%E3%84%A7-2a202aff4400
//https://ethereum.github.io/yellowpaper/paper.pdf
//https://github.com/wanshan1024/ethereum_yellowpaper
//https://github.com/wanshan1024/ethereum_yellowpaper/blob/master/ethereum_yellow_paper_cn.pdf
//https://etherscan.io/
