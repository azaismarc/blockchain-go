package main

import (
	"bytes"
	"crypto/sha512"
	"fmt"
	"os"
	"strconv"
	"sync"
)

type Block struct {
	nonce    int
	prevHash []byte
	data     string
}

func (b *Block) hash() []byte {
	h := sha512.New()
	h.Write([]byte(strconv.Itoa(b.nonce)))
	h.Write(b.prevHash)
	h.Write([]byte(b.data))
	return h.Sum(nil)
}

func (b *Block) String() string {
	return fmt.Sprintf("Nonce: %d\nData: %s\nHash: %x\nPrev: %x\n", b.nonce, b.data, b.hash(), b.prevHash)
}

type Blockchain struct {
	challenge int
	blocks    []*Block
	sync.RWMutex
}

func mine(challenge int, prevHash []byte, data string) *Block {
	nonce := 0
	for {
		b := &Block{nonce, prevHash, data}
		h := b.hash()
		for i := 0; i < challenge; i++ {
			if h[i] != 0 {
				break
			}
			if i == challenge-1 {
				return b
			}
		}
		nonce++
	}
}

func (bc *Blockchain) addBlock(data string) bool {
	defer bc.Unlock()
	prevHash := bc.blocks[len(bc.blocks)-1].hash()
	mine := mine(bc.challenge, prevHash, data)
	bc.Lock()
	val := bc.validate(mine.prevHash)
	if !val {
		return false
	}
	bc.blocks = append(bc.blocks, mine)
	return true
}

func (bc *Blockchain) String() string {
	s := ""
	for _, b := range bc.blocks {
		s += b.String() + "\n"
	}
	return s
}

func (bc *Blockchain) validate(prevHash []byte) bool {
	return bytes.Equal(prevHash, bc.blocks[len(bc.blocks)-1].hash())
}

func (bc *Blockchain) validateAll() bool {
	for i := 1; i < len(bc.blocks); i++ {
		if !bytes.Equal(bc.blocks[i].prevHash, bc.blocks[i-1].hash()) {
			return false
		}
	}
	return true
}

func genesis(challenge int) *Blockchain {
	return &Blockchain{challenge, []*Block{mine(challenge, []byte{}, "Genesis")}, sync.RWMutex{}}
}

func computer(id int, bc *Blockchain, data string, done chan<- bool) {
	data = fmt.Sprintf("Thread %d %s", id, data)
	for {
		if bc.addBlock(data) {
			done <- true
		}
	}
}

func main() {
	bc := genesis(2)

	nBlockArg := os.Args[2]
	nBlock, _ := strconv.Atoi(nBlockArg)

	done := make(chan bool, nBlock)
	arg := os.Args[1]
	threads, _ := strconv.Atoi(arg)

	for i := 0; i < threads; i++ {
		go computer(i, bc, "", done)
	}

	for i := 0; i < nBlock; i++ {
		<-done
	}

	fmt.Println(bc, "\n", len(bc.blocks))

	if bc.validateAll() {
		fmt.Println("Validated")
	} else {
		fmt.Println("Invalid")
	}
}
