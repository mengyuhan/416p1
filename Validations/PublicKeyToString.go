package main

import (
	"crypto/x509"
	"encoding/hex"
	"fmt"
)

var (
	privKey string = "3081a4020101043009ef49f461356155deb35307604cc6302a2224ecce36783cc6a8ea1d157104744f7c298cc043c08906006153812e3959a00706052b81040022a164036200042d4d058dfabc1ec4b7894e11706cf3c75a77670d209578c73ccf835062f73127c47ab66c7e2640c30ec7173c13e3a53ab77c019dfe31fccd862a1cff3501a47982513ef7a06fe3d34105e36efad2de17c17da89703a77f10730ac74b63b9af42"
)

func main() {
	keyAsBytes, _ := hex.DecodeString(privKey)
	myPrivKey, _ := x509.ParseECPrivateKey(keyAsBytes)

	str := fmt.Sprintf("%s%s", myPrivKey.PublicKey.X, myPrivKey.PublicKey.Y)
	fmt.Println(str)
}
