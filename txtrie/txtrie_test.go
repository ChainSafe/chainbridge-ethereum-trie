// Copyright 2020 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package txtrie

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
)

var (
	emptyHash = common.HexToHash("")
)

func addTrie(txTries *TxTries, root common.Hash, transactions types.Transactions) error {
	if transactions == nil {
		transactions = types.Transactions{}
	}

	err := txTries.CreateNewTrie(root, transactions)

	if err != nil {
		return err
	}

	return nil

}

func computeEthReferenceTrieHash(transactions types.Transactions) (common.Hash, error) {
	newTrie, err := trie.New(emptyRoot, trie.NewDatabase(nil))
	if err != nil {
		return emptyHash, err
	}

	for i, tx := range transactions {

		key, err := rlp.EncodeToBytes(uint(i))
		if err != nil {
			return emptyHash, err
		}

		value, err := rlp.EncodeToBytes(tx)
		if err != nil {
			return emptyHash, err
		}

		err = newTrie.TryUpdate(key, value)
		if err != nil {
			return emptyHash, err
		}
	}

	return newTrie.Hash(), nil
}

func TestAddEmptyTrie(t *testing.T) {
	txTries := NewTxTries()
	err := addTrie(txTries, emptyRoot, nil)
	if err != nil {
		t.Fatal(err)
	}

	if txTries.txRoots[0] != emptyRoot {
		t.Fatalf("failed to set txRoot in txTries properly, expected: %x, got: %x", emptyRoot, txTries.txRoots[0])
	}

	if txTries.txTries[txTries.txRoots[0]].Hash() != emptyRoot {
		t.Fatalf("trie does not have empty hash as root, expected: %x, got: %x", emptyRoot, txTries.txTries[txTries.txRoots[0]].Hash())
	}

}

func TestAddEmptyTrieRetrieveProof_Fails(t *testing.T) {
	txTries := NewTxTries()
	err := addTrie(txTries, emptyRoot, nil)
	if err != nil {
		t.Fatal(err)
	}

	if txTries.txRoots[0] != emptyRoot {
		t.Fatalf("failed to set txRoot in txTries properly, expected: %x, got: %x", emptyRoot, txTries.txRoots[0])
	}

	if txTries.txTries[txTries.txRoots[0]].Hash() != emptyRoot {
		t.Fatalf("trie does not have empty hash as root, expected: %x, got: %x", emptyRoot, txTries.txTries[txTries.txRoots[0]].Hash())
	}
}

func TestAddSingleTrieUpdate(t *testing.T) {
	vals := GetTransactions1()
	expectedRoot, err := computeEthReferenceTrieHash(vals)
	if err != nil {
		t.Fatal(err)
	}

	txTries := NewTxTries()
	err = addTrie(txTries, expectedRoot, vals)
	if err != nil {
		t.Fatal(err)
	}

	if txTries.txRoots[0] != expectedRoot {
		t.Fatalf("failed to set txRoot in txTries properly, expected: %x, got: %x", expectedRoot, txTries.txRoots[0])
	}

	if txTries.txTries[txTries.txRoots[0]].Hash() != expectedRoot {
		t.Fatalf("trie does not have empty hash as root, expected: %x, got: %x", expectedRoot, txTries.txTries[txTries.txRoots[0]].Hash())
	}
}

func TestAddSingleTrieRetrieveProof(t *testing.T) {
	vals := GetTransactions1()
	expectedRoot, err := computeEthReferenceTrieHash(vals)
	if err != nil {
		t.Fatal(err)
	}

	txTries := NewTxTries()
	err = addTrie(txTries, expectedRoot, vals)
	if err != nil {
		t.Fatal(err)
	}

	if txTries.txRoots[0] != expectedRoot {
		t.Fatalf("failed to set txRoot in txTries properly, expected: %x, got: %x", expectedRoot, txTries.txRoots[0])
	}

	if txTries.txTries[txTries.txRoots[0]].Hash() != expectedRoot {
		t.Fatalf("trie does not have empty hash as root, expected: %x, got: %x", expectedRoot, txTries.txTries[txTries.txRoots[0]].Hash())
	}

	keyToRetrieve, err := rlp.EncodeToBytes(uint(0))
	if err != nil {
		t.Fatal(err)
	}

	proofDb, err := txTries.RetrieveProof(expectedRoot, keyToRetrieve)
	if err != nil {
		t.Fatal(err)
	}

	exists, err := VerifyProof(expectedRoot, keyToRetrieve, proofDb)
	if err != nil {
		t.Fatal(err)
	}

	if exists != true {
		t.Fatalf("not able to verify retrieved proof!")
	}
}

func TestRetrieveEncodedProof(t *testing.T) {
	txTries := NewTxTries()

	vals1 := GetTransactions1()
	expectedRoot1, err := computeEthReferenceTrieHash(vals1)
	if err != nil {
		t.Fatal(err)
	}
	err = addTrie(txTries, expectedRoot1, vals1)
	if err != nil {
		t.Fatal(err)
	}

	key := uint(0)

	keyRlp, err := rlp.EncodeToBytes(key)

	if err != nil {
		t.Error("failed to encode key")
	}

	_, err = txTries.RetrieveEncodedProof(expectedRoot1, keyRlp)

	if err != nil {
		t.Error("unable to rerieve proof")
	}
}
