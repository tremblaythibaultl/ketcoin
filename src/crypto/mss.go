package crypto

import (
	"crypto/sha256"
	"fmt"
	"math"
)

// Merkle signature scheme parameters

// nbMessages lets this instance of the MSS sign at most nbMessages messages. The height of the tree is lg(nbMessages).
const nbMessages = 8

// height of the tree. nbMessages is 2^height.
const height = 3

// Main tree for the Merkle signature scheme. This object is the secret key.
type MerkleSigTree struct {
	hashTree       [2 * nbMessages][n]byte // root at [len-2]
	leaves         [nbMessages]*oneTimeSig
	traversalIndex int
}

// one can derive the MSS public key from this
type mssSignature struct {
	index        int
	otsSignature [t][n]byte
	otsPublicKey [t][n]byte
	authPath     [height][n]byte
}

func main() {
	d := sha256.Sum256([]byte("GM!"))

	//OTS := newWots()
	//signature :=
	//gn(OTS, d)
	//WotsVerify(signature, OTS.publicKey, d)

	merkleTree := NewMSS()
	signature := sign(merkleTree, d)
	verify(signature, merkleTree.hashTree[len(merkleTree.hashTree)-2], d)
}

func NewMSS() *MerkleSigTree {
	tree := MerkleSigTree{}
	treeInit(&tree)
	return &tree
}

func (sigTree *MerkleSigTree) GetPublicKey() []byte {
	return sigTree.hashTree[2*nbMessages-2][:]
}

// Initializes the new tree with its signature keys and hash-valued nodes
func treeInit(tree *MerkleSigTree) {
	for i := 0; i < nbMessages; i++ {
		tree.leaves[i] = newWots()

		tree.hashTree[i] = hashWotsPublicKey(tree.leaves[i].publicKey) // hash of the public key of the one-time signature
	}

	// construction of the node hashes
	for i := 0; i < nbMessages-1; i++ { //ignoring the last cell of the array
		concat := append(tree.hashTree[2*i][:], tree.hashTree[2*i+1][:]...)
		tree.hashTree[nbMessages+i] = sha256.Sum256(concat)
	}

	// set the first WOTS to use as the first one in the list
	tree.traversalIndex = 0
}

func hashWotsPublicKey(publicKey [t][n]byte) [32]byte {
	var concat []byte
	for j := 0; j < t; j++ {
		concat = append(concat, publicKey[j][:]...)
	}
	return sha256.Sum256(concat) // hash of the public key of the one-time signature
}

func sign(tree *MerkleSigTree, digest [n]byte) *mssSignature {
	signature := mssSignature{}
	signature.index = tree.traversalIndex

	// compute OTS signature
	var otsSignature [t][n]byte

	if tree.traversalIndex == nbMessages {
		signature.otsSignature = otsSignature
		return &signature
	} else {
		signature.otsSignature = wotsSign(tree.leaves[tree.traversalIndex], digest)
	}

	signature.otsPublicKey = tree.leaves[tree.traversalIndex].publicKey

	// compute authentication path, which is the sequence of
	// sibling nodes of the nodes in the path from the leaf to the root
	for i := 0; i < height; i++ {
		levelIndex := 0
		offset := 0
		for j := 0; j < i; j++ {
			levelIndex += int(nbMessages / math.Pow(2, float64(j)))
		}

		if int(math.Floor(float64(tree.traversalIndex)/math.Pow(2, float64(i))))%2 == 0 {
			offset = int(math.Floor(float64(tree.traversalIndex)/math.Pow(2, float64(i)) + 1))
		} else {
			offset = int(math.Floor(float64(tree.traversalIndex)/math.Pow(2, float64(i)) - 1))
		}

		signature.authPath[i] = tree.hashTree[levelIndex+offset]
	}

	tree.traversalIndex++
	return &signature
}

func verify(signature *mssSignature, mssPublicKey [n]byte, digest [n]byte) {
	wotsVerify(signature.otsSignature, signature.otsPublicKey, digest)

	// verify authenticity of the OTS public key by computing the root hash from the auth path
	// at the end of the loop, authPathHash is the hash tree root of the signer, which is also its public key
	authPathHash := hashWotsPublicKey(signature.otsPublicKey)
	for i := 1; i <= height; i++ {
		if int(math.Floor(float64(signature.index)/math.Pow(2, float64(i))))%2 == 0 {
			authPathHash = sha256.Sum256(append(authPathHash[:], signature.authPath[i-1][:]...))
		} else {
			authPathHash = sha256.Sum256(append(signature.authPath[i-1][:], authPathHash[:]...))
		}
	}

	if authPathHash != mssPublicKey {
		fmt.Println("Invalid MSS signature!")
		// TODO : handle invalid signature
	}
}
