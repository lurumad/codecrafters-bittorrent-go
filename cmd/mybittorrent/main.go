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
	ErrBencodeList    = errors.New("invalid bencode list")
)

type BencodeType = interface{}

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345
func decodeBencode(bencodedString string) (BencodeType, int, error) {
	if unicode.IsDigit(rune(bencodedString[0])) {
		return decodeString(bencodedString)
	} else if bencodedString[0] == 'i' {
		return decodeInteger(bencodedString)
	} else if bencodedString[0] == 'l' {
		return decodeList(bencodedString)
	} else {
		return "", 0, fmt.Errorf("only strings are supported at the moment")
	}
}

func decodeString(bencodedString string) (string, int, error) {
	firstColonIndex := strings.Index(bencodedString, ":")
	if firstColonIndex == -1 {
		return "", 0, ErrBencodeString
	}

	lengthStr := bencodedString[:firstColonIndex]

	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", 0, err
	}

	end := firstColonIndex + 1 + length
	return bencodedString[firstColonIndex+1 : end], end, nil
}

func decodeInteger(bencodedString string) (int, int, error) {
	endIndex := strings.Index(bencodedString, "e")
	if endIndex == -1 {
		return 0, 0, ErrBencodeInteger
	}

	integer, err := strconv.Atoi(bencodedString[1:endIndex])
	if err != nil {
		return 0, 0, err
	}
	return integer, endIndex + 1, nil
}

func decodeList(bencodedString string) (BencodeType, int, error) {
	list := make([]BencodeType, 0)
	length := 0
	processedBencodedString := bencodedString[1:]
	for processedBencodedString[0] != 'e' {
		value, end, err := decodeBencode(processedBencodedString)
		if err != nil {
			return make([]BencodeType, 0), 0, err
		}
		list = append(list, value)
		length += end
		processedBencodedString = processedBencodedString[end:]
	}

	return list, len(bencodedString) - len(processedBencodedString) + 1, nil
}

func main() {
	command := os.Args[1]

	if command == "decode" {
		bencodedValue := os.Args[2]

		decoded, _, err := decodeBencode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
