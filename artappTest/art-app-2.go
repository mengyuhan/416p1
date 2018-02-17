/*

A trivial application to illustrate how the blockartlib library can be
used from an application in project 1 for UBC CS 416 2017W2.

Usage:
go run art-app.go port
*/

package main

// Expects blockartlib.go to be in the ./blockartlib/ dir, relative to
// this art-app.go file
import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"os"

	"../blockartlib"
)

func main() {
	args := os.Args[1:]
	minerPort := args[0]
	fmt.Println(minerPort)

	minerAddr := "127.0.0.1:" + minerPort
	//privKey := // TODO: use crypto/ecdsa to read pub/priv keys from a file argument.
	privKey := flag.String("i", "", "RPC server ip:port")
	println(*privKey)
	// Open a canvas.
	var key ecdsa.PrivateKey
	key = *decode(*privKey)
	canvas, settings, err := blockartlib.OpenCanvas(minerAddr, key)
	println(settings.CanvasXMax)
	if checkError(err) != nil {
		return
	}
	var validateNum uint8
	validateNum = 2

	shapeHash, blockHash, ink, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 500 0 l 500 0 l 500 500 h -500 z", "non-transparent", "red")
	if checkError(err) != nil {
		return
	}
	fmt.Println(shapeHash)
	fmt.Println(blockHash)
	fmt.Println(ink)

	// Add a line.
	shapeHash2, blockHash2, ink2, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 0 499 H 500", "non-transparent", "blue")
	if checkError(err) != nil {
		return
	}
	fmt.Println(shapeHash2)
	fmt.Println(blockHash2)
	fmt.Println(ink2)

	// // Delete the first line.
	// ink3, err := canvas.DeleteShape(validateNum, shapeHash)
	// if checkError(err) != nil {
	// 	return
	// }
	// fmt.Println(ink3)

	// assert ink3 > ink2

	// Close the canvas.
	ink4, err := canvas.CloseCanvas()
	if checkError(err) != nil {
		return
	}
	println(ink4)
}

// If error is non-nil, print it out and return it.
func checkError(err error) error {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error ", err.Error())
		return err
	}
	return nil
}

func decode(privateKey string) *ecdsa.PrivateKey {
	block, _ := pem.Decode([]byte(privateKey))
	x509Encoded := block.Bytes
	pKey, _ := x509.ParseECPrivateKey(x509Encoded)

	return pKey
}
