package main

import "fmt"

/*********************************
Operation Validation
*********************************/

var (
	settings = MinerSettings{InkPerNoOpBlock: 10, InkPerOpBlock: 50}
)

type MinerSettings struct {
	InkPerOpBlock   uint32
	InkPerNoOpBlock uint32
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
	for _, minerOpSigs := range cOps {
		// 3. For each ink-miner, determine whether the set of operations on canvas contains
		//    the supplied shapeHash (which is the shape we wish to add)
		for _, opSig := range minerOpSigs {
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

// Test 7: validateSufficientInkMinerAll
func main() {
	bcTrue := initializeblockChainValidInkMinerAll1()
	bcFalse := initializeblockChainValidInkMinerAll2()
	val := validateSufficientInkAll(bcTrue)
	fmt.Println("Expected true")
	fmt.Println(val)

	val2 := validateSufficientInkAll(bcFalse)
	fmt.Println("Expected true")
	fmt.Println(val2)
}

// TEST FUNCTIONS

func initializeTestBlockA() Block {
	opsArr := make([]Operation, 3)
	opsArr[0] = Operation{AppShape: "circle", OpSig: "op1-sig", PubKeyArtNode: "artApp1"}
	opsArr[1] = Operation{AppShape: "box", OpSig: "op2-sig", PubKeyArtNode: "artApp1"}
	opsArr[2] = Operation{AppShape: "line", OpSig: "op3-sig", PubKeyArtNode: "artApp2"}
	mInks := make(map[string]InkAccount)
	mInks["miner1"] = InkAccount{inkMined: 100, inkRemain: 20, inkSpent: 30}
	cInks := make(map[Coordinate]PixelState)
	cOps := make(map[string][]string)
	cOps["miner1"] = []string{"op1-sig", "op2-sig", "op3-sig"}

	a := Block{
		PrevHash:         "genesis",
		Nonce:            32,
		Ops:              opsArr,
		NoOpBlock:        false,
		PubKeyMiner:      "miner1",
		Index:            0,
		MinerInks:        mInks,
		CanvasInks:       cInks,
		CanvasOperations: cOps,
	}

	return a
}

func initializeTestBlockB() Block {
	opsArr := make([]Operation, 3)
	opsArr[0] = Operation{AppShape: "circle", OpSig: "opa-sig", PubKeyArtNode: "artApp3"}
	opsArr[1] = Operation{AppShape: "box", OpSig: "opB-sig", PubKeyArtNode: "artApp3"}
	opsArr[2] = Operation{AppShape: "line", OpSig: "opC-sig", PubKeyArtNode: "artApp4"}
	mInks := make(map[string]InkAccount)
	mInks["miner1"] = InkAccount{inkMined: 100, inkRemain: 20, inkSpent: 80}
	mInks["miner2"] = InkAccount{inkMined: 50, inkRemain: 20, inkSpent: 30}
	cInks := make(map[Coordinate]PixelState)
	cOps := make(map[string][]string)
	cOps["miner1"] = []string{"op1-sig", "op2-sig", "op3-sig"}
	cOps["miner2"] = []string{"opA-sig", "opB-sig", "opC-sig"}

	b := Block{
		PrevHash:         "block1",
		Nonce:            32,
		Ops:              opsArr,
		NoOpBlock:        false,
		PubKeyMiner:      "miner2",
		Index:            0,
		MinerInks:        mInks,
		CanvasInks:       cInks,
		CanvasOperations: cOps,
	}

	return b
}

func initializeTestBlockC() Block {
	opsArr := make([]Operation, 3)
	opsArr[0] = Operation{AppShape: "circle", OpSig: "opa-sig", PubKeyArtNode: "artApp3"}
	opsArr[1] = Operation{AppShape: "box", OpSig: "opB-sig", PubKeyArtNode: "artApp3"}
	opsArr[2] = Operation{AppShape: "line", OpSig: "opC-sig", PubKeyArtNode: "artApp4"}
	mInks := make(map[string]InkAccount)
	mInks["miner1"] = InkAccount{inkMined: 50, inkRemain: 20, inkSpent: 30}
	mInks["miner2"] = InkAccount{inkMined: 50, inkRemain: 20, inkSpent: 30}
	mInks["miner3"] = InkAccount{inkMined: 50, inkRemain: 20, inkSpent: 30}
	cInks := make(map[Coordinate]PixelState)
	cOps := make(map[string][]string)
	cOps["miner1"] = []string{"op1-sig", "op2-sig", "op3-sig"}
	cOps["miner2"] = []string{"opA-sig", "opB-sig", "opC-sig"}
	cOps["miner3"] = []string{"op4-sig", "op5-sig", "op6-sig"}

	c := Block{
		PrevHash:         "block2",
		Nonce:            32,
		Ops:              opsArr,
		NoOpBlock:        true,
		PubKeyMiner:      "miner3",
		Index:            0,
		MinerInks:        mInks,
		CanvasInks:       cInks,
		CanvasOperations: cOps,
	}

	return c
}

// Use this to test the true case for haveEnoughInkToCommitOperations func
func initializeTestBlockD() Block {
	opsArr := make([]Operation, 3)
	// For test #6, comment out all 3 operations; otherwise uncomment
	// opsArr[0] = Operation{AppShape: "circle", OpSig: "opa-sig", PubKeyArtNode: "artApp3"}
	// opsArr[1] = Operation{AppShape: "box", OpSig: "opB-sig", PubKeyArtNode: "artApp3"}
	// opsArr[2] = Operation{AppShape: "line", OpSig: "opC-sig", PubKeyArtNode: "artApp4"}
	mInks := make(map[string]InkAccount)
	mInks["miner1"] = InkAccount{inkMined: 50, inkRemain: 20, inkSpent: 30}
	mInks["miner2"] = InkAccount{inkMined: 50, inkRemain: 20, inkSpent: 30}
	mInks["miner3"] = InkAccount{inkMined: 1000, inkRemain: 970, inkSpent: 30}
	cInks := make(map[Coordinate]PixelState)
	cOps := make(map[string][]string)
	cOps["miner1"] = []string{"op1-sig", "op2-sig", "op3-sig"}
	cOps["miner2"] = []string{"opA-sig", "opB-sig", "opC-sig"}
	cOps["miner3"] = []string{"op4-sig", "op5-sig", "op6-sig"}

	c := Block{
		PrevHash:         "block2",
		Nonce:            32,
		Ops:              opsArr,
		NoOpBlock:        false,
		PubKeyMiner:      "miner3",
		Index:            0,
		MinerInks:        mInks,
		CanvasInks:       cInks,
		CanvasOperations: cOps,
	}

	return c
}

// Use this to test the true case for haveEnoughInkToCommitOperations func
func initializeTestBlockE() Block {
	opsArr := make([]Operation, 0)
	mInks := make(map[string]InkAccount)
	mInks["miner1"] = InkAccount{inkMined: 50, inkRemain: 20, inkSpent: 30}
	mInks["miner2"] = InkAccount{inkMined: 50, inkRemain: 20, inkSpent: 30}
	mInks["miner3"] = InkAccount{inkMined: 1000, inkRemain: 970, inkSpent: 30}
	cInks := make(map[Coordinate]PixelState)
	cOps := make(map[string][]string)
	cOps["miner1"] = []string{"op1-sig", "op2-sig", "op3-sig"}
	cOps["miner2"] = []string{"opA-sig", "opB-sig", "opC-sig"}
	cOps["miner3"] = []string{"op4-sig", "op5-sig", "op6-sig"}

	c := Block{
		PrevHash:         "block2",
		Nonce:            32,
		Ops:              opsArr,
		NoOpBlock:        false,
		PubKeyMiner:      "miner4",
		Index:            0,
		MinerInks:        mInks,
		CanvasInks:       cInks,
		CanvasOperations: cOps,
	}

	return c
}

func initializeblockChain() []Block {
	bc := make([]Block, 3)
	bc[0] = initializeTestBlockA()
	bc[1] = initializeTestBlockB()
	bc[2] = initializeTestBlockC()

	return bc
}

func initializeblockChainTotalInkSpent() []Block {
	bc := make([]Block, 4)
	bc[0] = initializeTestBlockA()
	bc[1] = initializeTestBlockB()
	bc[2] = initializeTestBlockC()
	bc[3] = initializeTestBlockD()

	return bc
}

func initializeblockChainValidInkMiner() []Block {
	bc := make([]Block, 5)
	bc[0] = initializeTestBlockA()
	bc[1] = initializeTestBlockB()
	bc[2] = initializeTestBlockC()
	bc[3] = initializeTestBlockD()
	bc[4] = initializeTestBlockE()

	return bc
}

func initializeblockChainValidInkMinerAll1() []Block {
	bc := make([]Block, 2)
	bc[0] = initializeTestBlockX()
	bc[1] = initializeTestBlockY()

	return bc
}

func initializeblockChainValidInkMinerAll2() []Block {
	bc := make([]Block, 3)
	bc[0] = initializeTestBlockX()
	bc[1] = initializeTestBlockY()
	bc[2] = initializeTestBlockZ()

	return bc
}

func initializeTestBlockX() Block {
	opsArr := make([]Operation, 1)
	opsArr[0] = Operation{AppShape: "circle", OpSig: "opa-sig", PubKeyArtNode: "artApp1"}
	mInks := make(map[string]InkAccount)
	mInks["miner1"] = InkAccount{inkMined: 50, inkRemain: 20, inkSpent: 30}
	cInks := make(map[Coordinate]PixelState)
	cOps := make(map[string][]string)
	cOps["miner1"] = []string{"opa-sig"}

	c := Block{
		PrevHash:         "genesis",
		Nonce:            32,
		Ops:              opsArr,
		NoOpBlock:        false,
		PubKeyMiner:      "miner1",
		Index:            0,
		MinerInks:        mInks,
		CanvasInks:       cInks,
		CanvasOperations: cOps,
	}

	return c
}

func initializeTestBlockY() Block {
	opsArr := make([]Operation, 1)
	opsArr[0] = Operation{AppShape: "circle", OpSig: "opb-sig", PubKeyArtNode: "artApp2"}
	mInks := make(map[string]InkAccount)
	mInks["miner2"] = InkAccount{inkMined: 50, inkRemain: 20, inkSpent: 30}
	cInks := make(map[Coordinate]PixelState)
	cOps := make(map[string][]string)
	cOps["miner1"] = []string{"opa-sig"}
	cOps["miner2"] = []string{"opb-sig"}

	c := Block{
		PrevHash:         "prev-hash",
		Nonce:            32,
		Ops:              opsArr,
		NoOpBlock:        false,
		PubKeyMiner:      "miner2",
		Index:            0,
		MinerInks:        mInks,
		CanvasInks:       cInks,
		CanvasOperations: cOps,
	}

	return c
}

func initializeTestBlockZ() Block {
	opsArr := make([]Operation, 3)
	opsArr[0] = Operation{AppShape: "circle", OpSig: "opc-sig", PubKeyArtNode: "artApp2"}
	opsArr[1] = Operation{AppShape: "circle", OpSig: "opd-sig", PubKeyArtNode: "artApp2"}
	opsArr[2] = Operation{AppShape: "circle", OpSig: "ope-sig", PubKeyArtNode: "artApp2"}
	mInks := make(map[string]InkAccount)
	mInks["miner2"] = InkAccount{inkMined: 50, inkRemain: 20, inkSpent: 30}
	cInks := make(map[Coordinate]PixelState)
	cOps := make(map[string][]string)
	cOps["miner1"] = []string{"opa-sig"}
	cOps["miner2"] = []string{"opb-sig", "opc-sig", "opd-sig", "ope-sig"}

	c := Block{
		PrevHash:         "prev-hash",
		Nonce:            32,
		Ops:              opsArr,
		NoOpBlock:        false,
		PubKeyMiner:      "miner2",
		Index:            0,
		MinerInks:        mInks,
		CanvasInks:       cInks,
		CanvasOperations: cOps,
	}

	return c
}
