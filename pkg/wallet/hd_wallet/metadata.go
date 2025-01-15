package hdwallet

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Metadata struct {
	NumAccounts      uint              `json:"num_accounts"`
	MetadataAccounts []MetadataAccount `json:"metadata_accounts"`
}

type MetadataAccount struct {
	NumExternalWallets uint `json:"num_external_wallets"`
	NumInternalWallets uint `json:"num_internal_wallets"`
}

func SaveMetadata(path string, metadata *Metadata) error {
	// marshal the metadata
	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling metadata: %w", err)
	}

	// create the file
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	// write the content
	_, err = file.Write(jsonData)
	if err != nil {
		return fmt.Errorf("error writing metadata: %w", err)
	}

	return nil
}

func LoadMetadata(path string) (*Metadata, error) {
	// open the file
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// read the file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// unmarshal the JSON data
	var metadata Metadata
	err = json.Unmarshal(fileContent, &metadata)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling metadata: %w", err)
	}

	return &metadata, nil
}
