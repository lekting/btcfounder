package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"sync"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/joho/godotenv"
	"github.com/tyler-smith/go-bip39"
	"github.com/tyler-smith/go-bip39/wordlists"
	"github.com/valyala/tsvreader"
)

var oneWallet, threeWallet, bcWallet = readFileLineByLine("./blockchair_bitcoin_addresses_and_balance_LATEST.tsv")

func main() {
	err := godotenv.Load()
	if err != nil {
	  log.Fatal("Error loading .env file")
	}
	
	threads := flag.Int("t", 50, "threads")
	flag.Parse()
	var wg sync.WaitGroup

	// FOR TESTING
	// sort.Strings(oneWallet)
	// sort.Strings(threeWallet)
	// sort.Strings(bcWallet)

	bip39.SetWordList(wordlists.English)

	InitBot()

	log.Printf("Starting searching for wallets on %d threads...\n", *threads)

	wg.Add(1)
	for i := 0; i < *threads; i++ {
		wg.Add(1)
		go func() {

			defer wg.Done()

			for {
				mnemonic, err := NewMnemonic(12)

				if err != nil {
					continue
				}

				generator, err := FromMnemonic(mnemonic)

				if err != nil {
					continue
				}

				address, addressP2WPKH, addressP2WPKHInP2SH := generator.Generate()
				if oneWallet.TestString(address) {
					log.Println("BTC (address): ", address, mnemonic)
					SendBotMessage(fmt.Sprintf("Found (address) %s\n%s", address, mnemonic))
					continue
				}

				if bcWallet.TestString(addressP2WPKH) {
					log.Println("BTC (addressP2WPKH): ", addressP2WPKH, mnemonic)
					SendBotMessage(fmt.Sprintf("Found (addressP2WPKH) %s\n%s", addressP2WPKH, mnemonic))
					continue
				}

				if threeWallet.TestString(addressP2WPKHInP2SH) {
					log.Println("BTC (addressP2WPKHInP2SH): ", addressP2WPKHInP2SH, mnemonic)
					SendBotMessage(fmt.Sprintf("Found (addressP2WPKHInP2SH) %s\n%s", addressP2WPKHInP2SH, mnemonic))
					continue
				}

			}
		}()
	}

	wg.Wait()
}

// FOR TESTING
// func findWallet(wallets []string, wallet string) bool {
// 	_, found := slices.BinarySearch(wallets, wallet)

// 	return found
// }

func readFileLineByLine(filePath string) (*bloom.BloomFilter, *bloom.BloomFilter, *bloom.BloomFilter) {

	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	accuracy := 1 * math.Pow(10, -15)
	log.Printf("Creating bloom filters... (%e accuracy)\n", accuracy)
	oneWallet := bloom.NewWithEstimates(21092403, accuracy)
	bcWallet := bloom.NewWithEstimates(11532027, accuracy)
	threeWallet := bloom.NewWithEstimates(7900459, accuracy)

	bcWalletCount := 0
	threeWalletCount := 0
	oneWalletCount := 0

	r := tsvreader.New(file)
	for r.Next() {
		wallet := r.String()
		balance := r.Int64()

		if balance < 1000 {
			continue
		}
		
		if strings.HasPrefix(wallet, "bc1") {
			bcWallet.AddString(wallet)
			bcWalletCount++;
		} else if strings.HasPrefix(wallet, "3") {
			threeWallet.AddString(wallet)
			threeWalletCount++;
		} else if strings.HasPrefix(wallet, "1") {
			oneWallet.AddString(wallet)
			oneWalletCount++;
		}
	}
	if err := r.Error(); err != nil {
		fmt.Printf("unexpected error: %s", err)
	}

	fmt.Printf("Estimated bloom counts: %d %d %d\n", bcWalletCount, threeWalletCount, oneWalletCount)

	return oneWallet, threeWallet, bcWallet
}