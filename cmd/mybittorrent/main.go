package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

var (
	ErrBencodeInteger = errors.New("invalid bencode integer")
	ErrBencodeString  = errors.New("invalid bencode string")
)

type BencodeType = interface{}

type BencodeDecoded struct {
	value BencodeType
	end   int
	err   error
}

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345
func decodeBencode(bencodedString string) BencodeDecoded {
	if unicode.IsDigit(rune(bencodedString[0])) {
		return decodeString(bencodedString)
	} else if bencodedString[0] == 'i' {
		return decodeInteger(bencodedString)
	} else if bencodedString[0] == 'l' {
		return decodeList(bencodedString)
	} else if bencodedString[0] == 'd' {
		return decodeDictionary(bencodedString)
	} else {
		return BencodeDecoded{"", 0, fmt.Errorf("type not supported at the moment")}
	}
}

func decodeString(bencodedString string) BencodeDecoded {
	firstColonIndex := strings.Index(bencodedString, ":")
	if firstColonIndex == -1 {
		return BencodeDecoded{"", 0, ErrBencodeString}
	}
	lengthStr := bencodedString[:firstColonIndex]
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return BencodeDecoded{"", 0, err}
	}
	end := firstColonIndex + 1 + length
	return BencodeDecoded{bencodedString[firstColonIndex+1 : end], end, nil}
}

func decodeInteger(bencodedString string) BencodeDecoded {
	endIndex := strings.Index(bencodedString, "e")
	if endIndex == -1 {
		return BencodeDecoded{0, 0, ErrBencodeInteger}
	}
	integer, err := strconv.Atoi(bencodedString[1:endIndex])
	if err != nil {
		return BencodeDecoded{0, 0, err}
	}
	return BencodeDecoded{integer, endIndex + 1, nil}
}

func decodeList(bencodedString string) BencodeDecoded {
	list := make([]BencodeType, 0)
	length := 0
	processedBencodedString := bencodedString[1:]
	for processedBencodedString[0] != 'e' {
		bencode := decodeBencode(processedBencodedString)
		if bencode.err != nil {
			return BencodeDecoded{make([]BencodeType, 0), 0, bencode.err}
		}
		length += bencode.end
		processedBencodedString = processedBencodedString[bencode.end:]
		list = append(list, bencode.value)
	}
	end := len(bencodedString) - len(processedBencodedString) + 1
	return BencodeDecoded{list, end, nil}
}

func decodeDictionary(bencodedString string) BencodeDecoded {
	dict := map[interface{}]interface{}{}
	length := 0
	processedBencodedString := bencodedString[1:]
	for processedBencodedString[0] != 'e' {
		bencodeKey := decodeBencode(processedBencodedString)
		if bencodeKey.err != nil {
			return BencodeDecoded{map[interface{}]interface{}{}, 0, bencodeKey.err}
		}
		length += bencodeKey.end
		processedBencodedString = processedBencodedString[bencodeKey.end:]
		bencodeValue := decodeBencode(processedBencodedString)
		if bencodeValue.err != nil {
			return BencodeDecoded{map[interface{}]interface{}{}, 0, bencodeValue.err}
		}
		length += bencodeValue.end
		processedBencodedString = processedBencodedString[bencodeValue.end:]
		dict[bencodeKey.value] = bencodeValue.value
	}
	end := len(bencodedString) - len(processedBencodedString) + 1
	return BencodeDecoded{dict, end, nil}
}

func main() {
	command := os.Args[1]

	if command == "decode" {
		bencodedValue := os.Args[2]

		decoded := decodeBencode(bencodedValue)
		if decoded.err != nil {
			fmt.Println(decoded.err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded.value)
		fmt.Println(string(jsonOutput))
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
