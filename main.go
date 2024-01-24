package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"

	"github.com/everFinance/goar"
	"github.com/everFinance/goar/types"
)

type Options struct {
	Gateway,
	WalletFile,
	Filename,
	ContentType,
	Title,
	Description,
	Author string
}

func main() {
	var typ, title, desc, authr string
	flag.StringVar(&typ, "type", "", "set the value of content-type http header")
	flag.StringVar(&title, "title", "", "set the value of title http header")
	flag.StringVar(&desc, "desc", "", "set the value of description http header")
	flag.StringVar(&authr, "author", "", "set the value of author http header")

	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatal("missing filename argument")
	}
	fname := flag.Arg(0)

	opt := &Options{
		Gateway:     "https://arweave.net",
		WalletFile:  os.Getenv("AR_WALLET"),
		Filename:    fname,
		ContentType: typ,
		Title:       title,
		Description: desc,
		Author:      authr,
	}
	if opt.WalletFile == "" {
		log.Fatal("AR_WALLET must be defined in your env")
	}
	if opt.ContentType == "" {
		fmt.Println("warning: Content-Type header will be blank, use -type to set it")
	}

	data, err := os.ReadFile(fname)
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
	fmt.Println("🌳 done")
}

func sendDataStream(w *goar.Wallet, opt *Options) error {
	f, err := os.Open(opt.Filename)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	tags := []types.Tag{
		{Name: "Content-Type", Value: opt.ContentType},
		{Name: "Title", Value: opt.Title},
		{Name: "Description", Value: opt.Description},
		{Name: "Author", Value: opt.Author},
	}

	tx, err := assemblyDataTx(data, w, tags)
	if err != nil {
		return err
	}
	fmt.Println("tx id: ", tx.ID)
	fmt.Printf("will upload file with tags: %+v\n", tags)
	if !confirm() {
		return errors.New("user aborted")
	}

	uploader, err := goar.CreateUploader(w.Client, tx, nil)
	if err != nil {
		return err
	}

	for !uploader.IsComplete() {
		if err := uploader.UploadChunk(); err != nil {
			return fmt.Errorf("upload chunk failed: %w", err)
		}
		fmt.Printf("%.2f%% ", uploader.PctComplete())
	}
	return nil
}
