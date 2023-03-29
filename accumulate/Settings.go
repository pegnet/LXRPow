package accumulate

type MiningSettings struct {
	Block      uint32 // Activation block height
	Difficulty uint64 // Difficulty target at Activation and going forward
	Window     uint16 // Window for averaging block times
	BlockTime  uint32 // Target time (or fixed block time)
}
