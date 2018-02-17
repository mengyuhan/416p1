/*

A trivial application to illustrate how the blockartlib library can be
used from an application in project 1 for UBC CS 416 2017W2.

Usage:
go run art-app.go
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
	//fmt.Print(canvas, "ignore")
	validateNum := uint8(2)

	// Add a square.
	_, _, ink, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 600 600 l 39 0 v 39 h -39 z", "filled", "red")
	if checkError(err) != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("after add a square, ink remaining is %d\n", ink)

	// // Add 凸
	_, _, ink2, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 800 800 l 50 0 l 0 50 h 50 v 50  h -150 v -50 h 50 z", "transparent", "blue")
	if checkError(err) != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("after add 凸, ink remaining is %d\n", ink2)
	// // Add 凹
	_, _, ink3, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 500 500 l 30 0 l 0 30 h 30 v -30  h 30 v 60 h -90 z", "transparent", "green")
	if checkError(err) != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("after add 凹, ink remaining is %d\n", ink3)

	_, err := canvas.CloseCanvas()
	if checkError(err) != nil {
		return
	}
}

// If error is non-nil, print it out and return it.
func checkError(err error) error {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error ", err.Error())
		return err
	}
	return nil
}
