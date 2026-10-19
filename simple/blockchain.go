package main

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"
)

type Block struct {
	Timestamp string // время создания
	Value     int    // что передаем в блоках
	Hash      string // ИД блока
	PrevHash  string // ссылка на родителя
	Nonce     string // Соль или число которое помогает формировать хэши
}

type Blockchain []*Block

func calcHash(block *Block) string {
	record := block.Timestamp + strconv.Itoa(block.Value) + block.PrevHash + block.Nonce
	h := sha256.New()
	h.Write([]byte(record)) // преобразуем строку в байты и передаем в функцию
	hashed := h.Sum(nil)    // получаем хэш в виде байтов
	return hex.EncodeToString(hashed)
}

func isBlockValid(newBlock *Block, oldBlock *Block) bool {
	if newBlock.PrevHash != oldBlock.Hash {
		return false
	}

	if calcHash(newBlock) != newBlock.Hash {
		return false
	}
	return true
}

func newGenesisBlock() *Block {
	genesisBlock := &Block{}

	return &Block{
		Timestamp: time.Now().String(),
		Value:     0,
		Hash:      calcHash(genesisBlock),
	}
}

func NewBlockchain() Blockchain {
	return Blockchain{newGenesisBlock()}
}

func generateBlock(oldBlock *Block, value int) *Block {
	newBlock := &Block{}
	newBlock.Timestamp = time.Now().String() // текущее время в формате строки
	newBlock.Value = value
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Nonce = "0"
	newBlock.Hash = calcHash(newBlock)
	return newBlock
}
