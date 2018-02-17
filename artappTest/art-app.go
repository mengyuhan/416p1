/*

A trivial application to illustrate how the blockartlib library can be
used from an application in project 1 for UBC CS 416 2017W2.

Usage:
go run art-app.go port privKey
*/

package main

// Expects blockartlib.go to be in the ./blockartlib/ dir, relative to
// this art-app.go file
import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"../blockartlib"
)

func main() {
	args := os.Args[1:]
	minerPort := args[0]
	privKey := args[1]
	fmt.Println(minerPort)
	fmt.Println(privKey)

	minerAddr := "127.0.0.1:" + minerPort
	//privKey := // TODO: use crypto/ecdsa to read pub/priv keys from a file argument.

	// privKey := flag.String("i", "", "RPC server ip:port")
	// Open a canvas.
	//var key ecdsa.PrivateKey

	keyAsBytes, _ := hex.DecodeString(privKey)
	myPrivKey, _ := x509.ParseECPrivateKey(keyAsBytes)
	fmt.Println(myPrivKey)

	key := decode(myPrivKey)
	fmt.Println(myPrivKey)
	canvas, settings, err := blockartlib.OpenCanvas(minerAddr, *key)

	println(settings.CanvasXMax)
	if checkError(err) != nil {
		return
	}
	var validateNum uint8
	validateNum = 2

	// Add a line.
	shapeHash, blockHash, ink, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 0 0 L 0 5", "transparent", "red")
	if checkError(err) != nil {
		return
	}
	fmt.Println(shapeHash)
	fmt.Println(blockHash)
	fmt.Println(ink)

	// Add another line.
	shapeHash2, blockHash2, ink2, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 0 0 L 5 0", "transparent", "blue")
	if checkError(err) != nil {
		return
	}
	fmt.Println(shapeHash2)
	fmt.Println(blockHash2)
	fmt.Println(ink2)

	// Delete the first line.
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
	// block, _ := pem.Decode([]byte(privateKey))
	newPrivateKey := "-----BEGIN PRIVATE KEY-----\n" + privateKey + "\n-----END PRIVATE KEY-----"
	// Cite: https: //play.golang.org/p/T0jR2uzGp5
	r := strings.NewReader(newPrivateKey)
	pemBytes, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}

	block, _ := pem.Decode(pemBytes)
	if block == nil {
		log.Println(block)
	}
	fmt.Println(block)
	x509Encoded := block.Bytes
	pKey, _ := x509.ParseECPrivateKey(x509Encoded)

	return pKey
}
