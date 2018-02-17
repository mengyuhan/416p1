/*

Simple set of tests for server.go used in project 1 of UBC CS 416
2017W2. Runs through the server's RPCs and their error codes.

Usage:

$ go run tester.go
  -b int
    	Heartbeat interval in ms (default 10)
  -i string
    	RPC server ip:port
  -p int
    	start port (default 54320)

*/

package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"runtime"
	"time"
)

type MinerInfo struct {
	Address net.Addr
	Key     ecdsa.PublicKey
}

// Settings for a canvas in BlockArt.
type CanvasSettings struct {
	// Canvas dimensions
	CanvasXMax uint32 `json:"canvas-x-max"`
	CanvasYMax uint32 `json:"canvas-y-max"`
}

type MinerSettings struct {
	// Hash of the very first (empty) block in the chain.
	GenesisBlockHash string `json:"genesis-block-hash"`

	// The minimum number of ink miners that an ink miner should be
	// connected to.
	MinNumMinerConnections uint8 `json:"min-num-miner-connections"`

	// Mining ink reward per op and no-op blocks (>= 1)
	InkPerOpBlock   uint32 `json:"ink-per-op-block"`
	InkPerNoOpBlock uint32 `json:"ink-per-no-op-block"`

	// Number of milliseconds between heartbeat messages to the server.
	HeartBeat uint32 `json:"heartbeat"`

	// Proof of work difficulty: number of zeroes in prefix (>=0)
	PoWDifficultyOpBlock   uint8 `json:"pow-difficulty-op-block"`
	PoWDifficultyNoOpBlock uint8 `json:"pow-difficulty-no-op-block"`
}

// Settings for an instance of the BlockArt project/network.
type MinerNetSettings struct {
	MinerSettings

	// Canvas settings
	CanvasSettings CanvasSettings `json:"canvas-settings"`
}

func exitOnError(prefix string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s, err = %s\n", prefix, err.Error())
		os.Exit(1)
	}
}

var ExpectedError = errors.New("Expected error, none found")

///////////////////////////////////////////added
type MinerRPCs interface {
	Connect(privatekey string, reply *ValidMiner) error
	GetInk(privatekey string, reply *uint32) error
	AddShape(args AddShapeStruct, reply *AddShapeReply) error
	GetSvgString(shapeHash string, svgString *string) error
	DeleteShape(args DelShapeArgs, inkRemaining *uint32) error
	GetShapes(blockHash string, shapeHashes *[]string) error
	GetGenesisBlock(args int, blockHash *string) error
	GetChildren(blockHash string, blockHashes *[]string) error
	CloseCanvas(args int, inkRemaining *uint32) error
}

type MinerRPC int

type ValidMiner struct {
	MinerNetSets MinerNetSettings
	Valid        bool
}

type ShapeType int

const (
	// Path shape.
	PATH ShapeType = iota
	// Circle shape (extra credit).
	// CIRCLE
)

type AddShapeStruct struct {
	ValidateNum    uint8
	SType          ShapeType
	ShapeSvgString string
	Fill           string
	Stroke         string
	ArtNodePK      string
}

type AddShapeReply struct {
	ShapeHash    string
	BlockHash    string
	InkRemaining string
}

type DelShapeArgs struct {
	validateNum uint8
	shapeHash   string
	ArtNodePK   string
}

var myKeyPairInString string

// var recordedArtnodes map[string]bool
func (m *MinerRPC) Connect(minerprivatekey string, reply *ValidMiner) error {
	// rlp := CanvasSettings{2, 3}
	var v ValidMiner
	if myKeyPairInString == minerprivatekey {
		v = ValidMiner{Valid: true}
		fmt.Println("validKey:", minerprivatekey)
	}
	// recordedArtnodes[]
	//
	*reply = v
	return nil
}

func (m *MinerRPC) GetInk(minerprivatekey string, reply *uint32) error {
	if myKeyPairInString == minerprivatekey {
		// *reply= remaining amount of ink of this miner
		fmt.Println("@@@GetInk")
		return nil
	}
	return nil //return error
}

func (m *MinerRPC) AddShape(args AddShapeStruct, reply *AddShapeReply) error {
	// check ink sufficient
	// check shape overlap/ out of boundry
	// try add this shape return shape/block hash, remained ink
	svgStr := "<path d=\"" + args.ShapeSvgString + "\" stroke=\"" +
		args.Stroke + "\" fill=\"" + args.Fill + "\"/>"
	*reply = AddShapeReply{ShapeHash: svgStr}
	return nil
}

func (m *MinerRPC) GetSvgString(shapeHash string, svgString *string) error {
	// do we store shapehash pair with svgstring locally, or do we get it by go through blockchain?
	// *svgString = getbyhash(shapeHash)
	fmt.Println("@@@ GetSvgString")
	return nil
}

func (m *MinerRPC) DeleteShape(args DelShapeArgs, inkRemaining *uint32) error {
	// try delete shape by args
	*inkRemaining = 123
	fmt.Println("@@@ DeleteShape")
	return nil
}

func (m *MinerRPC) GetShapes(blockHash string, shapeHashes *[]string) error {
	// get shapeHashes
	fmt.Println("@@@ GetShapes")
	return nil
}

func (m *MinerRPC) GetGenesisBlock(args int, blockHash *string) error {
	// blockHash = hash of GenesisBlock
	fmt.Println("@@@ GetGenesisBlock")
	return nil
}

func (m *MinerRPC) GetChildren(blockHash string, blockHashes *[]string) error {
	// blockHashes = children of blockHash
	fmt.Println("@@@ GetChildren")
	return nil
}

func (m *MinerRPC) CloseCanvas(args int, inkRemaining *uint32) error {
	// close connection, anything else need do? like set online artnode to false
	//inkRemaining =
	fmt.Println("@@@ CloseCanvas")
	return nil
}

func registerServer(server *rpc.Server, s MinerRPCs) {
	// registers interface by name of `MyServer`.
	server.RegisterName("InkMinerRPC", s)
}

func main() {
	// gob.Register(&net.TCPAddr{})
	// gob.Register(&elliptic.CurveParams{})

	// ipPort := flag.String("i", "", "RPC server ip:port")
	// startPort := flag.Int("p", 54320, "start port")
	// heartBeat := flag.Int("b", 10, "Heartbeat interval in ms")
	// flag.Parse()
	// if *ipPort == "" || *startPort <= 1024 || *heartBeat <= 0 {
	// 	flag.PrintDefaults()
	// 	os.Exit(1)
	// }

	// // heartBeatInterval := time.Duration(*heartBeat) * time.Millisecond
	// twoHeartBeatIntervals := time.Duration(*heartBeat*2) * time.Millisecond

	// r, err := os.Open("/dev/urandom")
	// exitOnError("open /dev/urandom", err)
	// defer r.Close()

	priv1, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	exitOnError("generate key 1", err)
	// fmt.Println(priv1)
	privateKeyBytes, _ := x509.MarshalECPrivateKey(priv1)

	myKeyPairInString = hex.EncodeToString(privateKeyBytes)
	fmt.Println("s:", myKeyPairInString, ":end")

	// pp := "3081a40201010430671b56e9a7419ad72c9ae6c4c96547083d12b642757005dccc4483303f5ad1ba9f8ec27121d5b6386e09e3c1e313dad9a00706052b81040022a16403620004c6536f60d28f8e2819de327dd5264e468dc2044f03ae158e4c1fe4a59773b0cc945b7128a1f77401e2dd57245983d2d346214d5badb7b97774e29b7905f9f10be84c03a054fca49d41fc735303c7ae6149c881d1b20aeed369444377c3f6ffd8"
	// privateKeyBytesRestored, _ := hex.DecodeString(pp)
	// priv2, _ := x509.ParseECPrivateKey(privateKeyBytesRestored)
	// fmt.Println("ddd", priv2.PublicKey)

	// priv2, err := ecdsa.GenerateKey(elliptic.P384(), r)
	// exitOnError("generate key 2", err)

	// addr1, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("127.0.0.1:%d", *startPort))
	// exitOnError("resolve addr 1", err)
	// addr2, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("127.0.0.1:%d", *startPort+1))
	// exitOnError("resolve addr2", err)

	// c, err := rpc.Dial("tcp", *ipPort)
	// exitOnError("rpc dial", err)
	// defer c.Close()

	// var settings MinerNetSettings
	// var _ignored bool

	// // normal registration
	// err = c.Call("RServer.Register", MinerInfo{Address: addr1, Key: priv1.PublicKey}, &settings)
	// exitOnError(fmt.Sprintf("client registration for %s", addr1.String()), err)
	// err = c.Call("RServer.Register", MinerInfo{Address: addr2, Key: priv2.PublicKey}, &settings)
	// exitOnError(fmt.Sprintf("client registration for %s", addr2.String()), err)
	// time.Sleep(twoHeartBeatIntervals)

	// // late heartbeat
	// err = c.Call("RServer.Register", MinerInfo{Address: addr1, Key: priv1.PublicKey}, &settings)
	// exitOnError(fmt.Sprintf("client registration for %s", addr1.String()), err)
	// time.Sleep(twoHeartBeatIntervals)
	// err = c.Call("RServer.HeartBeat", priv1.PublicKey, &_ignored)
	// if err == nil {
	// 	exitOnError("late heartbeat", ExpectedError)
	// }

	// // register twice with same address
	// err = c.Call("RServer.Register", MinerInfo{Address: addr1, Key: priv1.PublicKey}, &settings)
	// exitOnError(fmt.Sprintf("client registration for %s", addr1.String()), err)
	// err = c.Call("RServer.Register", MinerInfo{Address: addr1, Key: priv2.PublicKey}, &settings)
	// if err == nil {
	// 	exitOnError("registering twice with the same address", ExpectedError)
	// }
	// time.Sleep(twoHeartBeatIntervals)

	// // register twice with same key
	// err = c.Call("RServer.Register", MinerInfo{Address: addr1, Key: priv1.PublicKey}, &settings)
	// exitOnError(fmt.Sprintf("client registration for %s", addr1.String()), err)
	// err = c.Call("RServer.Register", MinerInfo{Address: addr2, Key: priv1.PublicKey}, &settings)
	// if err == nil {
	// 	exitOnError("registering twice with the same key", ExpectedError)
	// }

	//////////////////////////////////////////
	if len(os.Args) != 2 {
		fmt.Println("Need Server address [ip:port]")
		return
	}

	serverAddr := os.Args[1]
	mRPC := new(MinerRPC)

	server := rpc.NewServer()
	registerServer(server, mRPC)

	// Listen for incoming tcp packets on specified port.
	l, e := net.Listen("tcp", serverAddr)
	if e != nil {
		log.Fatal("listen error:", e)
	}

	go server.Accept(l)
	runtime.Gosched()
	fmt.Println("done")
	time.Sleep(10000 * time.Second)
}
