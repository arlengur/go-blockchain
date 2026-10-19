package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

const difficulty = 1

type Block struct {
	Timestamp string // время создания
	Value     int    // что передаем в блоках
	Hash      string // ИД блока
	PrevHash  string // ссылка на родителя
	Nonce     string // Соль или число которое помогает формировать хэши
}

type Blockchain struct {
	sync.Mutex
	blocks []*Block
}

func (bc *Blockchain) AddBlock(newBlock *Block) {
	bc.Lock()
	defer bc.Unlock()
	bc.blocks = append(bc.blocks, newBlock)
}

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
	return Blockchain{sync.Mutex{}, []*Block{newGenesisBlock()}}
}

func generateBlockSalt(oldBlock *Block, value int) *Block {
	newBlock := &Block{}
	newBlock.Timestamp = time.Now().String()
	newBlock.Value = value
	newBlock.PrevHash = oldBlock.Hash

	i := 0
	for {
		newBlock.Nonce = fmt.Sprintf("%x", i)
		i++
		newHash := calcHash(newBlock)
		if !isHashValid(newHash, difficulty) {
			continue
		}
		newBlock.Hash = newHash
		break
	}

	return newBlock
}

func isHashValid(hash string, difficulty int) bool {
	prefix := strings.Repeat("0", difficulty)
	return strings.HasPrefix(hash, prefix)
}
