package main

// Expects blockartlib.go to be in the ./blockartlib/ dir, relative to
// this art-app.go file
import (
	"../SvgHelper"
)

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
}

type Coordinate struct {
	x int
	y int
}

// type PixelState struct {
// 	n           int    // number of overlapping shapes on the given x-y coordinate
// 	minerPubKey string // miner who "owns" the current pixel on shared canvas
// }

type M struct {
	count     int
	publicKey string
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
	CanvasInks       map[string]M
	CanvasOperations map[string][]string // Ink Miner to List of Operations on canvas
}

func main() {

	mapPoints := make(map[string]SvgHelper.MapPoint)
	//add triangle
	// SvgHelper.AddShapeToMap("M 4 0 L 0 4 h 8 l -4 -4", "123", "fill", 300, mapPoints)
	// //add square
	// SvgHelper.AddShapeToMap("M 9 0 l 4 0 v 4 h -4 z", "323", "fill", 300, mapPoints)
	// add 凹
	SvgHelper.AddShapeToMap("M 0 0 L 0 5", "123", "fill", 300, mapPoints)
	// remove 凹
	SvgHelper.RemoveShapeFromMap("M 5 0 l 3 0 l 0 3 h 3 v -3  h 3 v 6 h -9 z", "123", "fill", 300, mapPoints)
	// // add 凸
	SvgHelper.AddShapeToMap("M 5 5 l 3 0 l 0 3 h 3 v 3  h -9 v -3 h 3 z", "143", "fill", 300, mapPoints)
}
