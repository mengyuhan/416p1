/*

Test for overlapping shapes with the same artnode. This art app will
-Add a triangle
-Add a line which overlaps with the triangle

Usage:
go run art-app.go port
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
	"os"

	"../blockartlib"
)

func main() {
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

	shapeHash, blockHash, ink, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 400 0 L 0 400 h 800 l -400 -400", "non-transparent", "red")
	if checkError(err) != nil {
		return
	}
	fmt.Println(shapeHash)
	fmt.Println(blockHash)
	fmt.Println(ink)

	shapeHash2, blockHash2, ink2, err := canvas.AddShape(validateNum, blockartlib.PATH, "300 0 L 0 5000", "transparent", "red")
	if checkError(err) != nil {
		return
	}
	fmt.Print(shapeHash2, blockHash2, ink2)

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
