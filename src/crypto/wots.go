package crypto

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// Winternitz one-time signature scheme parameters

const n = 32 // Size (in bytes) of the message to sign
const w = 16 // Winternitz parameter
const t = 18 // computed in function of n and w

type oneTimeSig struct {
	SignatureKey [t][n]byte
	PublicKey    [t][n]byte
}

func newWots() *oneTimeSig {
	wots := oneTimeSig{}
	skInit(&wots)
	pkInit(&wots)

	return &wots
}

// Initializes the OTS secret (signature) key with t n-byte random strings
func skInit(wots *oneTimeSig) {
	rand.Seed(time.Now().UnixNano())
	var rvalue uint64

	for i := 0; i < t; i++ {
		for j := 0; j < n; j += 8 {
			rvalue = rand.Uint64()
			b := make([]byte, 8)
			binary.BigEndian.PutUint64(b, rvalue)
			for k := 0; k < 8; k++ {
				wots.SignatureKey[i][j+k] = b[k]
			}
		}
	}
}

// Initializes the OTS public key from the signature key
func pkInit(wots *oneTimeSig) {
	for i := 0; i < t; i++ {
		key := wots.SignatureKey[i]
		for j := 0; j < int(math.Pow(2, w)-1); j++ {
			key = sha256.Sum256(key[:])
		}
		wots.PublicKey[i] = key
	}
}

func computeBitStrings(digest [32]byte) [t]uint16 {
	var bitStrings [t]uint16
	var checksum uint32 = 0
	t1 := math.Ceil(n * 8 / w)

	// compute checksum of each of the t1 strings of length w,
	// and fill the bitStrings array with all the 16-bit chunks of the digest.
	for i := 0; i < int(t1); i++ {
		bitStrings[i] = binary.BigEndian.Uint16(digest[2*i : 2*(i+1)])
		checksum += uint32(math.Pow(2, w)) - uint32(bitStrings[i])
	}

	// t2 := int(math.Ceil((math.Floor(math.Log2(t1)) + 1 + w) / w))
	// computing the last t2 w-bit strings
	bitStrings[16] = uint16(checksum >> 16)
	bitStrings[17] = uint16(checksum)

	return bitStrings
}

// Signs a message digest of 256 bits
func wotsSign(wots *oneTimeSig, digest [n]byte) [t][n]byte {
	var bitStrings = computeBitStrings(digest)
	var signature [t][n]byte
	for i := 0; i < t; i++ {
		sig := wots.SignatureKey[i]
		for j := uint16(0); j < bitStrings[i]; j++ {
			sig = sha256.Sum256(sig[:])
		}
		signature[i] = sig
	}

	return signature
}

func wotsVerify(signature [t][n]byte, publicKey [t][n]byte, digest [n]byte) {
	var bitStrings = computeBitStrings(digest)
	error := false
	for i := 0; i < t; i++ {
		verif := signature[i]
		for j := uint32(0); j < uint32(math.Pow(2, w))-1-uint32(bitStrings[i]); j++ {
			verif = sha256.Sum256(verif[:])
		}
		for j := 0; j < len(verif); j++ {
			if verif[j] != publicKey[i][j] {
				error = true
				// TODO : handle invalid signatures
			}
		}
		if error {
			fmt.Printf("Invalid WOTS signature : \nwots verification=%x\npublic key=%x\n", verif, publicKey[i])
		}
	}
}
