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
	"math"
	"net"
	"net/rpc"
	"os"
	"regexp"
	"runtime"
	"strconv"
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
}
type MinerRPC int

type ValidMiner struct {
	MinerNetSets MinerNetSettings
	Valid        bool
}

var myKeyPairInString string

func (m *MinerRPC) Connect(privatekey string, reply *ValidMiner) error {
	// rlp := CanvasSettings{2, 3}
	var v ValidMiner
	if myKeyPairInString == privatekey {
		v = ValidMiner{Valid: true}
		fmt.Println("validKey:", privatekey)
	}
	*reply = v
	return nil
}

func registerServer(server *rpc.Server, s MinerRPCs) {
	// registers interface by name of `MyServer`.
	server.RegisterName("InkMinerRPC", s)
}

func main() {

	priv1, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	exitOnError("generate key 1", err)
	// fmt.Println(priv1)
	privateKeyBytes, _ := x509.MarshalECPrivateKey(priv1)

	myKeyPairInString = hex.EncodeToString(privateKeyBytes)
	fmt.Println("s:", myKeyPairInString, ":end")

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
	mapPoints = make(map[string]mapPoint)
	// transparentSvgToCoord("M 6 2 L 3 2", "123")
	// transparentSvgToCoord("M 5 6 L 5 3", "123")
	// transparentSvgToCoord("M 11 3 L 11 6", "123")
	// transparentSvgToCoord("M 3 11 L 6 11", "123")
	// transparentSvgToCoord("M 1 5 L 3 15", "123")
	// transparentSvgToCoord("M 1 15 L 3 5", "123")
	// transparentSvgToCoord("M 0 0 L 2 10", "123")
	// transparentSvgToCoord("M 0 10 L 2 0", "123")
	// transparentSvgToCoord("M 2 10 L 0 0", "123")
	// transparentSvgToCoord("M 2 0 L 0 10", "123")
	transparentSvgToCoord("M 0 0 L 0 2 H 4 V 4 Z", "123")
	transparentSvgToCoord("m 1 1 l 0 2 h 4 v 2 z", "123")
	transparentSvgToCoord("m 5 5 l 0 -2 h -4 v -2 z", "123")
	// transparentSvgToCoord("m 5 5 l 0 -6 h -4 v -2 z", "123")

	go server.Accept(l)
	runtime.Gosched()
	time.Sleep(10000 * time.Second)
}

var mapPoints map[string]mapPoint

type mapPoint struct {
	count     int
	publicKey string
}

type point struct {
	x int
	y int
}

func addShapeToMap(svgString string, publicKey string, transparent bool) (ink int, err error) {
	var localMapPoints map[int]string
	if transparent {
		localMapPoints, ink, err = transparentSvgToCoord(svgString, publicKey)
		if err != nil {
			println("point outside canvas")
			return 0, err
		}
		//check overlap
		for _, value := range localMapPoints {
			if checkOverlap(value, publicKey) {
				return 0, err
			}
		}
		// if no overlap add all points in map
		for _, value := range localMapPoints {
			addPoint(value, publicKey)
		}
		// filled
	} else {
		// todo
	}
	return ink, nil
}

func removeShapeFromMap(svgString string, publicKey string, transparent bool) (ink int, err error) {
	var localMapPoints map[int]string
	if transparent {
		localMapPoints, ink, err = transparentSvgToCoord(svgString, publicKey)
		if err != nil {
			println("point outside canvas")
			return 0, err
		}
		//check have all points to remove
		for _, value := range localMapPoints {
			if !havePoint(value, publicKey) {
				return 0, err
			}
		}
		for _, value := range localMapPoints {
			removePoint(value, publicKey)
		}
		// filled
	} else {
		// todo
	}
	return ink, nil
}

// this helper function convert from svg string to coordinates in the map
// return Map of Points of this svg and ink needed to draw such svgpath
// if any point is outside canvas size, return error
func transparentSvgToCoord(svgString string, publicKey string) (localMapPoints map[int]string, ink int, err error) {
	initialPoint := point{x: 0, y: 0}
	currentPoint := point{x: 0, y: 0}
	var temPoint point
	var p point
	var j int
	var s3 string
	var points []point
	localMapPoints = make(map[int]string)
	index := 0
	i := 0
	for i < len(svgString) {
		s := string(svgString[i : i+1])
		println(s)
		if s == "M" || s == "m" {
			s2 := s
			i++
			isNumber := false
			// parse string before next letter
			s = string(svgString[i : i+1])
			isNumber, _ = regexp.MatchString("[0-9]|[-]", s)
			for s == " " {
				i++
				s = string(svgString[i : i+1])
				isNumber, _ = regexp.MatchString("[0-9]|[-]", s)

			}
			// parse x
			if isNumber {
				for isNumber {
					s3 += s
					i++
					s = string(svgString[i : i+1])
					isNumber, _ = regexp.MatchString("[0-9]|[-]", s)
				}
				num, _ := strconv.Atoi(s3)
				if s2 == "M" {
					initialPoint.x = num
					currentPoint.x = num
				} else {
					initialPoint.x = num + initialPoint.x
					currentPoint.x = num + currentPoint.x
				}
				println("x1")
				println(initialPoint.x)
				s3 = ""
			}
			for s == " " {
				i++
				s = string(svgString[i : i+1])
				isNumber, _ = regexp.MatchString("[0-9]|[-]", s)
			}
			// parse y
			if isNumber {
				for isNumber {
					s3 += s
					i++
					s = string(svgString[i : i+1])
					isNumber, _ = regexp.MatchString("[0-9]|[-]", s)
				}
				num, _ := strconv.Atoi(s3)
				if s2 == "M" {
					initialPoint.x = num
					currentPoint.x = num
				} else {
					initialPoint.y = num + initialPoint.y
					currentPoint.y = num + currentPoint.y
				}
				if !checkCanvasSize(currentPoint) {
					println("current point outside canvas")
					return localMapPoints, ink, err
				}
				println("y1")
				println(initialPoint.y)
				s3 = ""
			}
		}
		if s == "L" || s == "l" {
			s2 := s
			i++
			isNumber := false
			// parse string before next letter
			s = string(svgString[i : i+1])
			isNumber, _ = regexp.MatchString("[0-9]|[-]", s)
			for s == " " && len(svgString) > i {
				i++
				s = string(svgString[i : i+1])
				isNumber, _ = regexp.MatchString("[0-9]|[-]", s)
			}
			// parse x
			if isNumber {
				for isNumber && len(svgString) > i {
					s3 += s
					i++
					s = string(svgString[i : i+1])
					isNumber, _ = regexp.MatchString("[0-9]|[-]", s)
				}
				num, _ := strconv.Atoi(s3)
				if s2 == "L" {
					temPoint.x = num
				} else {
					temPoint.x = num + currentPoint.x
				}
				println("x3")
				println(temPoint.x)
				s3 = ""
			}
			for s == " " && len(svgString) > i {
				i++
				s = string(svgString[i : i+1])
				isNumber, _ = regexp.MatchString("[0-9]|[-]", s)
			}
			// parse y
			if isNumber {
				for isNumber && len(svgString) > i {
					s3 += s
					i++
					if len(svgString) > i {
						s = string(svgString[i : i+1])
					}
					isNumber, _ = regexp.MatchString("[0-9]|[-]", s)
				}
				num, _ := strconv.Atoi(s3)
				if s2 == "L" {
					temPoint.y = num
				} else {
					temPoint.y = num + currentPoint.y
				}
				println("y3")
				println(temPoint.y)
				s3 = ""
			}
			if !checkCanvasSize(temPoint) {
				println("temPoint point outside canvas")
				return localMapPoints, ink, err
			}
			// detect conflict and put points into map
			fmt.Printf("in L, temPoint.x %d,temPoint.y %d,currentPoint.x %d,currentPoint.y %d\n", temPoint.x, temPoint.y, currentPoint.x, currentPoint.y)
			points = getPointsFromVertex(currentPoint.x, temPoint.x, currentPoint.y, temPoint.y)
			for j, p = range points {
				pString := strconv.Itoa(p.x) + "," + strconv.Itoa(p.y)
				localMapPoints[index+j] = pString
			}
			index = index + j
			for key, value := range localMapPoints {
				fmt.Printf("in L's localMapPoints, key %d, pstring %s, index %d\n", key, value, index)
			}
			currentPoint = temPoint
		}
		if s == "H" || s == "h" {
			s2 := s
			temPoint.y = currentPoint.y
			i++
			isNumber := false
			// parse string before next letter
			s = string(svgString[i : i+1])
			isNumber, _ = regexp.MatchString("[0-9]|[-]", s)
			for s == " " && len(svgString) > i {
				i++
				s = string(svgString[i : i+1])
				isNumber, _ = regexp.MatchString("[0-9]|[-]", s)
			}
			// parse x
			if isNumber {
				for isNumber && len(svgString) > i {
					s3 += s
					i++
					if len(svgString) > i {
						s = string(svgString[i : i+1])
					}
					isNumber, _ = regexp.MatchString("[0-9]|[-]", s)
				}
				num, _ := strconv.Atoi(s3)
				if s2 == "H" {
					temPoint.x = num
				} else {
					temPoint.x = num + currentPoint.x
				}
				println("x3")
				println(temPoint.x)
				s3 = ""
			}
			if !checkCanvasSize(temPoint) {
				println("temPoint point outside canvas")
				return localMapPoints, ink, err
			}
			// detect conflict and put points into map
			fmt.Printf("in H, temPoint.x %d,temPoint.y %d,currentPoint.x %d,currentPoint.y %d\n", temPoint.x, temPoint.y, currentPoint.x, currentPoint.y)
			points = getPointsFromVertex(currentPoint.x, temPoint.x, currentPoint.y, temPoint.y)
			for j, p = range points {
				pString := strconv.Itoa(p.x) + "," + strconv.Itoa(p.y)
				localMapPoints[index+j] = pString
			}
			index = index + j
			for key, value := range localMapPoints {
				fmt.Printf("in H's localMapPoints, key %d, pstring %s, index %d\n", key, value, index)
			}
			currentPoint = temPoint
		}
		if s == "V" || s == "v" {
			s2 := s
			temPoint.x = currentPoint.x
			i++
			isNumber := false
			// parse string before next letter
			s = string(svgString[i : i+1])
			isNumber, _ = regexp.MatchString("[0-9]|[-]", s)
			for s == " " && len(svgString) > i {
				i++
				s = string(svgString[i : i+1])
				isNumber, _ = regexp.MatchString("[0-9]|[-]", s)
			}
			// parse y
			if isNumber {
				for isNumber && len(svgString) > i {
					s3 += s
					i++
					if len(svgString) > i {
						s = string(svgString[i : i+1])
					}
					isNumber, _ = regexp.MatchString("[0-9]|[-]", s)
				}
				num, _ := strconv.Atoi(s3)
				if s2 == "V" {
					temPoint.y = num
				} else {
					temPoint.y = num + currentPoint.y
				}
				println("y3")
				println(num)
				s3 = ""
			}
			if !checkCanvasSize(temPoint) {
				println("temPoint point outside canvas")
				return localMapPoints, ink, err
			}
			// detect conflict and put points into map
			fmt.Printf("in V, temPoint.x %d,temPoint.y %d,currentPoint.x %d,currentPoint.y %d\n", temPoint.x, temPoint.y, currentPoint.x, currentPoint.y)
			points = getPointsFromVertex(currentPoint.x, temPoint.x, currentPoint.y, temPoint.y)
			for j, p = range points {
				pString := strconv.Itoa(p.x) + "," + strconv.Itoa(p.y)
				localMapPoints[index+j] = pString
			}
			index = index + j
			for key, value := range localMapPoints {
				fmt.Printf("in V's localMapPoints, key %d, pstring %s, index %d\n", key, value, index)
			}
			currentPoint = temPoint
		}
		if s == "Z" || s == "z" {
			temPoint.x = initialPoint.x
			temPoint.y = initialPoint.y
			i++
			// parse string before next letter
			for s == " " && len(svgString) > i {
				i++
				if len(svgString) > i {
					s = string(svgString[i : i+1])
				}
			}
			// detect conflict and put points into map
			fmt.Printf("in Z, temPoint.x %d,temPoint.y %d,currentPoint.x %d,currentPoint.y %d\n", temPoint.x, temPoint.y, currentPoint.x, currentPoint.y)
			points = getPointsFromVertex(currentPoint.x, temPoint.x, currentPoint.y, temPoint.y)
			// remove the last point because its the same as very first vertex
			points = points[:len(points)-1]
			for j, p = range points {
				pString := strconv.Itoa(p.x) + "," + strconv.Itoa(p.y)
				localMapPoints[index+j] = pString
			}
			index = index + j
			for key, value := range localMapPoints {
				fmt.Printf("in Z's localMapPoints, key %d, pstring %s, index %d\n", key, value, index)
			}
			currentPoint = temPoint
		}
		i++
	}
	return localMapPoints, ink, nil
}

// if overlap return true, else return false
func checkOverlap(pString string, publicKey string) bool {
	point, exist := mapPoints[pString]
	if exist {
		println("exist!")
		//check public key
		if point.publicKey == publicKey {
			return false
		} else {
			println("overlap!")
			return true
		}
	} else {
		return false
	}
}

// add point (pstring) to global map mapPoints
func addPoint(pString string, publicKey string) {
	point, exist := mapPoints[pString]
	if exist {
		println("exist!")
		//check public key
		point.count++
	} else {
		point := mapPoint{count: 1, publicKey: publicKey}
		mapPoints[pString] = point
	}
}

// if global map have such point, return true, else return false
func havePoint(pString string, publicKey string) bool {
	point, exist := mapPoints[pString]
	if exist {
		println("exist!")
		//check public key
		if point.publicKey == publicKey {
			return true
		} else {
			println("wrong public key!")
			return false
		}
	} else {
		return false
	}
}

// remove point (pstring) from global map mapPoints

func removePoint(pString string, publicKey string) {
	point, _ := mapPoints[pString]
	point.count--
	if point.count == 0 {
		delete(mapPoints, pString)
	}
}

//get all points between two vertexs
// return array of points
func getPointsFromVertex(x1 int, x2 int, y1 int, y2 int) []point {
	var num float64
	var slope int
	var x, y int
	if y2 == y1 {
		num = math.Abs(float64(x2 - x1))
	} else if x2 == x1 {
		num = math.Abs(float64(y2 - y1))
	} else {
		num = math.Min(math.Abs(float64(y2-y1)), math.Abs(float64(x2-x1)))
	}
	quantity := int(num)
	points := make([]point, quantity+1)
	ydiff := y2 - y1
	xdiff := x2 - x1
	if ydiff == 0 || xdiff == 0 {
		slope = 0
	} else if math.Abs(float64(ydiff)) > math.Abs(float64(xdiff)) {
		slope = ydiff / xdiff
	} else {
		slope = xdiff / ydiff
	}
	i := 0
	for i < quantity {
		if math.Abs(float64(xdiff)) < math.Abs(float64(ydiff)) {
			if slope == 0 {
				x = 0
			} else {
				x = xdiff / quantity * i
			}
			if slope == 0 {
				y = ydiff / quantity * i
			} else {
				if xdiff < 0 {
					y = -slope * i
				} else {
					y = slope * i
				}
			}
			fmt.Printf("quantity is %d, slope is %d, y is %d ydiff is %d\n", quantity, slope, y, ydiff)
			points[i] = point{x + x1, y + y1}
			i++
		} else {
			if slope == 0 {
				y = 0
			} else {
				y = ydiff / quantity * i
			}
			if slope == 0 {
				x = xdiff / quantity * i
			} else {
				if ydiff < 0 {
					x = -slope * i
				} else {
					x = slope * i
				}
			}
			fmt.Printf("quantity is %d, slope is %d, x is %d xdiff is %d\n", quantity, slope, x, xdiff)
			points[i] = point{x + x1, y + y1}
			i++
		}

	}
	points[quantity].x = x2
	points[quantity].y = y2
	fmt.Printf("%v\n", points)
	return points
}

func checkCanvasSize(temPoint point) bool {
	CanvasXMax := 1024
	CanvasYMax := 1024
	if temPoint.x > CanvasXMax || temPoint.x < 0 || temPoint.y > CanvasYMax || temPoint.y < 0 {
		return false
	}
	return true
}

// this helper function convert from filled svg string to a polygon
// return list of Points of polygon
// if any point is outside canvas size or not a closed polygon, return error
func filledSvgToPolygon(svgString string, publicKey string) (polygon []point, ink int, err error) {
	initialPoint := point{x: 0, y: 0}
	currentPoint := point{x: 0, y: 0}
	var temPoint point
	var s3 string
	i := 0
	for i < len(svgString) {
		s := string(svgString[i : i+1])
		println(s)
		if s == "M" || s == "m" {
			s2 := s
			i++
			isNumber := false
			// parse string before next letter
			s = string(svgString[i : i+1])
			isNumber, _ = regexp.MatchString("[0-9]|[-]", s)
			for s == " " {
				i++
				s = string(svgString[i : i+1])
				isNumber, _ = regexp.MatchString("[0-9]|[-]", s)

			}
			// parse x
			if isNumber {
				for isNumber {
					s3 += s
					i++
					s = string(svgString[i : i+1])
					isNumber, _ = regexp.MatchString("[0-9]|[-]", s)
				}
				num, _ := strconv.Atoi(s3)
				if s2 == "M" {
					initialPoint.x = num
					currentPoint.x = num
				} else {
					initialPoint.x = num + initialPoint.x
					currentPoint.x = num + currentPoint.x
				}
				println("x1")
				println(initialPoint.x)
				s3 = ""
			}
			for s == " " {
				i++
				s = string(svgString[i : i+1])
				isNumber, _ = regexp.MatchString("[0-9]|[-]", s)
			}
			// parse y
			if isNumber {
				for isNumber {
					s3 += s
					i++
					s = string(svgString[i : i+1])
					isNumber, _ = regexp.MatchString("[0-9]|[-]", s)
				}
				num, _ := strconv.Atoi(s3)
				if s2 == "M" {
					initialPoint.x = num
					currentPoint.x = num
				} else {
					initialPoint.y = num + initialPoint.y
					currentPoint.y = num + currentPoint.y
				}
				if !checkCanvasSize(currentPoint) {
					println("current point outside canvas")
					return polygon, ink, err
				}
				println("y1")
				println(initialPoint.y)
				s3 = ""
			}
		}
		if s == "L" || s == "l" {
			s2 := s
			i++
			isNumber := false
			// parse string before next letter
			s = string(svgString[i : i+1])
			isNumber, _ = regexp.MatchString("[0-9]|[-]", s)
			for s == " " && len(svgString) > i {
				i++
				s = string(svgString[i : i+1])
				isNumber, _ = regexp.MatchString("[0-9]|[-]", s)
			}
			// parse x
			if isNumber {
				for isNumber && len(svgString) > i {
					s3 += s
					i++
					s = string(svgString[i : i+1])
					isNumber, _ = regexp.MatchString("[0-9]|[-]", s)
				}
				num, _ := strconv.Atoi(s3)
				if s2 == "L" {
					temPoint.x = num
				} else {
					temPoint.x = num + currentPoint.x
				}
				println("x3")
				println(temPoint.x)
				s3 = ""
			}
			for s == " " && len(svgString) > i {
				i++
				s = string(svgString[i : i+1])
				isNumber, _ = regexp.MatchString("[0-9]|[-]", s)
			}
			// parse y
			if isNumber {
				for isNumber && len(svgString) > i {
					s3 += s
					i++
					if len(svgString) > i {
						s = string(svgString[i : i+1])
					}
					isNumber, _ = regexp.MatchString("[0-9]|[-]", s)
				}
				num, _ := strconv.Atoi(s3)
				if s2 == "L" {
					temPoint.y = num
				} else {
					temPoint.y = num + currentPoint.y
				}
				println("y3")
				println(temPoint.y)
				s3 = ""
			}
			if !checkCanvasSize(temPoint) {
				println("temPoint point outside canvas")
				return polygon, ink, err
			}
			// detect conflict and put points into map
			fmt.Printf("in L, temPoint.x %d,temPoint.y %d,currentPoint.x %d,currentPoint.y %d\n", temPoint.x, temPoint.y, currentPoint.x, currentPoint.y)

		}
		if s == "H" || s == "h" {
			s2 := s
			temPoint.y = currentPoint.y
			i++
			isNumber := false
			// parse string before next letter
			s = string(svgString[i : i+1])
			isNumber, _ = regexp.MatchString("[0-9]|[-]", s)
			for s == " " && len(svgString) > i {
				i++
				s = string(svgString[i : i+1])
				isNumber, _ = regexp.MatchString("[0-9]|[-]", s)
			}
			// parse x
			if isNumber {
				for isNumber && len(svgString) > i {
					s3 += s
					i++
					if len(svgString) > i {
						s = string(svgString[i : i+1])
					}
					isNumber, _ = regexp.MatchString("[0-9]|[-]", s)
				}
				num, _ := strconv.Atoi(s3)
				if s2 == "H" {
					temPoint.x = num
				} else {
					temPoint.x = num + currentPoint.x
				}
				println("x3")
				println(temPoint.x)
				s3 = ""
			}
			if !checkCanvasSize(temPoint) {
				println("temPoint point outside canvas")
				return polygon, ink, err
			}
			// detect conflict and put points into map
			fmt.Printf("in H, temPoint.x %d,temPoint.y %d,currentPoint.x %d,currentPoint.y %d\n", temPoint.x, temPoint.y, currentPoint.x, currentPoint.y)

		}
		if s == "V" || s == "v" {
			s2 := s
			temPoint.x = currentPoint.x
			i++
			isNumber := false
			// parse string before next letter
			s = string(svgString[i : i+1])
			isNumber, _ = regexp.MatchString("[0-9]|[-]", s)
			for s == " " && len(svgString) > i {
				i++
				s = string(svgString[i : i+1])
				isNumber, _ = regexp.MatchString("[0-9]|[-]", s)
			}
			// parse y
			if isNumber {
				for isNumber && len(svgString) > i {
					s3 += s
					i++
					if len(svgString) > i {
						s = string(svgString[i : i+1])
					}
					isNumber, _ = regexp.MatchString("[0-9]|[-]", s)
				}
				num, _ := strconv.Atoi(s3)
				if s2 == "V" {
					temPoint.y = num
				} else {
					temPoint.y = num + currentPoint.y
				}
				println("y3")
				println(num)
				s3 = ""
			}
			if !checkCanvasSize(temPoint) {
				println("temPoint point outside canvas")
				return polygon, ink, err
			}
			// detect conflict and put points into map
			fmt.Printf("in V, temPoint.x %d,temPoint.y %d,currentPoint.x %d,currentPoint.y %d\n", temPoint.x, temPoint.y, currentPoint.x, currentPoint.y)

		}
		if s == "Z" || s == "z" {
			temPoint.x = initialPoint.x
			temPoint.y = initialPoint.y
			i++
			// parse string before next letter
			for s == " " && len(svgString) > i {
				i++
				if len(svgString) > i {
					s = string(svgString[i : i+1])
				}
			}
			// detect conflict and put points into map
			fmt.Printf("in Z, temPoint.x %d,temPoint.y %d,currentPoint.x %d,currentPoint.y %d\n", temPoint.x, temPoint.y, currentPoint.x, currentPoint.y)

		}
		i++
	}
	return polygon, ink, nil
}
