# Go Repo Template

## TODO
- [ ] Add license

## Example Usage

Here is an example of how this library could be used.

```
// assume trieDB is some already instanciated leveldb instance
// assume the listener has retrieved the transactions root (txRoot), transactions (txList), and key of the transaction of interest (txPath) for some block while polling

// instanciate new instance of TxTries object
txTries := NewTxTries(3)

// add new trie to the txtries object with relevant txRoot, transactions, and triedb
txTries.AddNewTrie(txRoot, txList, trieDB)

// we can retrieve a proof for our transaction of interest and verify it as follows
txProof := txTries.RetrieveProof(txRoot, txPath)
exists := VerifyProof(txRoot, txPath, txProof)


if exists {
    // we know the transaction exists in our trie
    // perform some action
}



// we can also retrieve the encoded version of the proof for our transaction of interest as follows:
encodedTxProof := txTries.RetrieveEncodedProof(txRoot, txPath)

// encodedTxProof can then be submitted to chain to be verified



```
