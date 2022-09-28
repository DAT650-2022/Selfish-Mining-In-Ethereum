package main

import (
	"bufio"
	"crypto/ecdsa"
	"fmt"
	"os"
	"strconv"
)

type cli struct {
	scanner           *bufio.Scanner
	blockchain        *Blockchain
	UTXOset           UTXOSet
	transactionBuffer []*Transaction
	addressBook       map[string]keyPair //Map containing all privateKeys
}

type keyPair struct {
	privateKey    ecdsa.PrivateKey
	publicKeyByte []byte
}

func main() {
	fmt.Println("Welcome to Erik's Blockchain!")
	cli := NewCLI()
	scanner := bufio.NewScanner(os.Stdin)
	for {
		printIntro()
		scanner.Scan()
		input := scanner.Text()
		switch input {
		case "1":
			cli.createBlockchain()
		case "2":
			if cli.blockchain == nil {
				fmt.Println("You need to create a blockchain before adding transactions")
			} else {
				cli.addTransaction()
			}
		case "3":
			if cli.blockchain == nil {
				fmt.Println("You need to create a blockchain before mining a block")
			} else {
				cli.mineBlock()
			}
		case "4":
			if cli.blockchain == nil {
				fmt.Println("You need to create a blockchain before printing chain")
			} else {
				fmt.Println(cli.blockchain.String())
			}
		case "5":
			if cli.blockchain == nil {
				fmt.Println("You need to create a blockchain before printing block")
			} else {
				cli.printBlock()
			}
		case "6":
			if cli.blockchain == nil {
				fmt.Println("You need to create a blockchain before printing transaction")
			} else {
				cli.printTx()
			}
		case "7":
			cli.createAddress()
		case "8":
			cli.getBalance()
		case "9":
			return
		default:
			fmt.Println("This command doesnt exist")
		}
		fmt.Println("_____________________________________________")
	}
}

func NewCLI() *cli {
	scanner := bufio.NewScanner(os.Stdin)
	addressBook := make(map[string]keyPair)
	return &cli{scanner: scanner, blockchain: nil, UTXOset: nil, transactionBuffer: nil, addressBook: addressBook}
}

func (c *cli) createBlockchain() {
	fmt.Print("Enter your address: ")
	c.scanner.Scan()
	input := c.scanner.Text()
	if _, ok := c.addressBook[input]; !ok {
		fmt.Println("Address doesnt exist. Create new address and try again")
		return
	}
	blockchain, _ := NewBlockchain(GetStringAddress(c.addressBook[input].publicKeyByte))
	c.blockchain = blockchain
	c.UTXOset = c.blockchain.FindUTXOSet()
	fmt.Println("New blockchain created")
}

func (c *cli) addTransaction() {
	fmt.Print("Enter your address: ")
	c.scanner.Scan()
	address := c.scanner.Text()
	if _, ok := c.addressBook[address]; !ok {
		fmt.Println("Address name doesnt exist. Create new address and try again")
		return
	}
	fmt.Print("Recipient address: ")
	c.scanner.Scan()
	receiver := c.scanner.Text()
	if _, ok := c.addressBook[receiver]; !ok {
		fmt.Println("Receiver address name doesnt exist. Create new address and try again")
		return
	}
	fmt.Print("Amount: ")
	c.scanner.Scan()
	amount, _ := strconv.Atoi(c.scanner.Text())
	transaction, err := NewUTXOTransaction(c.addressBook[address].publicKeyByte, GetStringAddress(c.addressBook[receiver].publicKeyByte), amount, c.UTXOset)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = c.blockchain.SignTransaction(transaction, c.addressBook[address].privateKey)
	if err != nil {
		fmt.Println(err)
		return
	}
	c.transactionBuffer = append(c.transactionBuffer, transaction)
	fmt.Printf("Transaction: \n %s \n added to buffer \n", transaction)
}

func (c *cli) createAddress() {
	fmt.Print("Enter address name: ")
	c.scanner.Scan()
	name := c.scanner.Text()
	privateKey, publicKeyBytes := newKeyPair()
	c.addressBook[name] = keyPair{privateKey: privateKey, publicKeyByte: publicKeyBytes}
	fmt.Printf("Address created with Name: %s Address: %s \n", name, string(GetAddress(publicKeyBytes)))
}

func (c *cli) mineBlock() {
	fmt.Print("Enter your address: ")
	c.scanner.Scan()
	address := c.scanner.Text()
	coinbase, _ := NewCoinbaseTX(GetStringAddress(c.addressBook[address].publicKeyByte), "")
	c.transactionBuffer = append([]*Transaction{coinbase}, c.transactionBuffer...)
	block, err := c.blockchain.MineBlock(c.transactionBuffer)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Block %x added to blockchain \n", block.Hash)
		c.UTXOset.Update(c.transactionBuffer)
		c.transactionBuffer = nil
	}
}

func (c *cli) printBlock() {
	fmt.Print("Block hash: ")
	c.scanner.Scan()
	input := c.scanner.Text()
	block, err := c.blockchain.GetBlock(Hex2Bytes(input))
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(block)
	}
}

func (c *cli) printTx() {
	fmt.Print("Transaction ID: ")
	c.scanner.Scan()
	input := c.scanner.Text()
	tx, err := c.blockchain.FindTransaction(Hex2Bytes(input))
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Transaction: \n %s \n", tx)
	}
}

func (c *cli) getBalance() {
	balance := make(map[string]int)
	for name, key := range c.addressBook {
		pubKeyHash := GetPubKeyHashFromAddress(GetStringAddress(key.publicKeyByte))
		output := c.UTXOset.FindUTXO(pubKeyHash)
		for _, out := range output {
			if _, ok := balance[name]; ok {
				balance[name] += out.Value
			} else {
				balance[name] = out.Value
			}
		}
	}
	fmt.Println("Balance:")
	for name, bal := range balance {
		fmt.Printf("Name: %s, Address: %s, Amount: %d \n", name, GetAddress(c.addressBook[name].publicKeyByte), bal)
	}
}

func printIntro() {
	fmt.Println("1. Create Blockchain")
	fmt.Println("2. Add Transaction")
	fmt.Println("3. Mine Block")
	fmt.Println("4. Print Chain")
	fmt.Println("5. Print Block")
	fmt.Println("6. Print Transaction")
	fmt.Println("7. Create Address")
	fmt.Println("8. Get Balance")
	fmt.Println("9. Exit")
	fmt.Print("Select command: ")
}
