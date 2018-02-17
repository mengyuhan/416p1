package SvgHelper

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
)

// var mapPoints map[string]MapPoint

type MapPoint struct {
	count     int
	publicKey string
}

type point struct {
	x int
	y int
}

// Contains amount of ink remaining.
type InsufficientInkError uint32

func (e InsufficientInkError) Error() string {
	return fmt.Sprintf("BlockArt: Not enough ink to addShape [%d]", uint32(e))
}

// Empty
type OutOfBoundsError struct{}

func (e OutOfBoundsError) Error() string {
	return fmt.Sprintf("BlockArt: Shape is outside the bounds of the canvas")
}

// Contains the hash of the shape that this shape overlaps with.
type ShapeOverlapError string

func (e ShapeOverlapError) Error() string {
	return fmt.Sprintf("BlockArt: Shape overlaps with a previously added shape [%s]", string(e))
}

// Contains the bad shape hash string.
type ShapeOwnerError string

func (e ShapeOwnerError) Error() string {
	return fmt.Sprintf("BlockArt: Shape owned by someone else [%s]", string(e))
}

// Contains the offending svg string.
type InvalidShapeSvgStringError string

func (e InvalidShapeSvgStringError) Error() string {
	return fmt.Sprintf("BlockArt: Bad shape svg string [%s]", string(e))
}

// Contains the offending svg string.
type ShapeSvgStringTooLongError string

func (e ShapeSvgStringTooLongError) Error() string {
	return fmt.Sprintf("BlockArt: Shape svg string too long [%s]", string(e))
}

//------------------------------------------------------------------------------------------------
// add shape to map struct mapPoints
// args:
// - svgString : passed from client
// - shapType : fill or transparent
// - minerInk : currrent ink miner has
/////////////////
// Can return the following errors:
// - DisconnectedError
// - ShapeOverlapError
// - ShapeOwnerError
// - OutofBoundError: if any point is outside canvas size, return error
// - InsufficientInkError: if given minerInk is less then ink needed
// - InvalidShapeSvgStringError: if given filled type with not closed shape
func AddShapeToMap(svgString string, publicKey string, shapeType string, minerInk int, mapPoints map[string]MapPoint) (ink int, err error) {
	var transparentMapPoints map[int]point
	var polygon [][]bool
	var close bool
	if shapeType == "transparent" {
		transparentMapPoints, ink, _, err = TransparentSvgToCoord(svgString, publicKey, minerInk)
		if err != nil {
			fmt.Println(err)
			return 0, err
		}
		//check overlap
		for _, value := range transparentMapPoints {
			overlap := checkOverlap(value, publicKey, mapPoints)
			if overlap {
				pString := strconv.Itoa(value.x) + "," + strconv.Itoa(value.y)
				err = ShapeOverlapError(pString)
				fmt.Println(err)
				return 0, err
			}
		}
		// if no overlap add all points in map
		for _, value := range transparentMapPoints {
			addPoint(value, publicKey, mapPoints)
		}
		// filled
	} else {
		transparentMapPoints, _, close, err = TransparentSvgToCoord(svgString, publicKey, minerInk)
		if !close {
			err = InvalidShapeSvgStringError(svgString)
			fmt.Println(err)
			return 0, err
		}
		polygon, ink, err = FilledSvgToPolygon(transparentMapPoints, publicKey, minerInk)
		if err != nil {
			fmt.Println(err)
			return 0, err
		}
		//check overlap
		for j := range polygon {
			for k := range polygon[j] {
				if polygon[j][k] {
					value := point{x: k, y: j}
					overlap := checkOverlap(value, publicKey, mapPoints)
					if overlap {
						pString := strconv.Itoa(value.x) + "," + strconv.Itoa(value.y)
						err = ShapeOverlapError(pString)
						fmt.Println(err)
						return 0, err
					}
				}

			}
		}
		// if no overlap add all points in map
		for j := range polygon {
			for k := range polygon[j] {
				if polygon[j][k] {
					value := point{x: k, y: j}
					addPoint(value, publicKey, mapPoints)
				}
			}
		}
	}
	// for key, value := range mapPoints {
	// 	fmt.Printf("in glogal mapPoints, key %s, count %d, publickey %s\n", key, value.count, value.publicKey)
	// }
	return ink, nil
}

// remove shape from map struct mapPoints, return ink returned
// args:
// - svgString : passed from client
// - shapType : fill or transparent
/////////////////
// Can return the following errors:
// - DisconnectedError
// - ShapeOwnerError
// - OutofBoundError: if any point is outside canvas size, return error
// - InvalidShapeSvgStringError: if given filled type with not closed shape
func RemoveShapeFromMap(svgString string, publicKey string, shapeType string, mapPoints map[string]MapPoint) (ink int, err error) {
	var transparentMapPoints map[int]point
	var polygon [][]bool
	var close bool
	//var close bool
	if shapeType == "transparent" {
		transparentMapPoints, ink, _, err = RemoveTransparentSvgToCoord(svgString, publicKey)
		if err != nil {
			fmt.Println(err)
			return 0, err
		}
		//check have all points to remove
		for _, value := range transparentMapPoints {
			err = havePoint(value, publicKey, mapPoints)
			if err != nil {
				fmt.Println(err)
				return 0, err
			}
		}
		//if have all points, remove points from global map
		for _, value := range transparentMapPoints {
			removePoint(value, publicKey, mapPoints)
		}
		// filled
	} else {
		transparentMapPoints, _, close, err = RemoveTransparentSvgToCoord(svgString, publicKey)
		if !close {
			err = InvalidShapeSvgStringError(svgString)
			fmt.Println(err)
			return 0, err
		}
		polygon, ink, err = RemoveFilledSvgToPolygon(transparentMapPoints, publicKey)
		if err != nil {
			fmt.Println(err)
			return 0, err
		}
		//check have all points to remove
		for j := range polygon {
			for k := range polygon[j] {
				if polygon[j][k] {
					value := point{x: k, y: j}
					err = havePoint(value, publicKey, mapPoints)
					if err != nil {
						fmt.Println(err)
						return 0, err
					}
				}
			}
		}
		//if have all points, remove points from global map
		for j := range polygon {
			for k := range polygon[j] {
				if polygon[j][k] {
					value := point{x: k, y: j}
					removePoint(value, publicKey, mapPoints)
				}
			}
		}
	}
	// for key, value := range mapPoints {
	// 	fmt.Printf("in glogal mapPoints, key %s, count %d\n", key, value.count)
	// }
	return ink, nil
}

// this helper function convert from svg string to coordinates in the map
// return :
// localMapPoints map[int]point: Map of Points of this svg
// ink int: needed to draw such svgpath
// close bool: if svg is closed or not
// err error:
// 1: OutofBoundError: if any point is outside canvas size, return error
// 2: InsufficientInkError: if given minerInk is less then ink needed
func TransparentSvgToCoord(svgString string, publicKey string, minerInk int) (localMapPoints map[int]point, ink int, close bool, err error) {
	initialPoint := point{x: 0, y: 0}
	endPoint := point{x: 0, y: 0}
	currentPoint := point{x: 0, y: 0}
	var temPoint point
	var p point
	var j int
	var s3 string
	var points []point
	ink = 0
	close = false
	localMapPoints = make(map[int]point)
	index := 0
	i := 0
	for i < len(svgString) {
		s := string(svgString[i : i+1])
		//println(s)
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
				// println("x1")
				// println(initialPoint.x)
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
					initialPoint.y = num
					currentPoint.y = num
				} else {
					initialPoint.y = num + initialPoint.y
					currentPoint.y = num + currentPoint.y
				}
				if !checkCanvasSize(currentPoint) {
					err = OutOfBoundsError{}
					return localMapPoints, ink, close, err
				}
				// println("y1")
				// println(initialPoint.y)
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
				// println("x3")
				// println(temPoint.x)
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
				// println("y3")
				// println(temPoint.y)
				s3 = ""
			}
			if !checkCanvasSize(temPoint) {
				err = OutOfBoundsError{}
				return localMapPoints, ink, close, err
			}
			// detect conflict and put points into map
			//fmt.Printf("in L, temPoint.x %d,temPoint.y %d,currentPoint.x %d,currentPoint.y %d\n", temPoint.x, temPoint.y, currentPoint.x, currentPoint.y)
			points = getPointsFromVertex(currentPoint.x, temPoint.x, currentPoint.y, temPoint.y)
			for j, p = range points {
				localMapPoints[index+j] = p
			}
			index = index + j
			// for key, value := range localMapPoints {
			// 	fmt.Printf("in L's localMapPoints, key %d, pstring %s, index %d\n", key, value, index)
			// }
			currentPoint = temPoint
			endPoint = temPoint
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
				// println("x3")
				// println(temPoint.x)
				s3 = ""
			}
			if !checkCanvasSize(temPoint) {
				err = OutOfBoundsError{}
				return localMapPoints, ink, close, err
			}
			// detect conflict and put points into map
			//fmt.Printf("in H, temPoint.x %d,temPoint.y %d,currentPoint.x %d,currentPoint.y %d\n", temPoint.x, temPoint.y, currentPoint.x, currentPoint.y)
			points = getPointsFromVertex(currentPoint.x, temPoint.x, currentPoint.y, temPoint.y)
			for j, p = range points {
				localMapPoints[index+j] = p
			}
			index = index + j
			// for key, value := range localMapPoints {
			// 	fmt.Printf("in H's localMapPoints, key %d, pstring %s, index %d\n", key, value, index)
			// }
			currentPoint = temPoint
			endPoint = temPoint
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
				// println("y3")
				// println(num)
				s3 = ""
			}
			if !checkCanvasSize(temPoint) {
				err = OutOfBoundsError{}
				return localMapPoints, ink, close, err
			}
			// detect conflict and put points into map
			//fmt.Printf("in V, temPoint.x %d,temPoint.y %d,currentPoint.x %d,currentPoint.y %d\n", temPoint.x, temPoint.y, currentPoint.x, currentPoint.y)
			points = getPointsFromVertex(currentPoint.x, temPoint.x, currentPoint.y, temPoint.y)
			for j, p = range points {
				localMapPoints[index+j] = p
			}
			index = index + j
			// for key, value := range localMapPoints {
			// 	fmt.Printf("in V's localMapPoints, key %d, pstring %s, index %d\n", key, value, index)
			// }
			currentPoint = temPoint
			endPoint = temPoint
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
			//fmt.Printf("in Z, temPoint.x %d,temPoint.y %d,currentPoint.x %d,currentPoint.y %d\n", temPoint.x, temPoint.y, currentPoint.x, currentPoint.y)
			points = getPointsFromVertex(currentPoint.x, temPoint.x, currentPoint.y, temPoint.y)
			// remove the last point because its the same as very first vertex
			points = points[:len(points)-1]
			for j, p = range points {
				localMapPoints[index+j] = p
			}
			index = index + j
			currentPoint = temPoint
			endPoint = temPoint
		}
		i++
	}
	if endPoint == initialPoint {
		close = true
	}
	for _, _ = range localMapPoints {
		//fmt.Printf("in localMapPoints, key %d, pstring %s\n", key, value)
		ink++
	}
	fmt.Printf("close and ink--------------------------------------------------------------\n")
	println(close)
	println(ink)
	if ink > minerInk {
		ink32 := int32(ink)
		fmt.Printf("-------------------------------not enough ink---need %d----have %d-----------------\n", ink, minerInk)
		err = InsufficientInkError(ink32)
		return localMapPoints, ink, close, err
	}
	return localMapPoints, ink, close, nil
}

// this helper function convert from svg string to coordinates in the map
// return :
// localMapPoints map[int]point: Map of Points of this svg
// ink int: needed to draw such svgpath
// close bool: if svg is closed or not
// err error:
// 1: OutofBoundError: if any point is outside canvas size, return error
func RemoveTransparentSvgToCoord(svgString string, publicKey string) (localMapPoints map[int]point, ink int, close bool, err error) {
	initialPoint := point{x: 0, y: 0}
	endPoint := point{x: 0, y: 0}
	currentPoint := point{x: 0, y: 0}
	var temPoint point
	var p point
	var j int
	var s3 string
	var points []point
	ink = 0
	close = false
	localMapPoints = make(map[int]point)
	index := 0
	i := 0
	for i < len(svgString) {
		s := string(svgString[i : i+1])
		//println(s)
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
				// println("x1")
				// println(initialPoint.x)
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
					initialPoint.y = num
					currentPoint.y = num
				} else {
					initialPoint.y = num + initialPoint.y
					currentPoint.y = num + currentPoint.y
				}
				if !checkCanvasSize(currentPoint) {
					err = OutOfBoundsError{}
					return localMapPoints, ink, close, err
				}
				// println("y1")
				// println(initialPoint.y)
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
				// println("x3")
				// println(temPoint.x)
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
				// println("y3")
				// println(temPoint.y)
				s3 = ""
			}
			if !checkCanvasSize(temPoint) {
				err = OutOfBoundsError{}
				return localMapPoints, ink, close, err
			}
			// detect conflict and put points into map
			//fmt.Printf("in L, temPoint.x %d,temPoint.y %d,currentPoint.x %d,currentPoint.y %d\n", temPoint.x, temPoint.y, currentPoint.x, currentPoint.y)
			points = getPointsFromVertex(currentPoint.x, temPoint.x, currentPoint.y, temPoint.y)
			for j, p = range points {
				localMapPoints[index+j] = p
			}
			index = index + j
			// for key, value := range localMapPoints {
			// 	fmt.Printf("in L's localMapPoints, key %d, pstring %s, index %d\n", key, value, index)
			// }
			currentPoint = temPoint
			endPoint = temPoint
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
				// println("x3")
				// println(temPoint.x)
				s3 = ""
			}
			if !checkCanvasSize(temPoint) {
				err = OutOfBoundsError{}
				return localMapPoints, ink, close, err
			}
			// detect conflict and put points into map
			//fmt.Printf("in H, temPoint.x %d,temPoint.y %d,currentPoint.x %d,currentPoint.y %d\n", temPoint.x, temPoint.y, currentPoint.x, currentPoint.y)
			points = getPointsFromVertex(currentPoint.x, temPoint.x, currentPoint.y, temPoint.y)
			for j, p = range points {
				localMapPoints[index+j] = p
			}
			index = index + j
			// for key, value := range localMapPoints {
			// 	fmt.Printf("in H's localMapPoints, key %d, pstring %s, index %d\n", key, value, index)
			// }
			currentPoint = temPoint
			endPoint = temPoint
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
				// println("y3")
				// println(num)
				s3 = ""
			}
			if !checkCanvasSize(temPoint) {
				err = OutOfBoundsError{}
				return localMapPoints, ink, close, err
			}
			// detect conflict and put points into map
			//fmt.Printf("in V, temPoint.x %d,temPoint.y %d,currentPoint.x %d,currentPoint.y %d\n", temPoint.x, temPoint.y, currentPoint.x, currentPoint.y)
			points = getPointsFromVertex(currentPoint.x, temPoint.x, currentPoint.y, temPoint.y)
			for j, p = range points {
				localMapPoints[index+j] = p
			}
			index = index + j
			// for key, value := range localMapPoints {
			// 	fmt.Printf("in V's localMapPoints, key %d, pstring %s, index %d\n", key, value, index)
			// }
			currentPoint = temPoint
			endPoint = temPoint
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
			//fmt.Printf("in Z, temPoint.x %d,temPoint.y %d,currentPoint.x %d,currentPoint.y %d\n", temPoint.x, temPoint.y, currentPoint.x, currentPoint.y)
			points = getPointsFromVertex(currentPoint.x, temPoint.x, currentPoint.y, temPoint.y)
			// remove the last point because its the same as very first vertex
			points = points[:len(points)-1]
			for j, p = range points {
				localMapPoints[index+j] = p
			}
			index = index + j
			currentPoint = temPoint
			endPoint = temPoint
		}
		i++
	}
	if endPoint == initialPoint {
		close = true
	}
	for _, _ = range localMapPoints {
		//fmt.Printf("in localMapPoints, key %d, pstring %s\n", key, value)
		ink++
	}
	return localMapPoints, ink, close, nil
}

// if overlap return true, else return false
func checkOverlap(p point, publicKey string, mapPoints map[string]MapPoint) bool {
	pString := strconv.Itoa(p.x) + "," + strconv.Itoa(p.y)
	mappoint, exist := mapPoints[pString]
	if exist {
		//check public key
		if mappoint.publicKey == publicKey {
			return false
		}
		return true
	}
	return false

}

// add point (pstring) to global map mapPoints
func addPoint(p point, publicKey string, mapPoints map[string]MapPoint) {
	pString := strconv.Itoa(p.x) + "," + strconv.Itoa(p.y)
	mappoint, exist := mapPoints[pString]
	if exist {
		mappoint = MapPoint{count: mappoint.count + 1, publicKey: publicKey}
		mapPoints[pString] = mappoint
	} else {
		mappoint = MapPoint{count: 1, publicKey: publicKey}
		mapPoints[pString] = mappoint
	}
}

// if global map have such point, return true, else return false
//return ShapeOwnerError
func havePoint(p point, publicKey string, mapPoints map[string]MapPoint) error {
	pString := strconv.Itoa(p.x) + "," + strconv.Itoa(p.y)
	mappoint, exist := mapPoints[pString]
	if exist {
		//check public key
		if mappoint.publicKey == publicKey {
			return nil
		}
		return ShapeOwnerError(pString)
	}
	return ShapeOwnerError(pString)
}

// remove point (pstring) from global map mapPoints

func removePoint(p point, publicKey string, mapPoints map[string]MapPoint) {
	pString := strconv.Itoa(p.x) + "," + strconv.Itoa(p.y)
	mappoint, _ := mapPoints[pString]
	mappoint.count--
	if mappoint.count == 0 {
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
			//fmt.Printf("quantity is %d, slope is %d, y is %d ydiff is %d\n", quantity, slope, y, ydiff)
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
			//fmt.Printf("quantity is %d, slope is %d, x is %d xdiff is %d\n", quantity, slope, x, xdiff)
			points[i] = point{x + x1, y + y1}
			i++
		}

	}
	points[quantity].x = x2
	points[quantity].y = y2
	//fmt.Printf("%v\n", points)
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

// this helper function convert from transparent line of point to array of points inside polygon
// return list of Points of polygon and ink used
// - InsufficientInkError: if given minerInk is less then ink needed
func FilledSvgToPolygon(transparentMapPoints map[int]point, publicKey string, minerInk int) (polygon [][]bool, ink int, err error) {
	maxY := 0
	maxX := 0
	ink = 0
	for _, value := range transparentMapPoints {
		if value.x > maxX {
			maxX = value.x
		}
		if value.y > maxY {
			maxY = value.y
		}
	}
	//fmt.Printf("maxY %d, maxX %d\n", maxY, maxX)
	polygon = make([][]bool, maxY+1)
	for i := range polygon {
		polygon[i] = make([]bool, maxX+1)
	}
	for _, value := range transparentMapPoints {
		// fmt.Printf("x %d, y %d\n", value.x, value.y)
		polygon[value.y][value.x] = true
	}
	//printPolygon(polygon)
	// fill all points inside polygon
	for j := range polygon {
		include := false
		count := 0
		for k := range polygon[j] {
			if polygon[j][k] {
				count++
			}
		}
		if count == 1 {
			ink++
		}
		if count > 1 {
			// apply algorithm that include points between two vertex
			for k := range polygon[j] {
				if polygon[j][k] {
					include = !include
					ink++
				} else {
					if include {
						polygon[j][k] = true
						ink++
					}
				}
			}
		}
		//fmt.Printf("in polygon row %d and ink %d--------------------------------------------------------------\n", j, ink)
	}

	fmt.Printf("in polygon close and ink--------------------------------------------------------------\n")
	println(ink)
	if ink > minerInk {
		ink32 := int32(ink)
		fmt.Printf("-------------------------------not enough ink---need %d----have %d-----------------\n", ink, minerInk)
		return polygon, ink, InsufficientInkError(ink32)
	}
	//printPolygon(polygon)

	return polygon, ink, nil
}

// this helper function convert from transparent line of point to array of points inside polygon
// return list of Points of polygon and ink returned
// - InsufficientInkError: if given minerInk is less then ink needed
func RemoveFilledSvgToPolygon(transparentMapPoints map[int]point, publicKey string) (polygon [][]bool, ink int, err error) {
	maxY := 0
	maxX := 0
	ink = 0
	for _, value := range transparentMapPoints {
		if value.x > maxX {
			maxX = value.x
		}
		if value.y > maxY {
			maxY = value.y
		}
	}
	//fmt.Printf("maxY %d, maxX %d\n", maxY, maxX)
	polygon = make([][]bool, maxY+1)
	for i := range polygon {
		polygon[i] = make([]bool, maxX+1)
	}
	for _, value := range transparentMapPoints {
		// fmt.Printf("x %d, y %d\n", value.x, value.y)
		polygon[value.y][value.x] = true
	}
	//printPolygon(polygon)
	// fill all points inside polygon
	for j := range polygon {
		count := 0
		for k := range polygon[j] {
			if polygon[j][k] {
				count++
			}
		}
		if count == 1 {
			ink++
			println("count = 1 ****************************")
		}
		if count > 1 {
			// apply algorithm that include points between two vertex
			for k := range polygon[j] {
				include := false
				if polygon[j][k] {
					include = !include
					ink++
				} else {
					if include {
						polygon[j][k] = true
						ink++
					}
				}
			}
		}
	}
	//printPolygon(polygon)

	return polygon, ink, nil
}

func printPolygon(polygon [][]bool) {
	var value3 string
	for j := range polygon {
		for k := range polygon[j] {
			if polygon[j][k] {
				value3 = "O"
			} else {
				value3 = "x"
			}
			fmt.Printf("%2s", value3)
		}
		fmt.Printf("\n")
	}
	fmt.Printf("\n")
}

///////////////////////// main is used for testing  //////////////////
// func main() {
// 	mapPoints := make(map[string]MapPoint)
// 	// basic tests (for demo, x10 or x100 for all numbers)
// 	// -----------------------------------------------------------------------------
// 	// add triangle
// 	// AddShapeToMap("M 4 0 L 0 4 h 8 l -4 -4", "123", "fill", 300)
// 	// //add square
// 	// AddShapeToMap("M 9 0 l 4 0 v 4 h -4 z", "323", "fill", 300)
// 	// 	//remove triangle
// 	// 	RemoveShapeFromMap("M 4 0 L 0 4 h 8 l -4 -4", "123", "fill", 300)
// 	// 	//remove other owner`s square
// 	// 	RemoveShapeFromMap("M 9 0 l 4 0 v 4 h -4 z", "353", "fill", 300)
// 	//add Trapezoidal
// 	// AddShapeToMap("M 5 0 l 5 0 L 5 5 h -5 z", "123", "fill", 300)
// 	// AddShapeToMap("M 5 0 l 5 0 l 5 5 h -5 z", "123", "fill", 300)
// 	// // add 凹
// 	AddShapeToMap("M 5 0 l 3 0 l 0 3 h 3 v -3  h 3 v 6 h -9 z", "123", "fill", 300, mapPoints)
// 	// // add 凸
// 	AddShapeToMap("M 5 5 l 3 0 l 0 3 h 3 v 3  h -9 v -3 h 3 z", "123", "fill", 300, mapPoints)
// }
