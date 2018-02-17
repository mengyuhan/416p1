package block

type Operation struct {
	AppShape      string
	OpSig         string
	PubKeyArtNode string //key of the art node that generated the op
}

type Block struct {
	PrevHash    string // MD5 hash with 0s
	Nonce       uint32
	Ops         []Operation
	PubKeyMiner string
	Index       int
	MyInk       int
	MinerInks   map[string]int
}

/*
TODO: Functions on blocks
*/
