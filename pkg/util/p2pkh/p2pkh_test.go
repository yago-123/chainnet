package util_p2pkh

import (
	"github.com/btcsuite/btcutil/base58"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenerateP2PKHAddrFromPubKey(t *testing.T) {
	p2pkhAddr, err := GenerateP2PKHAddrFromPubKey(
		base58.Decode("aSq9DsNNvGhYxYyqA9wd2eduEAZ5AXWgJTbTJdddpT9aV3HbEPRuBpyEXFktCPCgrdp3FEXrfqjz2xoeQwTCqBs8qJtUFNmCLRTyVaTYuy7G8RZnHkABrMpH2cCG"),
		1,
	)

	require.NoError(t, err)
	require.Len(t, string(p2pkhAddr), P2PKHAddressLength)
	assert.Equal(t, "agr72ArMnsmdm9XTScgCpXnwkhAANyBCd", base58.Encode(p2pkhAddr))
}

func TestExtractPubKeyHashedFromP2PKHAddr(t *testing.T) {
	pubKeyHash, version, err := ExtractPubKeyHashedFromP2PKHAddr(
		base58.Decode("agr72ArMnsmdm9XTScgCpXnwkhAANyBCd"),
	)

	require.NoError(t, err)
	assert.Equal(t, 1, int(version))
	assert.Equal(t, "2ajHyKQLikZqXV9rpaSfnV6mh7a5", base58.Encode(pubKeyHash))
	assert.Len(t, pubKeyHash, 20)

	// modify byte to test checksum validation
	_, _, err = ExtractPubKeyHashedFromP2PKHAddr(
		base58.Decode("agr72ArMnsmd99XTScgCpXnwkhAANyBCd"),
	)

	require.Error(t, err)

	// make sure that length is checked
	_, _, err = ExtractPubKeyHashedFromP2PKHAddr(
		base58.Decode("agr72ArMnsmd9XTScgCpXnwkhAANyBCd"),
	)
	require.Error(t, err)
}
