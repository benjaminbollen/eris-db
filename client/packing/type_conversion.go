// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package packing

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	gethAbi "github.com/ethereum/go-ethereum/accounts/abi"
)

// func convertToInferedType(inputType *gethAbi.Type, argument interface{}) (interface{}, error) {
// 	// if (inputType.IsSlice || inputType.IsArray) &&
// 	// 	!(inputType.T == gethAbi.BytesTy || inputType.T == gethAbi.FixedBytesTy || inputType.T == gethAbi.FunctionTy ) 

// 	inferedType, err := inferType(inputType)

// 	return nil, nil
// }

// castToCloserType tries to recast golang type to golang type that is
// closer to the desired ABI type, errors if casting fails or conversion not defined.
func convertToCloserType(inputType *gethAbi.Type, argument interface{}) (interface{}, error) {
	if isNestedType(inputType) {
		fmt.Printf("MARMOT \n Ty %v " +
			" IsSlice %v IsArray %v \n" +
			" SliceSize %v Kind %v\n\n", inputType.T, inputType.IsSlice, inputType.IsArray, inputType.SliceSize, inputType.Kind)
		fmt.Printf("MARMOT ELEM\n Ty %v " +
			" IsSlice %v IsArray %v \n" +
			" SliceSize %v Kind %v\n\n", inputType.Elem.T, inputType.Elem.IsSlice, inputType.Elem.IsArray, inputType.Elem.SliceSize, inputType.Elem.Kind)
				
		// // NOTE: multidimensional arrays are not correctly supported by go-ethereum/abi
		// // so we need to make a more explicit check directly on the ABI definition file.
		// if isNestedType(inputType.Elem) {
		// 	return nil, fmt.Errorf("burrow-client currently does not support packing bytes for multi-dimensional arrays or slices.")
		// }

		var effectiveLength int = -1
		// if IsArray length is set by function signature;
		// otherwise leave indeterminate and set it with
		// provided argument length later
		if inputType.IsArray {
			// if Array, SliceSize is set for Type
			effectiveLength = inputType.SliceSize
		}

		switch reflect.TypeOf(argument).Kind() {
		// require argument to be a slice
		case reflect.Slice:
			s := reflect.ValueOf(argument)

			// if signature accepts variable length, set it
			// it to number of provided elements
			if effectiveLength == -1 {
				effectiveLength = s.Len()
			} else if effectiveLength != s.Len() {
				return nil, fmt.Errorf("Error expected length of array (%v) " +
					"does not match elements provided (%v).", effectiveLength, s)
			}

			inferedType, err := inferType(inputType)
			if err != nil {
				return nil, fmt.Errorf("Failed to infer type for %s: %s", inputType, err)
			}
			fmt.Printf("MARMOT INFERED TYPE: %s\n ", inferedType)
			arrayCloserTypes := reflect.New(reflect.SliceOf(inferedType))
			fmt.Printf("MARMOT INFERED SLICE: %s\n ", arrayCloserTypes)

			for i := 0; i < effectiveLength; i++ {
				// Value interface for ith element
				// t := s.Index(i)
				// convert value t to
			// 	switch inferedType.(type) {
			// 	case int8, int16, int32, int64:
			// 		return convertToInt(argument, inputType.Size)
			// 	case uint8, uint16, uint32, uint64:
			// 		return convertToUint(argument, inputType.Size)
			// 	case gethAbi.BoolTy:
			// 		return convertToBool(argument)
			// 	case gethAbi.StringTy:
			// 		return convertToString(argument)
			// 	// case gethAbi.SliceTy:
			// 	case gethAbi.AddressTy:
			// 		return convertToAddress(argument)
			// 	case gethAbi.FixedBytesTy:
			// 		// NOTE: if FixedBytesTy && varSize != 0 => SliceSize = varSize
			// 		return convertToFixedBytes(argument, inputType.SliceSize)
			// 	// case gethAbi.BytesTy:
			// 		// currently do not support 
			// 		// return converToSlice(argument)
			// 		// case gethAbi.HashTy:
			// 		// case gethAbi.FixedpointTy:
			// 		// case gethAbi.FunctionTy:
			// 		// default:
			// 	}

				// t := 2
				// z := interface{}(t)
				// arrayCloserTypes[i] = z
			}
			fmt.Printf("MARMOT KIND %v \n\n", inputType.Type)

			fmt.Printf("MARMOT LENGTH %v; len(s) %v; array %s \n\n", effectiveLength, s.Len(), arrayCloserTypes)
			return arrayCloserTypes, nil

		}

		fmt.Printf("MARMOT JSON %s\n", argument)
		return argument, nil

		// return nil, fmt.Errorf("MARMOT ERROR\n\n")
	}


	switch inputType.T {
	case gethAbi.IntTy:
		return convertToInt(argument, inputType.Size)
	case gethAbi.UintTy:
		return convertToUint(argument, inputType.Size)
	case gethAbi.BoolTy:
		return convertToBool(argument)
	case gethAbi.StringTy:
		return convertToString(argument)
	// case gethAbi.SliceTy:
	case gethAbi.AddressTy:
		return convertToAddress(argument)
	case gethAbi.FixedBytesTy:
		// NOTE: if FixedBytesTy && varSize != 0 => SliceSize = varSize
		return convertToFixedBytes(argument, inputType.SliceSize)
	// case gethAbi.BytesTy:
		// currently do not support 
		// return converToSlice(argument)
		// case gethAbi.HashTy:
		// case gethAbi.FixedpointTy:
		// case gethAbi.FunctionTy:
		// default:
	}
	return nil, nil
}

// func convertToInferedType(argument interface{})


// convertToInt is idempotent for int; for other types
// it tries to convert the value to var sized int, or fails
func convertToInt(argument interface{}, size int) (interface{}, error) {
	switch t := argument.(type) {
	case int, int8, int16, int32, int64:
		y, ok := t.(int64)
		if !ok {
			return nil, fmt.Errorf("Failed to assert intX as int64")
		}
		// TODO: not tested this works, or makes sense; currently un-used code path
		return reduceToVarSizeInt(y, size)
	case uint: // ignore uintptr for now
		// avoid overrunning
		if t <= math.MaxInt64 {
			return reduceToVarSizeInt(int64(t), size)
		} else {
			return nil, fmt.Errorf("Failed to convert uint to int: bigger than max int64")
		}
	case uint8:
		return reduceToVarSizeInt(int64(t), size)
	case uint16:
		return reduceToVarSizeInt(int64(t), size)
	case uint32:
		return reduceToVarSizeInt(int64(t), size)
	case uint64:
		if t <= math.MaxInt64 {
			return reduceToVarSizeInt(int64(t), size)
		} else {
			return nil, fmt.Errorf("Failed to convert uint to int: bigger than max int64")
		}
	case float32:
		var y int64
		// float can be significantly bigger than int
		if math.Abs(float64(t)) <= math.MaxInt64 {
			y = int64(t)
			if float32(y) == t {
				return reduceToVarSizeInt(y, size)
			} else {
				return nil, fmt.Errorf("Failed to convert float32 to int: non-integer value")
			}
		} else {
			return nil, fmt.Errorf("Failed to convert float32 to int: bigger than max int64")
		}
	case float64:
		var y int64
		// float can be significantly bigger than int
		if math.Abs(t) <= math.MaxInt64 {
			y = int64(t)
			if float64(y) == t {
				return reduceToVarSizeInt(y, size)
			} else {
				return nil, fmt.Errorf("Failed to convert float64 to int: non-integer value")
			}
		} else {
			return nil, fmt.Errorf("Failed to convert float64 to int: bigger than max int64")
		}
	case string:
		y, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			return nil, err
		}
		return reduceToVarSizeInt(y, size)
	case bool:
		if t {
			return reduceToVarSizeInt(int64(1), size)
		} else {
			return reduceToVarSizeInt(int64(0), size)
		}
	case complex64, complex128:
		return nil, fmt.Errorf("Failed to convert complex type to int")
	default:
		return nil, fmt.Errorf("Failed to convert unhandled type to int")
	}
}

// this is logic we do not want to keep but it solves the problem
// up to 64bit int for now; definitely in need of better solution
func reduceToVarSizeInt(integer int64, size int) (interface{}, error) {
	// for now map to golang type sizes + big.Int for 256bits
	switch size {
	case 8:
		var x int8
		x = int8(integer)
		if int64(x) == integer {
			return x, nil
		} else {
			return nil, fmt.Errorf("Failed to reduce int64 to int8: overflow")
		}
	case 16:
		var x int16
		x = int16(integer)
		if int64(x) == integer {
			return x, nil
		} else {
			return nil, fmt.Errorf("Failed to reduce int64 to int16: overflow")
		}
	case 32:
		var x int32
		x = int32(integer)
		if int64(x) == integer {
			return x, nil
		} else {
			return nil, fmt.Errorf("Failed to reduce int64 to int32: overflow")
		}
	case 64:
		return integer, nil
	case 256:
		i := new(big.Int)
		i.SetInt64(integer)
		return i, nil
	case -1:
		return nil, fmt.Errorf("Failed to reduce int64: size undefined")
	default:
		return nil, fmt.Errorf("Failed to reduce int64: size %v unhandled", size)
	}
}

// convertToUint is idempotent for uint; for other types
// it tries to convert the value to uint64, or fails
func convertToUint(argument interface{}, size int) (interface{}, error) {
	switch t := argument.(type) {
	case int:
		if t >= 0 {
			return reduceToVarSizeUint(uint64(t), size)
		} else {
			return nil, fmt.Errorf("Failed to convert int to uint: strictly negative")
		}
	case int8:
		if t >= 0 {
			return reduceToVarSizeUint(uint64(t), size)
		} else {
			return nil, fmt.Errorf("Failed to convert int to uint: strictly negative")
		}
	case int16:
		if t >= 0 {
			return reduceToVarSizeUint(uint64(t), size)
		} else {
			return nil, fmt.Errorf("Failed to convert int to uint: strictly negative")
		}
	case int32:
		if t >= 0 {
			return reduceToVarSizeUint(uint64(t), size)
		} else {
			return nil, fmt.Errorf("Failed to convert int to uint: strictly negative")
		}
	case int64:
		if t >= 0 {
			return reduceToVarSizeUint(uint64(t), size)
		} else {
			return nil, fmt.Errorf("Failed to convert int to uint: strictly negative")
		}
	case uint, uint8, uint16, uint32, uint64: // ignore uintptr for now
		y, ok := t.(uint64)
		if !ok {
			return nil, fmt.Errorf("Failed to assert uintX as uint64")
		}
		return reduceToVarSizeUint(y, size)
	case float32:
		var y uint64
		// float can be significantly bigger than int
		if math.Abs(float64(t)) <= math.MaxUint64 && t >= 0 {
			y = uint64(t)
			if float32(y) == t {
				return reduceToVarSizeUint(uint64(y), size)
			} else {
				return nil, fmt.Errorf("Failed to convert float32 to uint: non-integer value")
			}
		} else {
			return nil, fmt.Errorf("Failed to convert float32 to uint: bigger than max uint64 or negative")
		}
	case float64:
		var y uint64
		// float can be significantly bigger than int
		if math.Abs(t) <= math.MaxUint64 && t >= 0 {
			y = uint64(t)
			if float64(y) == t {
				return reduceToVarSizeUint(uint64(y), size)
			} else {
				return nil, fmt.Errorf("Failed to convert float64 to uint: non-integer value")
			}
		} else {
			return nil, fmt.Errorf("Failed to convert float64 to uint: bigger than max uint64 or negative")
		}
	case string:
		y, err := strconv.ParseUint(t, 10, 64)
		if err != nil {
			return nil, err
		}
		return reduceToVarSizeUint(y, size)
	case bool:
		if t {
			return reduceToVarSizeUint(uint64(1), size)
		} else {
			return reduceToVarSizeUint(uint64(0), size)
		}
	case complex64, complex128:
		return nil, fmt.Errorf("Failed to convert complex type to uint")
	default:
		return nil, fmt.Errorf("Failed to convert unhandled type to uint")
	}
}

// this is logic we do not want to keep but it solves the problem
// up to 64bit uint for now; definitely in need of better solution
func reduceToVarSizeUint(integer uint64, size int) (interface{}, error) {
	// for now map to golang type sizes + big.Int for 256bits
	switch size {
	case 8:
		var x uint8
		x = uint8(integer)
		if uint64(x) == integer {
			return x, nil
		} else {
			return nil, fmt.Errorf("Failed to reduce uint64 to uint8: overflow")
		}
	case 16:
		var x uint16
		x = uint16(integer)
		if uint64(x) == integer {
			return x, nil
		} else {
			return nil, fmt.Errorf("Failed to reduce uint64 to uint16: overflow")
		}
	case 32:
		var x uint32
		x = uint32(integer)
		if uint64(x) == integer {
			return x, nil
		} else {
			return nil, fmt.Errorf("Failed to reduce uint64 to uint32: overflow")
		}
	case 64:
		return integer, nil
	case 256:
		i := new(big.Int)
		i.SetUint64(integer)
		return i, nil
	case -1:
		return nil, fmt.Errorf("Failed to reduce uint64: size undefined")
	default:
		return nil, fmt.Errorf("Failed to reduce uint64: size %v unhandled", size)
	}
}

// convertToBool is idempotent for type bool, and for int, uint,
// float it converts 0 to false and 1 to true; and for string
// it converts the string "0" and case-insensitive "false" to false,
// and "1" and case-insensitive "true" to true. For other types
// it will fail. 
func convertToBool(argument interface{}) (interface{}, error) {
	switch t := argument.(type) {
	case bool:
		return argument, nil

	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		if t == 0 {
			return false, nil
		} else if t == 1 {
			return true, nil
		} else {
			return nil, fmt.Errorf("Failed to convert (U)Int* to Bool, value not mappable")
		}

	case float32, float64:
		if t == 0 {
			return false, nil
		} else if t == 1 {
			return true, nil
		} else {
			return nil, fmt.Errorf("Failed to convert Float* type to Bool")
		}

	case string:
		if t == "0" || strings.ToLower(t) == "false" {
			return false, nil
		} else if t == "1" || strings.ToLower(t) == "true" {
			return true, nil
		} else {
			return nil, fmt.Errorf("Failed to convert String type to Bool")
		}

	case complex64, complex128:
		return nil, fmt.Errorf("Failed to convert complex type to Bool")
	default:
		return nil, fmt.Errorf("Failed to convert unhandled type to string")
	}
}

// convertToString is idempotent for string; for other types it fails
// can be extended for other type
func convertToString(argument interface{}) (interface{}, error) {
	switch argument.(type) {
	case string:
		return argument, nil
	default:
		return nil, fmt.Errorf("Failed to convert unhandled type to string")
	}
}

func convertToFixedBytes(argument interface{}, size int) (interface{}, error) {
	switch t := argument.(type) {
	case []byte:
		if len(t) > size {
			return nil, fmt.Errorf("Failed to convert bytes to Fixed length bytes %v: overflow", size)
		}

		padded := rightPadBytes(t, size)
		return padded, nil
	case string:

		decoded, err := hex.DecodeString(t)
		if err != nil {
			return nil, err
		}

		if len(decoded) > size {
			return nil, fmt.Errorf("Failed to convert string to Fixed length bytes %v: overflow", size)
		}

		// Right pad into proper sized array
		padded := rightPadBytes(decoded, size)

		return padded, nil
	default:
		return nil, fmt.Errorf("Failed to convert unhandled type to bytes")
	}
}

// This is not working... Apparently its not the right type
func convertToAddress(argument interface{}) (interface{}, error) {
	switch t := argument.(type) {
	case []byte:
		if len(t) != 20 {
			return nil, fmt.Errorf("Failed to convert bytes to address: bad Length %v", len(t))
		}

		var padded [20]byte
		copy(padded[:], t)

		return padded, nil
	case string:

		decoded, err := hex.DecodeString(t)
		if err != nil {
			return nil, err
		}

		if len(decoded) != 20 {
			return nil, fmt.Errorf("Failed to convert string to address: bad Length %v", len(decoded))
		}

		var padded [20]byte
		copy(padded[:], decoded)

		return padded, nil
	default:
		return nil, fmt.Errorf("Failed to convert unhandled type to address")
	}
}

// func converToSlice(argument interface{}) (interface{}, error) {
// 	switch t := argument.(type) {
// 	case string:

// 	default:
// 		fmt.Printf("MARMOT: default")
// 		return nil, fmt.Errorf("Failed to convert unhandled type to slice")
// 	}
// }

// Ben, Sorry for putting these here, I couldn't find an existing function in our codebase that did this
// If one exists, It should be an easy change.

func rightPadBytes(inBytes []byte, size int) []byte {
	padded := make([]byte, size)
	copy(padded[0:len(inBytes)], inBytes)
	return padded
}

func leftPadBytes(inBytes []byte, size int) []byte {
	padded := make([]byte, size)
	copy(padded[size-len(inBytes):], inBytes)
	return padded
}
