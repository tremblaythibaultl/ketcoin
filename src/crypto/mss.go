package crypto

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
)

// Merkle signature scheme parameters

// nbMessages lets this instance of the MSS sign at most nbMessages messages. The height of the tree is lg(nbMessages).
const nbMessages = 1024

// height of the tree. nbMessages is 2^height.
const height = 3

// Main tree for the Merkle signature scheme. This object is the secret key.
type MerkleSigTree struct {
	hashTree       [2 * nbMessages][n]byte // root at [len-2]
	leaves         [nbMessages]*oneTimeSig
	traversalIndex int
}

// one can derive the MSS public key from this
type MssSignature struct {
	Index        int
	OtsSignature [t][n]byte
	OtsPublicKey [t][n]byte
	AuthPath     [height][n]byte
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

		tree.hashTree[i] = hashWotsPublicKey(tree.leaves[i].PublicKey) // hash of the public key of the one-time signature
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

func (tree *MerkleSigTree) Sign(digest [n]byte) *MssSignature {
	signature := MssSignature{}
	signature.Index = tree.traversalIndex

	// compute OTS signature
	var otsSignature [t][n]byte

	if tree.traversalIndex == nbMessages {
		signature.OtsSignature = otsSignature
		return &signature
	} else {
		signature.OtsSignature = wotsSign(tree.leaves[tree.traversalIndex], digest)
	}

	signature.OtsPublicKey = tree.leaves[tree.traversalIndex].PublicKey
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

		signature.AuthPath[i] = tree.hashTree[levelIndex+offset]
	}

	tree.traversalIndex++
	return &signature
}

func Verify(signature *MssSignature, mssPublicKey [n]byte, digest [n]byte) bool {
	wotsVerify(signature.OtsSignature, signature.OtsPublicKey, digest)

	// verify authenticity of the OTS public key by computing the root hash from the auth path
	// at the end of the loop, authPathHash is the hash tree root of the signer, which is also its public key
	authPathHash := hashWotsPublicKey(signature.OtsPublicKey)
	for i := 0; i < height; i++ {
		if int(math.Floor(float64(signature.Index)/math.Pow(2, float64(i))))%2 == 0 {
			authPathHash = sha256.Sum256(append(authPathHash[:], signature.AuthPath[i][:]...))
		} else {
			authPathHash = sha256.Sum256(append(signature.AuthPath[i][:], authPathHash[:]...))
		}
	}

	if authPathHash != mssPublicKey {
		fmt.Println("Invalid MSS signature!")
		return false
	}

	return true
}

func UnmarshalJSON(data []byte) *MerkleSigTree {
	p := &struct {
		HashTree       [2 * nbMessages][n]byte
		Leaves         [nbMessages]*oneTimeSig
		TraversalIndex int
	}{}

	json.Unmarshal(data, p)
	mss := &MerkleSigTree{
		hashTree:       p.HashTree,
		leaves:         p.Leaves,
		traversalIndex: p.TraversalIndex,
	}

	return mss
}

func (mss *MerkleSigTree) MarshalJSON() ([]byte, error) {
	j, err := json.Marshal(struct {
		HashTree       [2 * nbMessages][n]byte
		Leaves         [nbMessages]*oneTimeSig
		TraversalIndex int
	}{
		HashTree:       mss.hashTree,
		Leaves:         mss.leaves,
		TraversalIndex: mss.traversalIndex,
	})

	if err != nil {
		return nil, err
	}
	return j, nil
}

func GetByteArrayAsString(array [t][n]byte) string {
	var s []byte
	for i, row := range array {
		s = append(s, row[i])
	}
	return string(s)
}
