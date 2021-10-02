package spider

type ContractType string

const (
	ContractERC20  ContractType = "erc20"
	ContractERC721 ContractType = "erc721"
	ContractGame   ContractType = "game"
	ContractOther  ContractType = "other"
)

type ContractInfo struct {
	Address string       `toml:"Address"`
	Type    ContractType `toml:"Type"`
}

type GameInfo struct {
	Name      string         `toml:"Name"`
	ID        string         `toml:"ID"`
	Contracts []ContractInfo `toml:"Contracts"`
}
