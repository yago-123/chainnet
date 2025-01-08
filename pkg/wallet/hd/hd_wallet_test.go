package hd //nolint:testpackage // don't create separate package for tests
import "testing"

const privateKeyHDTest = `
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEII4Ci5GHuv3rO3q1L+2BHcBHqO/immA45VTmswGXxUYkoAoGCCqGSM49
AwEHoUQDQgAETm+qq1qRGebJyaGa6lBmgkC0NlaAo4iKOGEDczvj5A3lK6TLLe9u
0MF7c9jWuMaNt3/lUjAtu8ja9uIALbQyHw==
-----END EC PRIVATE KEY-----
`

func TestHDWallet_Sync(t *testing.T) {
	t.Errorf("implement")
}
