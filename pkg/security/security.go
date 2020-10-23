package security

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"time"
)

// GetRandomString GetRandomString
func GetRandomString(l int) string {
	// str := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	bytes := []byte(str)
	result := []byte{}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}

	return string(result)
}

// GetRandomNum GetRandomNum
func GetRandomNum(l int) string {
	str := "0123456789"
	bytes := []byte(str)
	result := []byte{}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}

	return string(result)
}

// GetRandomSixDigital GetRandomSixDigital
func GetRandomSixDigital() string {
	var b [8]byte
	_, err := rand.Read(b[:])
	if err != nil {
		log.Panicln(err)
	}
	c := binary.LittleEndian.Uint64(b[:]) % 1000000
	if c < 100000 {
		c = 100000 + c
	}
	return fmt.Sprintf("%d", c)
}

// RandomKey random key
func RandomKey(l int) (string, error) {
	key := make([]byte, l)

	_, e := rand.Read(key)

	if e != nil {
		fmt.Printf("random key error:%v", e)
		return "", e
	}

	result := base64.StdEncoding.EncodeToString(key)
	return result[:l], nil
}
