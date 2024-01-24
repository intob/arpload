package main

import (
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"

	"github.com/everFinance/goar"
	"github.com/everFinance/goar/types"
)

type Options struct {
	Gateway, WalletFile, Filename, Title, Description, Author string
}

func main() {
	if len(os.Args) == 1 {
		log.Fatal("missing file argument")
	}
	opt := &Options{
		Gateway:    "https://arweave.net",
		WalletFile: os.Getenv("AR_WALLET"),
	}
	if opt.WalletFile == "" {
		log.Fatal("AR_WALLET must be defined in your env")
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	w, err := goar.NewWalletFromPath(opt.WalletFile, opt.Gateway)
	if err != nil {
		log.Fatal(err)
	}

	reward, err := w.Client.GetTransactionPrice(len(data), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("this will cost %sAR\n", convertWinstonToAr(big.NewInt(reward)))
	if !confirm() {
		log.Fatal("user aborted")
	}

	fmt.Println("uploading...")
	err = sendDataStream(w, opt)
	if err != nil {
		log.Fatal(err)
	}
}

func sendDataStream(w *goar.Wallet, opt *Options) error {
	contentType, err := detectContentType(opt.Filename)
	if err != nil {
		return fmt.Errorf("unable to detect content type: %w", err)
	}
	data, err := os.Open(opt.Filename)
	if err != nil {
		return err
	}
	defer data.Close()
	tags := []types.Tag{
		{Name: "Content-Type", Value: contentType},
		{Name: "Title", Value: opt.Title},
		{Name: "Description", Value: opt.Description},
		{Name: "Author", Value: opt.Author},
	}
	tx, err := w.SendDataStreamSpeedUp(data, tags, 10)
	fmt.Println(tx, err)

	return err
}

func detectContentType(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()
	// read first 512 bytes
	buf := make([]byte, 512)
	_, err = file.Read(buf)
	if err != nil {
		return "", err
	}
	contentType := http.DetectContentType(buf)
	return contentType, err
}

func convertWinstonToAr(winston *big.Int) string {
	// Define the conversion factor: 1 AR = 10^12 Winston
	conversionFactor := new(big.Float).SetFloat64(1e12)
	// Convert winston to big.Float for division
	winstonFloat := new(big.Float).SetInt(winston)
	// Divide winston by the conversion factor
	ar := new(big.Float).Quo(winstonFloat, conversionFactor)
	return ar.Text('f', 6)
}

func confirm() bool {
	input := ""
	fmt.Println("Do you confirm? (type 'yes' to confirm)")
	fmt.Scanln(&input)
	return strings.ToLower(input) == "yes"
}
