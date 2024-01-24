package main

import (
	"fmt"
	"io"
	"log"
	"math/big"
	"os"

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

	f, err := os.Open(opt.Filename)
	if err != nil {
		return err
	}
	defer f.Close()

	tags := []types.Tag{
		{Name: "Content-Type", Value: contentType},
		{Name: "Title", Value: opt.Title},
		{Name: "Description", Value: opt.Description},
		{Name: "Author", Value: opt.Author},
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	tx, err := assemblyDataTx(data, w, tags)
	if err != nil {
		return err
	}
	fmt.Println("tx id: ", tx.ID)

	uploader, err := goar.CreateUploader(w.Client, tx, nil)
	if err != nil {
		return err
	}

	return uploader.Once()
}
