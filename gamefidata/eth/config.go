package eth

type ContractType string

const (
	ContractERC20  ContractType = "erc20"
	ContractERC721 ContractType = "erc721"
	ContractGame   ContractType = "game"
	ContractOther  ContractType = "other"
)

type ContractInfo struct {
	Address    string       `toml:"Address"`
	StartBlock string       `toml:"StartBlock"`
	Type       ContractType `toml:"Type"`
}

type GameInfo struct {
	Chain     string         `toml:"Chain"`
	RPCAddr   string         `toml:"RPCAddr"`
	ChainID   int32          `toml:"ChainID"`
	Name      string         `toml:"Name"`
	URL       string         `toml:"URL"`
	Contracts []ContractInfo `toml:"Contracts"`
}

type Config struct {
	Games []GameInfo `toml:"Games"`
}
