package main

/*
	Usage:
	go run ink-miner.go [server ip:port] [priv-key] [miner listen port] [art-app listen port]
*/

// package ink-miner

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/md5"
	"crypto/x509"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"../SvgHelper"
)

var (
	blockChain        []Block = make([]Block, 0)
	_ignored          bool
	settings          MinerNetSettings
	myMinerInfo       MinerInfo
	minersConnectedTo allMinersConnectedTo = allMinersConnectedTo{currentNumNeighbours: 0, all: make([]string, 10)}
	myPrivKey         *ecdsa.PrivateKey
	serverIPPOrt      string
	miners            []net.Addr
	localIPPortStr    string
	localIPPortArr    [2]string
	artAppListenPort  string
	globalPubKeyStr   string = ""
)

type allMinersConnectedTo struct {
	sync.RWMutex
	currentNumNeighbours int
	all                  []string // network address of neighbour miners
}

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

type Operation struct {
	AppShape      string
	OpSig         string
	PubKeyArtNode string //key of the art node that generated the op
	ShapeCommand  string // e.g. "M 0 0 L 0 3"
	ShapeFill     string // fill or transparent
}

type Coordinate struct {
	x int
	y int
}

type PixelState struct {
	n           int    // number of overlapping shapes on the given x-y coordinate
	minerPubKey string // miner who "owns" the current pixel on shared canvas
}

type InkAccount struct {
	inkMined  uint32
	inkSpent  uint32
	inkRemain uint32
}

type Block struct {
	PrevHash         string // MD5 hash with 0s
	Nonce            uint32
	Ops              []Operation
	NoOpBlock        bool // if a NoOpBlock, then true. False otherwise
	PubKeyMiner      string
	Index            int
	MinerInks        map[string]InkAccount
	CanvasInks       map[string]SvgHelper.MapPoint
	CanvasOperations map[string][]string // Ink Miner to List of Operations on canvas
}

/********************************
Structs for RPC calls for Artnode
********************************/
type MinerRPCs interface {
	// Art node to Miner RPC
	Connect(privatekey string, reply *ValidMiner) error
	GetInk(privatekey string, reply *uint32) error
	AddShape(args AddShapeStruct, reply *AddShapeReply) error
	GetSvgString(shapeHash string, svgString *string) error
	DeleteShape(args DelShapeArgs, inkRemaining *uint32) error
	GetShapes(blockHash string, shapeHashes *[]string) error
	GetGenesisBlock(args int, blockHash *string) error
	GetChildren(blockHash string, blockHashes *[]string) error
	CloseCanvas(args int, reply *CloseCanvReply) error
}

func getBlockchain() []Block {
	return blockChain
}

/*************************************
Structs for RPC calls for miner2miner
**************************************/
type Miner2MinerRPCs interface {
	PrintText(textToPrint string, reply *string) error
	EstablishReverseRPC(addr string, reply *string) error
	SendBlockchain(bc *[]Block, reply *string) error
}

// Interface between art app and ink miner
type MinerRPC int

// Interface between ink miner to ink miner
type MinerToMinerRPC int

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
	InkRemaining uint32
}

var myKeyPairInString string

type DelShapeArgs struct {
	validateNum uint8
	shapeHash   string
	ArtNodePK   string
}

type CloseCanvReply struct {
	Ops          []Operation
	inkRemaining uint32
}

// Provided by server.go code as part of repository
func exitOnError(prefix string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s, err = %s\n", prefix, err.Error())
		os.Exit(1)
	}
}

var ExpectedError = errors.New("Expected error, none found")
var UnknownKeyError = errors.New("Server does not know given miner key")

type InvalidMinerPKError string

func (e InvalidMinerPKError) Error() string {
	return fmt.Sprintf("BlockArt: Invalid miner's private/public key [%s]", string(e))
}

type InvalidShapeHashError string

func (e InvalidShapeHashError) Error() string {
	return fmt.Sprintf("BlockArt: Invalid shape hash [%s]", string(e))
}

type ShapeOwnerError string

func (e ShapeOwnerError) Error() string {
	return fmt.Sprintf("BlockArt: Shape owned by someone else [%s]", string(e))
}

type InvalidBlockHashError string

func (e InvalidBlockHashError) Error() string {
	return fmt.Sprintf("BlockArt: Invalid block hash [%s]", string(e))
}

type InsufficientInkError uint32

func (e InsufficientInkError) Error() string {
	return fmt.Sprintf("BlockArt: Not enough ink to addShape [%d]", uint32(e))
}

func main() {
	// Read in command line args
	// args[0] is server:port, args[1] is private key, args[2] is miner port, args[3] is art-app port
	args := os.Args[1:]
	ipPort := args[0]
	myKeyPairInString = args[1]
	port := args[2]
	artAppListenPort = args[3]
	portInt, err := strconv.Atoi(port)
	if err != nil {
		exitOnError("Port is invalid", err)
	}

	conn, _ := net.Dial("tcp", ipPort)
	localIPPortArr := strings.Split(conn.LocalAddr().String(), ":")
	fmt.Println("my address")
	fmt.Println(localIPPortArr)
	conn.Close()
	localIPPortStr = fmt.Sprintf("%s:%d", localIPPortArr[0], portInt)
	fmt.Println(localIPPortStr)
	// Register with the Server and get settings
	addr, err := net.ResolveTCPAddr("tcp", localIPPortStr)

	exitOnError("resolve addr", err)
	keyAsBytes, _ := hex.DecodeString(myKeyPairInString)
	myPrivKeyLocal, _ := x509.ParseECPrivateKey(keyAsBytes)
	myPrivKey = myPrivKeyLocal

	gob.Register(&net.TCPAddr{})
	gob.Register(&elliptic.CurveParams{})

	// Establish RPC connection to server
	cRPC, err := rpc.Dial("tcp", ipPort)
	defer cRPC.Close()
	if err != nil {
		fmt.Println("Error dialing to server ", err.Error)
	}
	fmt.Println("Miner address is ====== " + addr.String())
	myMinerInfo = MinerInfo{Address: addr, Key: myPrivKey.PublicKey}
	err = cRPC.Call("RServer.Register", myMinerInfo, &settings)
	exitOnError(fmt.Sprintf("client registration for %s", myMinerInfo.Address), err)
	listenToArtnode(ipPort)

	go sendHeartBeats(ipPort, myMinerInfo, settings.HeartBeat)
	go listenForIncomingConnections(portInt)

	go monitorNumConnections(ipPort)

	for {
		sleep_time := 100 * time.Millisecond
		time.Sleep(sleep_time)

		//fmt.Println("Main still alive")
		if globalPubKeyStr == "" {
			myPubKeyStr := getPubKeyInStr(myPrivKey.PublicKey)
			globalPubKeyStr = myPubKeyStr
			//fmt.Printf("In main, global pubkeystr is: %s\n", globalPubKeyStr)
		}

		mineNoOpBlocks(globalPubKeyStr)
		//fmt.Printf("Mined a block. Blockchain is now %d\n", len(blockChain))
		//lastOne := len(blockChain) - 1
		//fmt.Printf("lastOne index: %d\n", lastOne)
		//fmt.Printf("Last blk index: %d\n", blockChain[lastOne].Index)
		//fmt.Printf("globalPubKeyStr: %s\n", globalPubKeyStr)
		// inkMinedRightNow := blockChain[lastOne].MinerInks[globalPubKeyStr].inkMined
		// inkRemainingRightNow := blockChain[lastOne].MinerInks[globalPubKeyStr].inkRemain

		//fmt.Printf("My ink mined is %d remaining is: %d\n", inkMinedRightNow, inkRemainingRightNow)
	}
}

// This function mines NoOpBlocks idly
func mineNoOpBlocks(minerPubKey string) {
	blockChain = append(blockChain, generateNoOpBlock(minerPubKey))
}

func generateNoOpBlock(minerPubKey string) Block {
	var difficulty uint8

	if len(blockChain) < 1 {
		blk, _ := generateFirstBlock()
		return blk
	}

	lastBlockIndex := len(blockChain) - 1
	lastBlk := blockChain[lastBlockIndex]
	if lastBlk.NoOpBlock {
		difficulty = settings.PoWDifficultyNoOpBlock
	} else {
		difficulty = settings.PoWDifficultyOpBlock
	}

	opsArr := make([]Operation, 0)
	cInks := lastBlk.CanvasInks
	cOps := lastBlk.CanvasOperations

	lastBlkHash, _ := calculateHash(lastBlk, difficulty)

	blk := Block{
		PrevHash:         lastBlkHash,
		Nonce:            0,
		Ops:              opsArr,
		NoOpBlock:        true,
		PubKeyMiner:      globalPubKeyStr,
		Index:            lastBlockIndex + 1,
		MinerInks:        lastBlk.MinerInks,
		CanvasInks:       cInks,
		CanvasOperations: cOps,
	}

	oldMinerInks := lastBlk.MinerInks

	if myInkAccount, ok := oldMinerInks[minerPubKey]; ok {
		fmt.Println("incrementing ink")
		fmt.Println(myInkAccount)
		myInkAccount.inkMined = myInkAccount.inkMined + settings.InkPerNoOpBlock
		myInkAccount.inkRemain = myInkAccount.inkRemain + settings.InkPerNoOpBlock
		oldMinerInks[minerPubKey] = myInkAccount
		str := minerPubKey
		fmt.Printf("in gen noop block: %s\n", str)
		blk.MinerInks = oldMinerInks
	} else {
		fmt.Println("setting ink for first time")
		oldMinerInks[minerPubKey] = InkAccount{inkMined: settings.InkPerNoOpBlock, inkSpent: 0, inkRemain: settings.InkPerNoOpBlock}
		fmt.Println(oldMinerInks)
		blk.MinerInks = oldMinerInks
	}

	_, currNonce := calculateHash(blk, settings.PoWDifficultyNoOpBlock)
	nonceUInt64, _ := strconv.ParseUint(currNonce, 10, 32)
	blk.Nonce = uint32(nonceUInt64)

	return blk
}

/***************************
Block validation helpers
****************************/

func generateBlock(oldBlock Block) (Block, error) {
	var newBlock Block
	newBlock.Index = oldBlock.Index + 1

	return newBlock, nil
}

func generateFirstBlock() (Block, error) {
	opsArr := make([]Operation, 0)
	mInks := make(map[string]InkAccount)
	mInks[globalPubKeyStr] = InkAccount{inkMined: settings.InkPerNoOpBlock, inkRemain: settings.InkPerNoOpBlock, inkSpent: 0}
	cInks := make(map[string]SvgHelper.MapPoint)
	cOps := make(map[string][]string)

	blk := Block{
		PrevHash:         settings.GenesisBlockHash,
		Nonce:            0,
		Ops:              opsArr,
		NoOpBlock:        true,
		PubKeyMiner:      globalPubKeyStr,
		Index:            1,
		MinerInks:        mInks,
		CanvasInks:       cInks,
		CanvasOperations: cOps,
	}

	return blk, nil
}

func blkToString(b Block) string {
	return b.PrevHash + convertOpToString(b.Ops) + b.PubKeyMiner + string(b.Index)
}

// [prev-hash, op, op-signature, pub-key, nonce, other data structures]
func calculateHash(b Block, powDifficulty uint8) (hash, nonce string) {
	blockString := blkToString(b)

	j := int64(0)
	for {
		nonce = strconv.FormatInt(j, 10)
		hash = computeNonceSecretHash(blockString, nonce)

		if hasNZeros(hash, powDifficulty) {
			break
		}
		j++
	}
	return hash, nonce
}

// [prev-hash, op, op-signature, pub-key, nonce, other data structures]
func convertOpToString(ops []Operation) string {
	opsString := ""
	for _, element := range ops {
		opsString += element.AppShape + element.OpSig + element.PubKeyArtNode
	}
	return opsString
}

func hasNZeros(hash string, n uint8) bool {
	zeros := strings.Repeat("0", int(n))
	return strings.HasSuffix(hash, zeros)
}

// Returns the MD5 hash as a hex string for the (nonce + secret) value.
func computeNonceSecretHash(nonce string, secret string) string {
	h := md5.New()
	h.Write([]byte(nonce + secret))
	str := hex.EncodeToString(h.Sum(nil))
	return str
}

func isSentChainLonger(newBlocks []Block) bool {
	if len(newBlocks) > len(blockChain) {
		return true
	}

	return false
}

// Function to request additional miner nodes if the current miner is below
// the threshold
func monitorNumConnections(ipPort string) {
	for {
		sleep_time := 20000 * time.Millisecond
		time.Sleep(sleep_time)

		var neighbours []net.Addr

		helperGetNodes(ipPort, myMinerInfo, &neighbours)
		if len(neighbours) > 0 {
			fmt.Println(neighbours)
			fmt.Println("monitoring connections")
			connectToMiners(neighbours)
		}
	}
}

func isNoOpBlock(block Block) bool {
	// Return True if it is a NoOp block, False otherwise
	if len(block.Ops) == 0 {
		return true
	}
	return false
}

/***************************
Miner-Server Communication
****************************/

/*
Send heartbeats to server at regular intervals to maintain RPC connection
*/
func sendHeartBeats(ipPort string, miner MinerInfo, heartBeatInterval uint32) {

	cRPC, err := rpc.Dial("tcp", ipPort)
	defer cRPC.Close()
	if err != nil {
		fmt.Println(err.Error)
	}

	hbInMilliSec := time.Duration(heartBeatInterval) * time.Millisecond
	timeToSleep := hbInMilliSec / 20
	fmt.Println(timeToSleep)
	for {
		time.Sleep(timeToSleep)
		err = cRPC.Call("RServer.HeartBeat", miner.Key, &_ignored)
		if err != nil {
			exitOnError("late heartbeat", ExpectedError)
		}
	}
}

/*
A wrapper on the GetNodes RPC call. It invokes a GetNodes RPC call only if the
current number of connections is less than the minimum.

@returns: true if addresses were obtained and false otherwise
*/
func helperGetNodes(ipPort string, miner MinerInfo, addrSet *[]net.Addr) bool {
	minersConnectedTo.Lock()
	defer minersConnectedTo.Unlock()
	if minersConnectedTo.currentNumNeighbours < int(settings.MinNumMinerConnections) {
		fmt.Println("Inside helper get node, ready to make RPC call")
		cRPC, err := rpc.Dial("tcp", ipPort)
		defer cRPC.Close()
		if err != nil {
			fmt.Println(err.Error)
		}

		err = cRPC.Call("RServer.GetNodes", miner.Key, addrSet)
		if err != nil {
			exitOnError(miner.Address.String(), err)
		}

		return true
	}

	return false
}

func connectToMiners(addrSet []net.Addr) {
	fmt.Println("Inside connectToMiners")
	for _, addr := range addrSet {
		go connectToMiner(addr)
	}
}

/*
The miner shall establish TCP connections to the supplied neighbour miner
*/
func connectToMiner(addr net.Addr) {
	fmt.Println("Inside connectToMiner")
	// Establish RPC connection to server
	fmt.Println(addr.String())
	miner2minerRPC, err := rpc.Dial("tcp", addr.String())
	if err != nil {
		fmt.Println(err.Error)
	}
	minersConnectedTo.Lock()
	defer minersConnectedTo.Unlock()

	for _, addrAlreadyConnectedTo := range minersConnectedTo.all {
		if addr.String() == addrAlreadyConnectedTo {
			return
		}
	}

	minersConnectedTo.all = append(minersConnectedTo.all, addr.String())

	reply := ""
	err = miner2minerRPC.Call("MinerToMinerRPC.EstablishReverseRPC", myMinerInfo.Address.String(), &reply)
	if err != nil {
		fmt.Println("Issue with EstablishReverseRPC", err)
	}
	fmt.Printf("Did other side connect to me?: %s\n", reply)
	go handleMiner(*miner2minerRPC, addr)
}

/*
A handler that handles all logic between two miners
*/
func handleMiner(otherMiner rpc.Client, otherMinerAddr net.Addr) {
	defer otherMiner.Close()
	minersConnectedTo.Lock()
	minersConnectedTo.currentNumNeighbours = minersConnectedTo.currentNumNeighbours + 1
	fmt.Printf("Curr num neighbours connected to: %d\n", minersConnectedTo.currentNumNeighbours)
	minersConnectedTo.Unlock()
	reply := ""
	fmt.Println("About to make RPC call")
	err := otherMiner.Call("MinerToMinerRPC.PrintText", "Hi from your neighbour!", &reply)
	if err != nil {
		fmt.Println("Issue with RPC call in handleMiner")
	}
	fmt.Println("Finished RPC call")
	fmt.Println(reply)
	for {
		fmt.Println("Connection still alive")
		sleep_time := 5000 * time.Millisecond
		time.Sleep(sleep_time)

		var reply string
		fmt.Println("Sending block chain to neighbour")
		err := otherMiner.Call("SendBlockChain", &blockChain, &reply)
		if err != nil {
			fmt.Println("Sendblockchain RPC call err, ", err)
		}
		fmt.Printf("Did other side receive it?: %s\n", reply)
	}
}

/*********************************
RPC calls for Artnodes to inkMiner
*********************************/
func listenToArtnode(ipPort string) {
	mRPC := new(MinerRPC)
	server := rpc.NewServer()
	registerServer(server, mRPC)
	// Listen for incoming tcp packets on specified port.
	l, e := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%s", artAppListenPort))
	if e != nil {
		log.Fatal("listen error:", e)
	}

	go server.Accept(l)
	runtime.Gosched()
}

func (m *MinerRPC) Connect(minerprivatekey string, reply *ValidMiner) error {
	var v ValidMiner
	// fmt.Println(getPrivKeyInStr(myPrivKey))
	// fmt.Println(minerprivatekey)

	if myKeyPairInString == minerprivatekey {
		v = ValidMiner{MinerNetSets: settings, Valid: true}
		fmt.Println("validKey:", minerprivatekey)
		*reply = v
		return nil
	}
	*reply = ValidMiner{Valid: false}
	fmt.Println("valafads")
	return InvalidMinerPKError(minerprivatekey)
}

func (m *MinerRPC) GetInk(minerprivatekey string, reply *uint32) error {

	if myKeyPairInString == minerprivatekey {
		remainInk := minerInkRemain()
		fmt.Println("@@@GetInk")
		*reply = remainInk
		return nil
	}
	return InvalidMinerPKError(minerprivatekey)
}

func minerInkRemain() uint32 {
	if len(blockChain) == 0 {
		return 0
	}
	lastOne := len(blockChain) - 1
	remainInk := blockChain[lastOne].MinerInks[globalPubKeyStr]
	return remainInk.inkRemain
}

// TODO:
func (m *MinerRPC) AddShape(args AddShapeStruct, reply *AddShapeReply) error {
	// try add this shape return shape/block hash, remained ink
	svgStr := "<path d=\"" + args.ShapeSvgString + "\" stroke=\"" +
		args.Stroke + "\" fill=\"" + args.Fill + "\"/>"

	remainInk := int(minerInkRemain())
	lastBlockIndex := len(blockChain) - 1
	lastBlk := blockChain[lastBlockIndex]
	previousMap := lastBlk.CanvasInks
	fmt.Println("@@@ADDDD1", args.ShapeSvgString)
	spentInk, err := SvgHelper.AddShapeToMap(args.ShapeSvgString, args.ArtNodePK, args.Fill,
		remainInk, previousMap)

	currentInkRemain := remainInk - spentInk
	// fmt.Println("@@@ink remaining!!!! %d-----------spend!!!! %d-------currentinkRemain %d-------", remainInk, spentInk, currentInkRemain)
	if err != nil {
		return err
	}

	pkStr := getPubKeyInStr(myPrivKey.PublicKey)
	shapeHash := computeNonceSecretHash(svgStr, pkStr) // use miner's public key
	newOp := Operation{svgStr, shapeHash, args.ArtNodePK, args.ShapeSvgString, args.Fill}

	lastOne := len(blockChain) - 1
	var newBlock Block
	var err1 error
	if len(blockChain) == 0 {
		newBlock, err1 = generateFirstBlock()
		lastOne = 0
		return InsufficientInkError(spentInk)
	}
	newBlock, err1 = generateBlock(blockChain[lastOne])
	preHash, _ := calculateHash(blockChain[lastOne], settings.PoWDifficultyOpBlock)

	newOps := blockChain[lastOne].Ops
	newOps = append(newOps, newOp)
	mInks := blockChain[lastOne].MinerInks
	incAcc := mInks[globalPubKeyStr]
	fmt.Println("@@@ADD23DD")

	_, inkMined := totalInkSpentAndMinedByMiner(blockChain, pkStr)
	incAcc.inkMined = inkMined
	incAcc.inkSpent = uint32(spentInk) + incAcc.inkSpent
	incAcc.inkRemain = inkMined - incAcc.inkSpent
	fmt.Printf("@@@in incAcc inkMined!!!! %d----------- total inkSpent!!!! %d------Inkspendthistime %d-------incAcc.inkRemain %d-------", inkMined, incAcc.inkSpent, spentInk, incAcc.inkRemain)
	mInks[globalPubKeyStr] = incAcc

	canvOps := blockChain[lastOne].CanvasOperations
	myOps := canvOps[globalPubKeyStr]
	svgAndHash := svgStr + ":" + shapeHash
	myOps = append(myOps, svgAndHash)
	canvOps[globalPubKeyStr] = myOps
	newBlock = Block{preHash, 0, newOps, false, globalPubKeyStr, lastOne + 1, mInks,
		previousMap, canvOps} // need update CanvasInks
	blockHash, nonce := calculateHash(newBlock, settings.PoWDifficultyOpBlock)
	tmp, _ := strconv.ParseUint(nonce, 10, 32)
	newBlock.Nonce = uint32(tmp)
	blockChain = append(blockChain, newBlock)
	fmt.Println("@@@ADD3DD")

	for {
		last := len(blockChain) - 1
		if last > lastOne+int(args.ValidateNum) {
			break
		}
		time.Sleep(3 * time.Second)
	}
	*reply = AddShapeReply{shapeHash, blockHash, uint32(currentInkRemain)}
	return err1
}

func (m *MinerRPC) GetSvgString(shapeHash string, svgString *string) error {
	lastOne := len(blockChain) - 1
	operations := blockChain[lastOne].Ops
	for i := 0; i < len(operations); i++ {
		if operations[i].OpSig == shapeHash {
			*svgString = operations[i].AppShape // svgString
			return nil
		}
	}
	fmt.Println("@@@ GetSvgString fail")
	return InvalidShapeHashError(shapeHash)
}

func (m *MinerRPC) DeleteShape(args DelShapeArgs, inkRemaining *uint32) error {
	// try delete shape by args
	lastOne := len(blockChain) - 1
	if lastOne < 0 {
		return InvalidShapeHashError(args.shapeHash)
	}
	fmt.Print(args, "##KKK(99999)6KK")
	operations := blockChain[lastOne].Ops
	for i := 0; i < len(operations); i++ {
		fmt.Print(args.shapeHash, "##KKK6666666KK", operations[i].OpSig)
		if operations[i].OpSig == args.shapeHash {
			if args.ArtNodePK == operations[i].PubKeyArtNode {

				fmt.Print("##KKKKKKK")
				newOp := Operation{"delete", args.shapeHash, args.ArtNodePK, "", ""}
				newBlock, _ := generateBlock(blockChain[lastOne])
				var noOp uint8
				if blockChain[lastOne].NoOpBlock {
					noOp = settings.PoWDifficultyNoOpBlock
				} else {
					noOp = settings.PoWDifficultyOpBlock
				}
				preHash, _ := calculateHash(blockChain[lastOne], noOp)

				newOps := blockChain[lastOne].Ops
				newOps = append(newOps, newOp)

				lastBlk := blockChain[lastOne]
				mInks := lastBlk.MinerInks
				incAcc := mInks[globalPubKeyStr]
				previousMap := lastBlk.CanvasInks

				returnedInk, err2 := SvgHelper.RemoveShapeFromMap(operations[i].ShapeCommand, args.ArtNodePK,
					operations[i].ShapeFill, previousMap)

				incAcc.inkRemain = incAcc.inkRemain + uint32(returnedInk)
				fmt.Println("@@@ADD23DD")

				incAcc.inkSpent = incAcc.inkSpent - uint32(returnedInk)

				mInks[globalPubKeyStr] = incAcc

				canvOps := blockChain[lastOne].CanvasOperations
				myOps := canvOps[globalPubKeyStr]
				svgAndHash := "delete:" + args.shapeHash
				myOps = append(myOps, svgAndHash)
				canvOps[globalPubKeyStr] = myOps
				newBlock = Block{preHash, 0, newOps, false, globalPubKeyStr, lastOne + 1, mInks,
					previousMap, canvOps} // need update CanvasInks globalPubKeyStr
				_, nonce := calculateHash(newBlock, settings.PoWDifficultyOpBlock)
				tmp, _ := strconv.ParseUint(nonce, 10, 32)
				newBlock.Nonce = uint32(tmp)
				blockChain = append(blockChain, newBlock)

				for {
					last := len(blockChain) - 1
					if last > lastOne+int(args.validateNum) {
						break
					}
					time.Sleep(3 * time.Second)
				}
				ink := blockChain[lastOne].MinerInks[globalPubKeyStr]
				*inkRemaining = ink.inkRemain
				return err2
			}
			return ShapeOwnerError(args.shapeHash)
		}
	}

	fmt.Println("@@@ DeleteShape")
	return InvalidShapeHashError(args.shapeHash)
}

func (m *MinerRPC) GetShapes(blockHash string, shapeHashes *[]string) error {
	// get shapeHashes
	fmt.Println("@@@ GetShapes")
	lastOne := len(blockChain) - 1
	if lastOne < 0 {
		return InvalidBlockHashError(blockHash)
	}
	var noOp uint8
	if blockChain[lastOne].NoOpBlock {
		noOp = settings.PoWDifficultyNoOpBlock
	} else {
		noOp = settings.PoWDifficultyOpBlock
	}
	lastblockHash, _ := calculateHash(blockChain[lastOne], noOp)
	if lastblockHash == blockHash {
		ops := blockChain[lastOne].Ops
		for j := 0; j < len(ops); j++ {
			(*shapeHashes)[j] = ops[j].OpSig
		}
		return nil
	}
	for i := len(blockChain) - 1; i >= 0; i-- {
		if blockChain[i].PrevHash == blockHash {
			ops := blockChain[i-1].Ops
			for j := 0; j < len(ops); j++ {
				(*shapeHashes)[j] = ops[j].OpSig
			}
			return nil
		}
	}
	return InvalidBlockHashError(blockHash)
}

func (m *MinerRPC) GetGenesisBlock(args int, blockHash *string) error {
	*blockHash = settings.GenesisBlockHash
	return nil
}

func (m *MinerRPC) GetChildren(blockHash string, blockHashes *[]string) error {
	// blockHashes = children of blockHash

	lastOne := len(blockChain) - 1
	if lastOne < 0 {
		return InvalidBlockHashError(blockHash)
	}
	fmt.Println("@@@ GetChildren")
	var strs []string
	for i := len(blockChain) - 1; i >= 0; i-- {
		if blockChain[i].PrevHash == blockHash {
			var noOp uint8
			if blockChain[i].NoOpBlock {
				noOp = settings.PoWDifficultyNoOpBlock
			} else {
				noOp = settings.PoWDifficultyOpBlock
			}
			mblockHash, _ := calculateHash(blockChain[i], noOp)
			strs = append(strs, mblockHash)
			*blockHashes = strs
			return nil
		}
	}
	return InvalidBlockHashError(blockHash)
}

func (m *MinerRPC) CloseCanvas(args int, reply *CloseCanvReply) error {

	lastOne := len(blockChain) - 1
	if lastOne < 0 {
		*reply = CloseCanvReply{inkRemaining: 0}
		return nil
	}
	ink := blockChain[lastOne].MinerInks[myKeyPairInString]

	*reply = CloseCanvReply{blockChain[lastOne].Ops, ink.inkRemain}
	fmt.Println("@@@ CloseCanvas")
	return nil
}

/*********************************
RPC calls for inkMIner to inkMiner
*********************************/
func (m *MinerToMinerRPC) PrintText(textToPrint string, reply *string) error {
	fmt.Println("Inside PrintText")
	fmt.Println(textToPrint)
	*reply = "We printed the text you requested"
	return nil
}

func (m *MinerToMinerRPC) EstablishReverseRPC(addr string, reply *string) error {
	minersConnectedTo.Lock()
	defer minersConnectedTo.Unlock()
	for _, currentNeighbour := range minersConnectedTo.all {
		if addr == currentNeighbour {
			*reply = "Already connected to this miner"
			return nil
		}
	}
	addrTCP, e := net.ResolveTCPAddr("tcp", addr)
	if e != nil {
		fmt.Println("Error resolving address in EstablishReverseRPC")
	}
	go connectToMiner(addrTCP)
	*reply = "Successfully established reverse connection"
	return nil
}

func (m *MinerToMinerRPC) SendBlockchain(bc *[]Block, reply *string) error {
	// 1. Check if the sent block is longer than our block.
	fmt.Println("Inside sbc")
	if isSentChainLonger(*bc) {
		fmt.Println("sbc: Received a longer chain than what we have.")
		// 1.2 If the sent block <bc> is longer, validate that it is a good block chain
		if validateSufficientInkAll(*bc) && validateBlockChain(*bc) {
			// 2.2 Otherwise acquire the lock for global blockchain and set it to sent block
			fmt.Println("sbc: longer chain is valid, we'll throw ours away")
			blockChain = *bc
			*reply = strconv.FormatBool(true)
			return nil
		}
		// 2.1 If the longer sent block <bc> is bad, silently return
	}
	// 1.1 If the sent block <bc> is not longer, silently return
	*reply = strconv.FormatBool(false)
	return nil
}

func registerServer(server *rpc.Server, s MinerRPCs) {
	// registers interface by name of `MyServer`.
	server.RegisterName("InkMinerRPC", s)
}

func registerServerMinerToMiner(server *rpc.Server, s Miner2MinerRPCs) {
	// registers interface by name of `MyServer`.
	server.RegisterName("MinerToMinerRPC", s)
}

// if helperGetNodes(ipPort, myMinerInfo, &miners) {
// 	fmt.Println("HelperGetNodes returned true")
// 	fmt.Println(miners)
// 	minersConnectedTo.Lock()
// 	defer minersConnectedTo.Unlock()
// 	// Item is a net.Addr in the array "miners"
// 	for _, item := range miners {
// 		shouldWeConnect := true
// 		for _, existingNeighbour := range minersConnectedTo.all {
// 			if item.String() == existingNeighbour {
// 				shouldWeConnect = false
// 				break
// 			}
// 		}
// 		if shouldWeConnect {
// 			connectToMiner(item)
// 			minersConnectedTo.all = append(minersConnectedTo.all, item.String())
// 		}
// 	}
// }

func getPrivKeyInStr(privKey ecdsa.PrivateKey) string {
	privateKeyBytes, _ := x509.MarshalECPrivateKey(&privKey)
	privKeyInString := hex.EncodeToString(privateKeyBytes)
	return privKeyInString
}

func getPubKeyInStr(pubKey ecdsa.PublicKey) string {
	str := fmt.Sprintf("%s%s", pubKey.X, pubKey.Y)
	return str
}

func listenForIncomingConnections(port int) {
	gob.Register(&net.TCPAddr{})
	minerToMinerRPC := new(MinerToMinerRPC)

	server := rpc.NewServer()
	registerServerMinerToMiner(server, minerToMinerRPC)

	l, e := net.Listen("tcp", fmt.Sprintf("%s:%d", localIPPortArr[0], port))
	if e != nil {
		exitOnError("Error listening in for incoming connection requests", e)
	}

	for {
		conn, _ := l.Accept()
		fmt.Println("Received a connection request")
		go server.ServeConn(conn)
	}
}

// func callAddShapeHelper() {
// 	AddShapeToMap()
// }

// func callRemoveShapeHelper() {
// 	RemoveShapeFromMap()
// }

/*********************************
Operation Validation
*********************************/

// Traverse the given block chain and returns a list of all miners in the block
func minersInBlockChain(bc []Block) []string {
	var miners []string
	for _, blk := range bc {
		if !contains(miners, blk.PubKeyMiner) {
			miners = append(miners, blk.PubKeyMiner)
		}
	}
	return miners
}

func contains(miners []string, miner string) bool {
	for _, m := range miners {
		if miner == m {
			return true
		}
	}
	return false
}

// Calculates the ink cost of an operation
func shapeInkCost(shapeSVG string) uint32 {
	return 30
}

// For a given block, calculates ink cost to commit the operations in the block
func costOfOperations(ops []Operation) uint32 {
	var sum uint32
	sum = 0
	for _, op := range ops {
		sum += shapeInkCost(op.AppShape)
	}

	return sum
}

// Given a block chain and miner, tallies the total amount of ink
// mined and total ink spent and returns them, respectively
// IMPORTANT: the current function traverses the entire block chain
//            and tallies total spent and mined including the current block
//            A different function will calculate whether the current operations
//            to commit into the existing block chain can be done with the
//            ink quantity pre-new-block-generation
func totalInkSpentAndMinedByMiner(bc []Block, miner string) (inkSpent, inkMined uint32) {
	inkMined = 0
	inkSpent = 0

	for _, blk := range bc {
		if miner == blk.PubKeyMiner {
			// Increment InkMined
			if blk.NoOpBlock {
				inkMined += settings.InkPerNoOpBlock
			} else {
				inkMined += settings.InkPerOpBlock
			}

			inkSpent += costOfOperations(blk.Ops)
		}
	}

	return inkSpent, inkMined
}

// Given a blockChain, validates that the miner (identified by public key)
// has sufficient ink to perform all the operations specified in the block chain
func validateSufficientInkMiner(bc []Block, key string) bool {
	// the miner is identified by their key
	inkSpent, inkMined := totalInkSpentAndMinedByMiner(bc, key)
	fmt.Println("v")
	fmt.Println(inkSpent)
	fmt.Println(inkMined)
	if inkMined >= inkSpent {
		return true
	}

	return false
}

// Given a blockChain, validates that the miner (identified by public key)
// has sufficient ink to perform all the operations specified in the block chain
func validateSufficientInkAll(bc []Block) bool {
	miners := minersInBlockChain(bc)

	for _, miner := range miners {
		// if the miner doesn't have enough ink, then the helper
		// returns false, so we negate to enter the block and return false overall
		if !validateSufficientInkMiner(bc, miner) {
			return false
		}
	}
	return true
}

// Checks if the ink-miner has enough ink to commit the current set of
// operations given the ink that they have (without counting the ink from
// the current block that they are generating.
func haveEnoughInkToCommitOperations(ops []Operation, b Block, miner string) bool {
	cost := costOfOperations(ops)
	if cost > b.MinerInks[miner].inkRemain {
		return false
	}

	return true
}

// TODO: the canvas operations field stores miner -> svg:shapeHash/op-sig mappings
// Given a block and a shapeHash, checks if shapeHash matches any operation signatures
// in the block.
func identicalShapeOnCanvas(b Block, shapeHash string) bool {
	// 1. Obtain map of canvas operations
	cOps := b.CanvasOperations
	// 2. Iterate through every ink-miner in the map
	for _, minerCanvasOps := range cOps {
		// 3. For each ink-miner, determine whether the set of operations on canvas contains
		//    the supplied shapeHash (which is the shape we wish to add)
		for _, svgOpSig := range minerCanvasOps {
			pair := strings.Split(svgOpSig, ":")
			opSig := pair[0]
			if shapeHash == opSig {
				return true
			}
		}
	}
	return false
}

// TODO: the canvas operations field stores miner -> svg:shapeHash/op-sig mappings
// Verifies that the existing shapeHash belongs on canvas to the owner
func shapeExistsAndOwnedByMiner(b Block, miner string, shapeHash string) bool {
	// 1. Obtain map of canvas operations
	cOps := b.CanvasOperations
	// 2. Obtain list of operations (array of op-sigs/shape hashes)
	//    of the specified miner.
	var minerCanvasOps []string
	for k, v := range cOps {
		// miner pub key and list of op-sigs
		if k == miner {
			minerCanvasOps = v
			break
		}
	}
	// 4. Iterate through the array and return true if the shapeHash matches one
	for _, op := range minerCanvasOps {
		if op == shapeHash {
			return true
		}
	}
	return false
}

/*********************************
Block & Blockchain Validation
*********************************/

// Given a block, determines whether the PrevHash has the requisite
// zeros and that the nonce proof-of-work was correctly performed
func validateBlockHashNonce(b Block) (bool, string) {
	var difficulty uint8
	// 1. Determine whether we have a OP or NO-OP block
	if b.NoOpBlock {
		difficulty = settings.PoWDifficultyNoOpBlock
	} else {
		difficulty = settings.PoWDifficultyOpBlock
	}
	// 1. If block is 2nd block and above, determine if PrevHash
	//    has requisite number of zeros
	if b.Index > 1 {
		if !hasNZeros(b.PrevHash, difficulty) {
			return false, ""
		}
	}

	currHash, n := calculateHash(b, difficulty)

	val := (n == strconv.FormatUint(uint64(b.Nonce), 10))

	return val, currHash
}

// Given a block, determines whether each of the operation signatures
// are valid given the block's ink-miner public key
func validateBlockOpSigs(b Block) bool {
	minerKey := b.PubKeyMiner

	// Iterate through operations array
	for _, op := range b.Ops {
		// Verify that our calculated operation signature
		// matches the supplied operation signature
		// TODO: change minerKey to myKeyPairInString (which is a global variable)
		ourOpSig := computeNonceSecretHash(minerKey, op.AppShape)
		if !(ourOpSig == op.OpSig) {
			return false
		}
	}

	return true
}

// Traverses the given block chain, and determines its overall validity.
// Validity is composed of 3 components:
//      (1) Block points to a previous legal block
//      (2) Block has correct nonce proof-of-work
//      (3) Block has correct operation signatures
func validateBlockChain(bc []Block) bool {
	var hashVal string
	var boolValidNonce bool
	var boolValidOpSig bool

	for _, b := range bc {
		if b.Index > 1 {
			if !(hashVal == b.PrevHash) {
				return false
			}
		}

		boolValidNonce, hashVal = validateBlockHashNonce(b)
		boolValidOpSig = validateBlockOpSigs(b)

		if !boolValidNonce || !boolValidOpSig {
			return false
		}
	}

	return true
}
