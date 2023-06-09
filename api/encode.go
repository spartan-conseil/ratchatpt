package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

func generateKey(length int) []byte {
	letters := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	rand.Seed(time.Now().UnixNano())

	s := make([]byte, length)

	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}

	return s
}

func encrypt(key []byte, message []byte) []byte {
	keyLen := len(key)
	encrypted := make([]byte, len(message))

	for i, m := range message {
		encrypted[i] = m ^ key[i%keyLen]
	}

	return encrypted
}

func EncodePayload(payload string) string {
	bytePayload := []byte(payload)
	key := generateKey(32)
	encrypted := encrypt(key, bytePayload)
	b64Encrypted := base64.StdEncoding.EncodeToString(encrypted)

	return "{\"prompt\":\"" + string(key) + "\", \"completion\":\"" + b64Encrypted + "\"}"

}

type GPTFile struct {
	Prompt     string `json:"prompt"`
	Completion string `json:"completion"`
}

func DecodePayload(payload []byte) string {
	gptFile := new(GPTFile)

	err := json.Unmarshal(payload, gptFile)
	if err != nil {
		fmt.Println("error:", err)
		return ""
	}

	key := []byte(gptFile.Prompt)
	encrypted, _ := base64.StdEncoding.DecodeString(gptFile.Completion)

	decrypted := encrypt(key, encrypted)

	return string(decrypted)
}

// FNV1A is not a cryptographic hash but it fit the need here
func FNV1A(text string) string {
	hash := uint64(0xcbf29ce484222325)

	data := []byte(text)

	for _, b := range data {
		hash = hash ^ uint64(b)
		hash = hash * uint64(0x00000100000001B3)
	}
	hash = hash & 0xFFFFFFFF
	h := fmt.Sprintf("%08x", hash)
	return h
}
