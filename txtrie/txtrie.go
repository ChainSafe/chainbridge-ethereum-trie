// Copyright 2020 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package txtrie

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/ethdb"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	ethtrie "github.com/ethereum/go-ethereum/trie"
)

// TxTries stores all the instances of tries we have on disk
type TxTries struct {
	// TODO: the memory allocated for these is hard to get back, look for better way to have a queue
	txTries map[common.Hash]*ethtrie.Trie
	txRoots []common.Hash // needed to track insertion order
}

var (
	// from https://github.com/ethereum/go-ethereum/blob/bcb308745010675671991522ad2a9e811938d7fb/trie/trie.go#L32
	emptyRoot = common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")
)

// NewTxTries creates a new instance of a TxTries object
func NewTxTries() *TxTries {
	txTrie := &TxTries{
		txTries: make(map[common.Hash]*ethtrie.Trie),
	}
	return txTrie

}

// AddNewTrie adds a new transaction trie to an existing TxTries object
func (t *TxTries) CreateNewTrie(root common.Hash, transactions types.Transactions) error {

	if transactions == nil {
		return errors.New("transactions cannot be nil")
	}

	trie, err := ethtrie.New(emptyRoot, ethtrie.NewDatabase(nil))
	if err != nil {
		return nil
	}

	err = updateTrie(trie, transactions, root)

	if err != nil {
		return err
	}

	t.txRoots = append(t.txRoots, root)
	t.txTries[root] = trie

	if err != nil {
		return err
	}

	return nil
}

// updateTrie updates the transaction trie with root transactionRoot with given transactions
// note that this assumes the slice transactions is in the same order they are in the block
func updateTrie(trie *ethtrie.Trie, transactions types.Transactions, transactionRoot common.Hash) error {
	for i, tx := range transactions {

		key, err := rlp.EncodeToBytes(uint(i))
		if err != nil {
			return err
		}

		value, err := rlp.EncodeToBytes(tx)
		if err != nil {
			return err
		}

		trie.Update(key, value)
	}

	// check if the root hash of the trie matches the transactionRoot
	if trie.Hash().Hex() != transactionRoot.Hex() {
		return errors.New("transaction roots don't match")
	}

	return nil
}

// RetrieveEncodedProof retrieves an encoded Proof for a value at key in trie with root root
func (t *TxTries) RetrieveEncodedProof(root common.Hash, key []byte) ([]byte, error) {
	proofDB, err := t.RetrieveProof(root, key)
	if err != nil {
		return nil, err
	}
	return encodeProofDB(root, key, proofDB)
}

// RetrieveProof retrieves a Proof for a value at key in trie with root root
func (t *TxTries) RetrieveProof(root common.Hash, key []byte) (*ProofDatabase, error) {
	trieToRetrieve := t.txTries[root]

	if trieToRetrieve == nil {
		return nil, errors.New("transaction trie for this transaction root does not exist")
	}

	return retrieveProof(trieToRetrieve, key)
}

func retrieveProof(trie *ethtrie.Trie, key []byte) (*ProofDatabase, error) {
	var proof = NewProofDatabase()
	err := trie.Prove(key, 0, proof)
	if err != nil {
		return nil, err
	}

	return proof, nil
}

// VerifyProof verifies merkle proof on path key against the provided root
func VerifyProof(root common.Hash, key []byte, proof *ProofDatabase) (bool, error) {
	exists, err := verifyProof(root, key, proof)

	if err != nil {
		return false, err
	}

	return exists != nil, nil
}

func verifyProof(rootHash common.Hash, key []byte, proofDb ethdb.KeyValueReader) (value []byte, err error) {
	key = keybytesToHex(key)
	wantHash := rootHash
	for i := 0; ; i++ {
		buf, _ := proofDb.Get(wantHash[:])
		if buf == nil {
			return nil, fmt.Errorf("proof node %d (hash %064x) missing", i, wantHash)
		}
		n, err := decodeNode(wantHash[:], buf)
		if err != nil {
			return nil, fmt.Errorf("bad proof node %d: %v", i, err)
		}
		keyrest, cld := get(n, key, true)
		switch cld := cld.(type) {
		case nil:
			// The trie doesn't contain the key.
			return nil, nil
		case hashNode:
			key = keyrest
			copy(wantHash[:], cld)
		case valueNode:
			return cld, nil
		}
	}
}
