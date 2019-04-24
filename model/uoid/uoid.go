package uoid

import (
	"crypto/rand"
	"log"
	"strconv"
	"time"
)

type UOID string

func New() UOID {
	return UOID(secureRandomString(6) + strconv.FormatInt(time.Now().UTC().Unix(), 16))
}

func (u *UOID) String() string {
	return string(*u)
}

func FromString(str string) UOID {
	return UOID(str)
}

func secureRandomBytes(length int) []byte {
	var randomBytes = make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		log.Fatal("Unable to generate random bytes")
	}
	return randomBytes
}

func secureRandomString(length int) string {
	letters := "abcdefghijklmnopqrstuvwxyz01234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890"

	// Compute bitMask
	availableCharLength := len(letters)
	if availableCharLength == 0 || availableCharLength > 256 {
		panic("availableCharBytes length must be greater than 0 and less than or equal to 256")
	}
	var bitLength byte
	var bitMask byte
	for bits := availableCharLength - 1; bits != 0; {
		bits = bits >> 1
		bitLength++
	}
	bitMask = 1<<bitLength - 1

	// Compute bufferSize
	bufferSize := length + length / 3

	// Create random string
	result := make([]byte, length)
	for i, j, randomBytes := 0, 0, []byte{}; i < length; j++ {
		if j%bufferSize == 0 {
			// Random byte buffer is empty, get a new one
			randomBytes = secureRandomBytes(bufferSize)
		}
		// Mask bytes to get an index into the character slice
		if idx := int(randomBytes[j%length] & bitMask); idx < availableCharLength {
			result[i] = letters[idx]
			i++
		}
	}

	return string(result)
}