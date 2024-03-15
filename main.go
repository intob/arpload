package main

import (
	"encoding/json"
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

type Job struct {
	Gateway,
	WalletFile,
	Fname,
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

	job := &Job{
		Gateway:     "https://arweave.net",
		WalletFile:  os.Getenv("AR_WALLET"),
		Fname:       fname,
		ContentType: typ,
		Title:       title,
		Description: desc,
		Author:      authr,
	}
	if job.WalletFile == "" {
		log.Fatal("AR_WALLET must be defined in your env")
	}
	if job.ContentType == "" {
		fmt.Println("warning: Content-Type header will be blank, use -type to set it")
	}

	data, err := os.ReadFile(fname)
	if err != nil {
		log.Fatal(err)
	}

	w, err := goar.NewWalletFromPath(job.WalletFile, job.Gateway)
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
	txId, err := job.sendDataStream(w)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\n🌳 Done. Tx:%s\n", txId)
}

func (j *Job) sendDataStream(w *goar.Wallet) (string, error) {
	f, err := os.Open(j.Fname)
	if err != nil {
		return "", err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	// resume or create new uploader
	uploader, err := j.readUploader(w, data)
	if uploader == nil && err == nil {
		// create new one
		uploader, err = j.createUploader(w, data)
		if err != nil {
			return "", fmt.Errorf("err starting new upload: %w", err)
		}
		fmt.Println("starting new upload:", uploader.Transaction.ID)
	} else {
		fmt.Printf("resuming upload at chunk %d: %s\n", uploader.ChunkIndex, uploader.Transaction.ID)
	}

	jsonFname := uploaderJsonFname(j.Fname)

	for !uploader.IsComplete() {
		err := uploader.UploadChunk()
		if err != nil {
			return uploader.Transaction.ID, fmt.Errorf("upload chunk failed: %w", err)
		}
		uploaderJson, err := json.Marshal(uploader)
		if err != nil {
			return uploader.Transaction.ID, err
		}

		err = os.WriteFile(jsonFname, uploaderJson, 0777)
		if err != nil {
			return uploader.Transaction.ID, fmt.Errorf("err writing uploader json file (%s): %w", jsonFname, err)
		}
		fmt.Printf("\r%.2f%%\033[0K", uploader.PctComplete())
	}

	// remove uploader json
	err = os.Remove(jsonFname)
	if err != nil {
		return "", fmt.Errorf("err removing uploader json: %w", err)
	}

	return uploader.Transaction.ID, nil
}

func (j *Job) readUploader(w *goar.Wallet, data []byte) (*goar.TransactionUploader, error) {
	uploaderBuf, err := os.ReadFile(uploaderJsonFname(j.Fname))
	if err != nil {
		return nil, nil
	}
	lastUploader := &goar.TransactionUploader{}
	err = json.Unmarshal(uploaderBuf, lastUploader)
	if err != nil {
		return nil, fmt.Errorf("err unmarshaling uploader: %w", err)
	}

	newUploader, err := goar.CreateUploader(w.Client, lastUploader.FormatSerializedUploader(), data)
	if err != nil {
		return nil, fmt.Errorf("err creating uploader: %w", err)
	}

	err = newUploader.Once()
	if err != nil {
		return nil, fmt.Errorf("err calling uploader's Once(): %w", err)
	}

	return newUploader, nil
}

func (j *Job) createUploader(w *goar.Wallet, data []byte) (*goar.TransactionUploader, error) {
	tags := []types.Tag{
		{Name: "Content-Type", Value: j.ContentType},
		{Name: "Title", Value: j.Title},
		{Name: "Description", Value: j.Description},
		{Name: "Author", Value: j.Author},
	}

	tx, err := assemblyDataTx(data, w, tags)
	if err != nil {
		return nil, err
	}
	fmt.Println("tx id: ", tx.ID)
	fmt.Printf("will upload file with tags: %+v\n", tags)
	if !confirm() {
		return nil, errors.New("user aborted")
	}

	uploader, err := goar.CreateUploader(w.Client, tx, nil)
	if err != nil {
		return nil, err
	}

	return uploader, nil
}

func uploaderJsonFname(fname string) string {
	return fmt.Sprintf("%s.uploader.json", fname)
}
