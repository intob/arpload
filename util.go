package main

import (
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
)

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
