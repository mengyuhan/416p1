// Operation validation tests

// // Test 1: shapeExistsAndOwnedByMiner GOOD
// func main() {
// 	b := initializeTestBlock()

// 	val := shapeExistsAndOwnedByMiner(b, "miner1", "op1-sig")
// 	fmt.Println("Expect true")
// 	fmt.Println(val)
// 	val = shapeExistsAndOwnedByMiner(b, "miner1", "op5-sig")
// 	fmt.Println("Expect true")
// 	fmt.Println(val)
// 	val = shapeExistsAndOwnedByMiner(b, "miner2", "op5-sig")
// 	fmt.Println("Not sure what to expect")
// 	fmt.Println(val)
// }

// // Test 2: identicalShapeOnCanvas GOOD
// func main() {
// 	b := initializeTestBlock()

// 	val := identicalShapeOnCanvas(b, "op1-sig")
// 	fmt.Println("Expect true")
// 	fmt.Println(val)
// 	val = identicalShapeOnCanvas(b, "op9-sig")
// 	fmt.Println("Expect true")
// 	fmt.Println(val)
// 	val = identicalShapeOnCanvas(b, "opE-sig")
// 	fmt.Println("Expect false")
// 	fmt.Println(val)
// }

// Test 3: haveEnoughInkToCommitOperations GOOD
// Indirectly tests costOfOperations & shakeInkCost
func main() {
	//settings := MinerSettings{InkPerNoOpBlock: 10, InkPerOpBlock: 50}
	blkA := initializeTestBlockA()
	blkD := initializeTestBlockD()

	inkCost := shapeInkCost("svg")
	fmt.Println("Expected cost: 30")
	fmt.Printf("Actual cost: %d", inkCost)

	val1 := haveEnoughInkToCommitOperations(blkA.Ops, blkA, "miner1")
	fmt.Println("Expect false")
	fmt.Println(val1)
	val2 := haveEnoughInkToCommitOperations(blkD.Ops, blkD, "miner3")
	fmt.Println("Expect true")
	fmt.Println(val2)
}

// Test 4: minersInBlockChain() & contains()
func main() {
	//settings := MinerSettings{InkPerNoOpBlock: 10, InkPerOpBlock: 50}
	bc := initializeblockChain()

	miners := minersInBlockChain(bc)
	fmt.Println("Miners in Block Chain")
	fmt.Println(miners)

	val1 := contains(miners, "miner1")
	fmt.Println("Expect true")
	fmt.Println(val1)
	val2 := contains(miners, "miner5")
	fmt.Println("Expect false")
	fmt.Println(val2)
}

// Test 5: totalInkSpentAndMinedByMiners
func main() {
	bc := initializeblockChainTotalInkSpent()

	inkSpent1, inkMined1 := totalInkSpentAndMinedByMiner(bc, "miner1")
	fmt.Println("Expected Spent: 90, Expected Mined: 50")
	fmt.Printf("Actual Spent1: %d\n", inkSpent1)
	fmt.Printf("Actual Mined1: %d\n", inkMined1)

	inkSpent3, inkMined3 := totalInkSpentAndMinedByMiner(bc, "miner3")
	fmt.Println("Expected Spent: 180, Expected Mined: 60")
	fmt.Printf("Actual Spent3: %d\n", inkSpent3)
	fmt.Printf("Actual Mined3: %d\n", inkMined3)
}

// Test 6: validateSufficientInkMiner
func main() {
	bc := initializeblockChainValidInkMiner()
	val := validateSufficientInkMiner(bc, "miner1")
	fmt.Println("Expected false")
	fmt.Println(val)

	val2 := validateSufficientInkMiner(bc, "miner4")
	fmt.Println("Expected true")
	fmt.Println(val2)
}

=======================================================

// Block validation tests

// Test 1: validateBlockHashNonce
func main() {
	a := initializeTestBlockA()
	val, hash := validateBlockHashNonce(a)
	fmt.Printf("Expect true: %s\n", strconv.FormatBool(val))
	fmt.Printf("Hash: %s\n", hash)

	b := initializeTestBlockBadBlock()
	val2, hash2 := validateBlockHashNonce(b)
	fmt.Printf("Expect false: %s\n", strconv.FormatBool(val2))
	fmt.Printf("Hash: %s\n", hash2)
}

// Test 2: validateOperationSignatures
func main() {
	a := initializeTestBlockA()
	val1 := validateBlockOpSigs(a)
	fmt.Printf("Expect true: %s\n", strconv.FormatBool(val1))

	b := initializeTestBlockBadBlock()
	val2 := validateBlockOpSigs(b)
	fmt.Printf("Expect false: %s\n", strconv.FormatBool(val2))
}

