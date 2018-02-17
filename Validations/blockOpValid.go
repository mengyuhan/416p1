package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

/*********************************
Block & Blockchain Validation
*********************************/

var (
	settings = MinerSettings{
		InkPerNoOpBlock:        10,
		InkPerOpBlock:          50,
		PoWDifficultyOpBlock:   3,
		PoWDifficultyNoOpBlock: 3,
		GenesisBlockHash:       "83218ac34c1834c26781fe4bde918ee4",
	}
)

type MinerSettings struct {
	InkPerOpBlock          uint32
	InkPerNoOpBlock        uint32
	PoWDifficultyOpBlock   uint8
	PoWDifficultyNoOpBlock uint8
	GenesisBlockHash       string
}

type Operation struct {
	AppShape      string
	OpSig         string
	PubKeyArtNode string // key of the art node that generated the op
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
	CanvasInks       map[Coordinate]PixelState
	CanvasOperations map[string][]string // Ink Miner to List of Operations (op-sigs) on canvas
}

// Test 3: validateBlockChain
func main() {
	bc := initializeGoodBlockChain()
	fmt.Println("Good")
	fmt.Println(validateBlockChain(bc))

	bc2 := initializeBadBlockChain()
	fmt.Println("Bad")
	fmt.Println(validateBlockChain(bc2))
}

/*********************************
BLOCK VALIDATION FUNCTIONS
*********************************/

// TODO:
func generateBlock(oldBlock Block) (Block, error) {
	var newBlock Block

	return newBlock, nil
}

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

/*********************************
HELPER FUNCTIONS
*********************************/

func blkToString(b Block) string {
	return b.PrevHash + convertOpToString(b.Ops) + b.PubKeyMiner + string(b.Index)
}

// [prev-hash, op, op-signature, pub-key, nonce, other data structures]
func calculateHash(b Block, powDifficulty uint8) (hash, nonce string) {
	// TODO: Include other data structures
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

/*********************************
BLOCK INITIALIZATION
*********************************/

func initializeTestBlockA() Block {
	opsArr := make([]Operation, 1)
	opsArr[0] = Operation{AppShape: "circle", OpSig: "ddcbd137454ff4b1631e14bd2ef9788e", PubKeyArtNode: "artApp1"}
	mInks := make(map[string]InkAccount)
	mInks["miner1"] = InkAccount{inkMined: 50, inkRemain: 20, inkSpent: 30}
	cInks := make(map[Coordinate]PixelState)
	cOps := make(map[string][]string)
	cOps["miner1"] = []string{"circle"}

	a := Block{
		PrevHash:         "83218ac34c1834c26781fe4bde918ee4",
		Nonce:            2541,
		Ops:              opsArr,
		NoOpBlock:        false,
		PubKeyMiner:      "miner1",
		Index:            1,
		MinerInks:        mInks,
		CanvasInks:       cInks,
		CanvasOperations: cOps,
	}

	return a
}

func initializeTestBlockB() Block {
	opsArr := make([]Operation, 1)
	opsArr[0] = Operation{AppShape: "polygon", OpSig: "56512e507ffefac1ed9984185df0b25e", PubKeyArtNode: "artApp2"}
	mInks := make(map[string]InkAccount)
	mInks["miner1"] = InkAccount{inkMined: 50, inkRemain: 20, inkSpent: 30}
	mInks["miner2"] = InkAccount{inkMined: 50, inkRemain: 20, inkSpent: 30}
	cInks := make(map[Coordinate]PixelState)
	cOps := make(map[string][]string)
	cOps["miner1"] = []string{"circle"}
	cOps["miner1"] = []string{"polygon"}

	a := Block{
		PrevHash:         "1425f824236435682003c5b3360bb000",
		Nonce:            3397,
		Ops:              opsArr,
		NoOpBlock:        false,
		PubKeyMiner:      "miner2",
		Index:            2,
		MinerInks:        mInks,
		CanvasInks:       cInks,
		CanvasOperations: cOps,
	}

	return a
}

func initializeTestBlockC() Block {
	opsArr := make([]Operation, 1)
	opsArr[0] = Operation{AppShape: "line", OpSig: "op3-sig", PubKeyArtNode: "artApp3"}
	mInks := make(map[string]InkAccount)
	mInks["miner1"] = InkAccount{inkMined: 50, inkRemain: 20, inkSpent: 30}
	mInks["miner2"] = InkAccount{inkMined: 50, inkRemain: 20, inkSpent: 30}
	mInks["miner3"] = InkAccount{inkMined: 50, inkRemain: 20, inkSpent: 30}
	cInks := make(map[Coordinate]PixelState)
	cOps := make(map[string][]string)
	cOps["miner1"] = []string{"circle"}
	cOps["miner2"] = []string{"polygon"}
	cOps["miner3"] = []string{"line"}

	a := Block{
		PrevHash:         "0b8756d1385817104692689a68a60000",
		Nonce:            2781,
		Ops:              opsArr,
		NoOpBlock:        false,
		PubKeyMiner:      "miner3",
		Index:            3,
		MinerInks:        mInks,
		CanvasInks:       cInks,
		CanvasOperations: cOps,
	}

	return a
}

func initializeTestBlockBadBlock() Block {
	opsArr := make([]Operation, 3)
	opsArr[0] = Operation{AppShape: "line", OpSig: "op3-sig", PubKeyArtNode: "artApp3"}
	mInks := make(map[string]InkAccount)
	mInks["miner1"] = InkAccount{inkMined: 50, inkRemain: 20, inkSpent: 30}
	mInks["miner2"] = InkAccount{inkMined: 50, inkRemain: 20, inkSpent: 30}
	mInks["miner3"] = InkAccount{inkMined: 50, inkRemain: 20, inkSpent: 30}
	cInks := make(map[Coordinate]PixelState)
	cOps := make(map[string][]string)
	cOps["miner1"] = []string{"circle"}
	cOps["miner2"] = []string{"polygon"}
	cOps["miner3"] = []string{"line"}

	a := Block{
		PrevHash:         "4ec987b07c082e4a602b39fb28d27000",
		Nonce:            123,
		Ops:              opsArr,
		NoOpBlock:        false,
		PubKeyMiner:      "miner3",
		Index:            3,
		MinerInks:        mInks,
		CanvasInks:       cInks,
		CanvasOperations: cOps,
	}

	return a
}

func initializeGoodBlockChain() []Block {
	b := make([]Block, 2)
	b[0] = initializeTestBlockA()
	b[1] = initializeTestBlockB()

	return b
}

func initializeBadBlockChain() []Block {
	b := make([]Block, 3)
	b[0] = initializeTestBlockA()
	b[1] = initializeTestBlockB()
	b[2] = initializeTestBlockC()

	return b
}
