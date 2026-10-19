package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"
)

const (
	difficulty = 1

	genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
)

type Block struct {
	Timestamp    string // time of creation
	Transactions []*Transaction
	Hash         string // Block Id
	PrevHash     string // ссылка на родителя
	Nonce        string // Соль или число которое помогает формировать хэши
}

type Blockchain struct {
	sync.Mutex
	blocks []*Block
}

func NewBlockchain() Blockchain {
	return Blockchain{sync.Mutex{}, []*Block{newGenesisBlock()}}
}

func (bc *Blockchain) AddBlock(newBlock *Block) {
	bc.Lock()
	defer bc.Unlock()
	bc.blocks = append(bc.blocks, newBlock)
}

func newGenesisBlock() *Block {
	genesisBlock := &Block{}

	return &Block{
		Timestamp:    time.Now().String(),
		Transactions: []*Transaction{NewCoinbaseTX("Arlen", genesisCoinbaseData)},
		Hash:         calcHash(genesisBlock),
	}
}

func calcHash(block *Block) string {
	record := block.Timestamp + block.PrevHash + block.Nonce
	h := sha256.New()
	h.Write(append([]byte(record), block.HashTransactions()...))
	hashed := h.Sum(nil) // get hash as bytes
	return hex.EncodeToString(hashed)
}

func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, []byte(tx.ID))
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
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

func generateBlock(oldBlock *Block, m SendMessage) (*Block, error) {
	newBlock := new(Block)

	newTranaction, err := NewUTXOTransaction(m.From, m.To, m.Value, &bc)
	if err != nil {
		return nil, err
	}
	newBlock.Timestamp = time.Now().String()
	newBlock.Transactions = []*Transaction{newTranaction}
	newBlock.PrevHash = oldBlock.Hash

	i := 0
	for {
		newBlock.Nonce = fmt.Sprintf("%x", i)
		i++
		newHash := calcHash(newBlock)
		if !isHashValid(newHash, difficulty) {
			continue
		}
		fmt.Println(newHash, " work done!")
		newBlock.Hash = newHash
		break

	}
	return newBlock, nil
}

func isHashValid(hash string, difficulty int) bool {
	prefix := strings.Repeat("0", difficulty)
	return strings.HasPrefix(hash, prefix)
}

// FindUnspentTransactions returns a list of transactions containing unspent outputs
func (bc *Blockchain) FindUnspentTransactions(address string) []*Transaction {
	var unspentTXs []*Transaction
	spentTXOs := make(map[string][]int)

	for _, b := range slices.Backward(bc.blocks) {
		if len(b.Transactions) == 0 {
			return unspentTXs
		}

		for _, tx := range b.Transactions {
			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
						spentTXOs[in.Txid] = append(spentTXOs[in.Txid], in.Vout)
					}
				}
			}
		Outputs:
			for outIdx, out := range tx.Vout {
				for _, spentOut := range spentTXOs[tx.ID] {
					if spentOut == outIdx {
						continue Outputs
					}
				}

				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, tx)
				}
			}
		}
	}

	return unspentTXs
}

// FindSpendableOutputs finds and returns unspent outputs to reference in inputs
func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (
	int, map[string][]int) {

	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(address)
	accumulated := 0

	for _, tx := range unspentTXs {
		for idx, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[tx.ID] = append(unspentOutputs[tx.ID], idx)

				if accumulated >= amount {
					return accumulated, unspentOutputs
				}
			}
		}
	}
	return accumulated, unspentOutputs
}

// FindUTXO finds and returns all unspent transaction outputs
func (bc *Blockchain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := bc.FindUnspentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}
