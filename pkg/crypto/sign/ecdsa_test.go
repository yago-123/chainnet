package sign

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestECDSASigner_Verify(t *testing.T) {
	ecdsa := NewECDSASignature()

	// generate public and private keys
	pubKey, privKey, err := ecdsa.NewKeyPair()
	require.NoError(t, err)
	assert.NotEmpty(t, pubKey)
	assert.NotEmpty(t, privKey)

	// sign hash
	hash := []byte("837c7456ccf4f09aac5ccf1f250449005353c83a92340f16e72d0af1a2504d42444faefa53d4300ec3bb6902b50b94e1b26c506fdcf5342552e387335ec35d91d830baedf34350a89cff4bb04e132156ec15c87abb639c1cc82bf7fc3b82d6eec6215b138100922bb407702c067503912d54eaef8530d0439616e967bd27b7e6e02e921a6124fb34f656ae4b4ff7b91815108a6e47d43b024ce25bb2dc430bfbe80598e0518a17e7c5e309ed3c3905be487816f1f41d86cbe64e4497dbaaccfdb687021f549cfb01384f5ce5a6842d24793a6319a99e6da87c44754042d7411ef88afedd7c81786405ee75d83dea9962cac9da2257598242b60df07cb927df8c32e3dc45fc45d456925739103836858b93df039419a4eda331690a2f76ecb4686c0246d2200b082f549eb2eea4386f1d50c1917e85979c99790915bd70489e85")
	signature, err := ecdsa.Sign(hash, privKey)
	require.NoError(t, err)

	// make sure that the signature can be verified
	verified, err := ecdsa.Verify(signature, hash, pubKey)
	require.NoError(t, err)
	assert.True(t, verified)

	// modify the message in order to fail verifying
	hashModified := hash
	hashModified[0] = byte(1)
	verified, err = ecdsa.Verify(signature, hashModified, pubKey)
	require.NoError(t, err)
	assert.False(t, verified)

	// modify the signature in order to fail verifying
	signatureModified := signature
	signatureModified[0] = byte(1)
	verified, err = ecdsa.Verify(signatureModified, hash, pubKey)
	require.NoError(t, err)
	assert.False(t, verified)
}
