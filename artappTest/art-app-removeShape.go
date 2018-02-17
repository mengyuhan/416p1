/*

Add and remove a single shape.

Usage:
go run art-app.go <miner-addr:art-app-port> <privKey>
*/

package main

// Expects blockartlib.go to be in the ./blockartlib/ dir, relative to
// this art-app.go file
import (
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"os"

	"../blockartlib"
)

func main() {
	// minerAddr := "127.0.0.1:8088"
	// privKey := // TODO: use crypto/ecdsa to read pub/priv keys from a file argument.

	if len(os.Args) != 3 {
		fmt.Println("Server address [ip:port] privatekeyString")
		return
	}
	minerAddr := os.Args[1]
	privString := os.Args[2]
	privateKeyBytesRestored, _ := hex.DecodeString(privString)
	privKey, _ := x509.ParseECPrivateKey(privateKeyBytesRestored)

	// Open a canvas.
	// canvas, settings, err := blockartlib.OpenCanvas(minerAddr, *privKey)
	canvas, _, err := blockartlib.OpenCanvas(minerAddr, *privKey)
	if checkError(err) != nil {
		fmt.Println(err)
		return
	}

	validateNum := uint8(2)
	fmt.Print(canvas, "ignore", validateNum)

	/************************
	Add a parallelogram
	*************************/
	shapeHash, blockHash, ink, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 1000 0 l 50 0 l 50 50 h -50 z", "transparent", "red")
	if checkError(err) != nil {
		return
	}
	println("----------------------------")
	fmt.Print(shapeHash, blockHash, ink)
	println("----------------------------")

	// Delete the first line.
	ink3, err := canvas.DeleteShape(validateNum, shapeHash)
	if checkError(err) != nil {
		return
	}
	fmt.Println(ink3)

	// assert ink3 > ink2

	// Close the canvas.
	ink4, err := canvas.CloseCanvas()
	if checkError(err) != nil {
		return
	}
	fmt.Println(ink4)
}

// If error is non-nil, print it out and return it.
func checkError(err error) error {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error ", err.Error())
		return err
	}
	return nil
}
