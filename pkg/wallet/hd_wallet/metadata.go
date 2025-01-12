package hd_wallet

type Metadata struct {
	NumAccounts      uint
	MetadataAccounts []MetadataAccount
}

type MetadataAccount struct {
	NumExternalWallets uint
	NumInternalWallets uint
}

func SaveMetadata(path string, metadata *Metadata) error {
	return nil
}

func LoadMetadata(path string) (*Metadata, error) {
	return &Metadata{}, nil
}
