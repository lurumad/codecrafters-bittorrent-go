package main

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

var (
	ErrBencodeInteger = errors.New("invalid bencode integer")
	ErrBencodeString  = errors.New("invalid bencode string")
)

type Bencode struct{}

type BencodeType = interface{}

type BencodeDecoded struct {
	value BencodeType
	end   int
	err   error
}

type BencodeEncoded struct {
	value string
	err   error
}

func NewBencode() *Bencode {
	return &Bencode{}
}

func (b *Bencode) decode(bencode string) BencodeDecoded {
	if unicode.IsDigit(rune(bencode[0])) {
		return b.decodeString(bencode)
	} else if bencode[0] == 'i' {
		return b.decodeInteger(bencode)
	} else if bencode[0] == 'l' {
		return b.decodeList(bencode)
	} else if bencode[0] == 'd' {
		return b.decodeDictionary(bencode)
	} else {
		return BencodeDecoded{"", 0, fmt.Errorf("type not supported at the moment: %s", bencode)}
	}
}

func (b *Bencode) decodeString(bencode string) BencodeDecoded {
	firstColonIndex := strings.Index(bencode, ":")
	if firstColonIndex == -1 {
		return BencodeDecoded{"", 0, ErrBencodeString}
	}
	lengthStr := bencode[:firstColonIndex]
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return BencodeDecoded{"", 0, err}
	}
	end := firstColonIndex + 1 + length
	return BencodeDecoded{bencode[firstColonIndex+1 : end], end, nil}
}

func (b *Bencode) decodeInteger(bencode string) BencodeDecoded {
	endIndex := strings.Index(bencode, "e")
	if endIndex == -1 {
		return BencodeDecoded{0, 0, ErrBencodeInteger}
	}
	integer, err := strconv.Atoi(bencode[1:endIndex])
	if err != nil {
		return BencodeDecoded{0, 0, err}
	}
	return BencodeDecoded{integer, endIndex + 1, nil}
}

func (b *Bencode) decodeList(bencode string) BencodeDecoded {
	list := make([]BencodeType, 0)
	processedBencode := bencode[1:]
	for processedBencode[0] != 'e' {
		bencodeDecoded := b.decode(processedBencode)
		if bencodeDecoded.err != nil {
			return BencodeDecoded{make([]BencodeType, 0), 0, bencodeDecoded.err}
		}
		processedBencode = processedBencode[bencodeDecoded.end:]
		list = append(list, bencodeDecoded.value)
	}
	end := len(bencode) - len(processedBencode) + 1
	return BencodeDecoded{list, end, nil}
}

func (b *Bencode) decodeDictionary(bencode string) BencodeDecoded {
	dict := make(map[string]interface{})
	processedBencode := bencode[1:]
	for processedBencode[0] != 'e' {
		key := b.decode(processedBencode)
		if key.err != nil {
			return BencodeDecoded{map[interface{}]interface{}{}, 0, key.err}
		}
		processedBencode = processedBencode[key.end:]
		value := b.decode(processedBencode)
		if value.err != nil {
			return BencodeDecoded{map[interface{}]interface{}{}, 0, value.err}
		}
		processedBencode = processedBencode[value.end:]
		dict[key.value.(string)] = value.value
	}
	end := len(bencode) - len(processedBencode) + 1
	return BencodeDecoded{dict, end, nil}
}

func (b *Bencode) encode(bencodeType BencodeType) BencodeEncoded {
	if reflect.TypeOf(bencodeType).Kind() == reflect.String {
		return BencodeEncoded{b.encodeString(bencodeType.(string)), nil}
	} else if reflect.TypeOf(bencodeType).Kind() == reflect.Int {
		return BencodeEncoded{b.encodeInteger(bencodeType.(int)), nil}
	} else if reflect.TypeOf(bencodeType).Kind() == reflect.Slice {
		return BencodeEncoded{b.encodeList(bencodeType.([]interface{})), nil}
	} else if reflect.TypeOf(bencodeType).Kind() == reflect.Map {
		return BencodeEncoded{b.encodeDictionary(bencodeType.(map[string]interface{})), nil}
	} else {
		return BencodeEncoded{"", fmt.Errorf("type not supported at the moment: %v", bencodeType)}
	}
}

func (b *Bencode) encodeString(value string) string {
	return fmt.Sprintf("%d:%s", len(value), value)
}

func (b *Bencode) encodeInteger(value int) string {
	return fmt.Sprintf("i%de", value)
}

func (b *Bencode) encodeList(value []interface{}) string {
	var builder strings.Builder
	builder.WriteString("l")
	for _, value := range value {
		builder.WriteString(b.encode(value).value)
	}
	builder.WriteString("e")
	return builder.String()
}

func (b *Bencode) encodeDictionary(dict map[string]interface{}) string {
	var builder strings.Builder
	builder.WriteString("d")
	for key, value := range dict {
		builder.WriteString(b.encode(key).value)
		builder.WriteString(b.encode(value).value)
	}
	builder.WriteString("e")
	return builder.String()
}
