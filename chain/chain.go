package chain

type Blockchain struct {
	Chain  []string
	Blocks map[string]*Block

	cfg *Config
}

func NewBlockchain(cfg *Config) *Blockchain {
	bc := &Blockchain{
		Chain:  []string{},
		Blocks: make(map[string]*Block),
		cfg:    cfg,
	}

	bc.NewGenesisBlock()

	return bc
}

func (bc *Blockchain) NewGenesisBlock() *Block {
	newBlock := NewBlock("Genesis block", []byte{}, bc.cfg)
	bc.Blocks[string(newBlock.Hash)] = newBlock
	bc.Chain = append(bc.Chain, string(newBlock.Hash))

	return newBlock
}

func (bc *Blockchain) AddBlock(data string) *Block {
	prevBlock := bc.Blocks[bc.Chain[len(bc.Chain)-1]]
	newBlock := NewBlock(data, prevBlock.Hash, bc.cfg)
	bc.Blocks[string(newBlock.Hash)] = newBlock
	bc.Chain = append(bc.Chain, string(newBlock.Hash))

	return newBlock
}
