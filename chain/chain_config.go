/*
   Copyright 2021 Erigon contributors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package chain

import (
	"fmt"
	"math/big"
	"sort"
	"strconv"

	"github.com/ledgerwatch/erigon-lib/common"
)

// Config is the core config which determines the blockchain settings.
//
// Config is stored in the database on a per block basis. This means
// that any network, identified by its genesis block, can have its own
// set of configuration options.
type Config struct {
	ChainName string
	ChainID   *big.Int `json:"chainId"` // chainId identifies the current chain and is used for replay protection

	Consensus ConsensusName `json:"consensus,omitempty"` // aura, ethash or clique

	// *Block fields activate the corresponding hard fork at a certain block number,
	// while *Time fields do so based on the block's time stamp.
	// nil means that the hard-fork is not scheduled,
	// while 0 means that it's already activated from genesis.

	// ETH mainnet upgrades
	// See https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades
	HomesteadBlock        *big.Int `json:"homesteadBlock,omitempty"`
	DAOForkBlock          *big.Int `json:"daoForkBlock,omitempty"`
	TangerineWhistleBlock *big.Int `json:"eip150Block,omitempty"`
	SpuriousDragonBlock   *big.Int `json:"eip155Block,omitempty"`
	ByzantiumBlock        *big.Int `json:"byzantiumBlock,omitempty"`
	ConstantinopleBlock   *big.Int `json:"constantinopleBlock,omitempty"`
	PetersburgBlock       *big.Int `json:"petersburgBlock,omitempty"`
	IstanbulBlock         *big.Int `json:"istanbulBlock,omitempty"`
	MuirGlacierBlock      *big.Int `json:"muirGlacierBlock,omitempty"`
	BerlinBlock           *big.Int `json:"berlinBlock,omitempty"`
	LondonBlock           *big.Int `json:"londonBlock,omitempty"`
	ArrowGlacierBlock     *big.Int `json:"arrowGlacierBlock,omitempty"`
	GrayGlacierBlock      *big.Int `json:"grayGlacierBlock,omitempty"`

	// EIP-3675: Upgrade consensus to Proof-of-Stake (a.k.a. "Paris", "The Merge")
	TerminalTotalDifficulty       *big.Int `json:"terminalTotalDifficulty,omitempty"`       // The merge happens when terminal total difficulty is reached
	TerminalTotalDifficultyPassed bool     `json:"terminalTotalDifficultyPassed,omitempty"` // Disable PoW sync for networks that have already passed through the Merge
	MergeNetsplitBlock            *big.Int `json:"mergeNetsplitBlock,omitempty"`            // Virtual fork after The Merge to use as a network splitter; see FORK_NEXT_VALUE in EIP-3675

	// Mainnet fork scheduling switched from block numbers to timestamps after The Merge
	ShanghaiTime     *big.Int `json:"shanghaiTime,omitempty"`
	CancunTime       *big.Int `json:"cancunTime,omitempty"`
	ShardingForkTime *big.Int `json:"shardingForkTime,omitempty"`
	PragueTime       *big.Int `json:"pragueTime,omitempty"`

	// Parlia fork blocks
	RamanujanBlock  *big.Int `json:"ramanujanBlock,omitempty" toml:",omitempty"`  // ramanujanBlock switch block (nil = no fork, 0 = already activated)
	NielsBlock      *big.Int `json:"nielsBlock,omitempty" toml:",omitempty"`      // nielsBlock switch block (nil = no fork, 0 = already activated)
	MirrorSyncBlock *big.Int `json:"mirrorSyncBlock,omitempty" toml:",omitempty"` // mirrorSyncBlock switch block (nil = no fork, 0 = already activated)
	BrunoBlock      *big.Int `json:"brunoBlock,omitempty" toml:",omitempty"`      // brunoBlock switch block (nil = no fork, 0 = already activated)
	EulerBlock      *big.Int `json:"eulerBlock,omitempty" toml:",omitempty"`      // eulerBlock switch block (nil = no fork, 0 = already activated)
	GibbsBlock      *big.Int `json:"gibbsBlock,omitempty" toml:",omitempty"`      // gibbsBlock switch block (nil = no fork, 0 = already activated)
	NanoBlock       *big.Int `json:"nanoBlock,omitempty" toml:",omitempty"`       // nanoBlock switch block (nil = no fork, 0 = already activated)
	MoranBlock      *big.Int `json:"moranBlock,omitempty" toml:",omitempty"`      // moranBlock switch block (nil = no fork, 0 = already activated)
	PlanckBlock     *big.Int `json:"planckBlock,omitempty" toml:",omitempty"`     // planckBlock switch block (nil = no fork, 0 = already activated)
	LubanBlock      *big.Int `json:"lubanBlock,omitempty" toml:",omitempty"`      // lubanBlock switch block (nil = no fork, 0 = already activated)
	PlatoBlock      *big.Int `json:"platoBlock,omitempty" toml:",omitempty"`      // platoBlock switch block (nil = no fork, 0 = already activated)
	HertzBlock      *big.Int `json:"hertzBlock,omitempty" toml:",omitempty"`      // hertzBlock switch block (nil = no fork, 0 = already activated)
	// Gnosis Chain fork blocks
	PosdaoBlock *big.Int `json:"posdaoBlock,omitempty"`

	Eip1559FeeCollector           *common.Address `json:"eip1559FeeCollector,omitempty"`           // (Optional) Address where burnt EIP-1559 fees go to
	Eip1559FeeCollectorTransition *big.Int        `json:"eip1559FeeCollectorTransition,omitempty"` // (Optional) Block from which burnt EIP-1559 fees go to the Eip1559FeeCollector

	// Various consensus engines
	Ethash *EthashConfig `json:"ethash,omitempty"`
	Clique *CliqueConfig `json:"clique,omitempty"`
	Aura   *AuRaConfig   `json:"aura,omitempty"`
	Parlia *ParliaConfig `json:"parlia,omitempty" toml:",omitempty"`
	Bor    *BorConfig    `json:"bor,omitempty"`
}

func (c *Config) String() string {
	engine := c.getEngine()

	if c.Consensus == ParliaConsensus {
		return fmt.Sprintf("{ChainID: %v Ramanujan: %v, Niels: %v, MirrorSync: %v, Bruno: %v, Euler: %v, Gibbs: %v, Nano: %v, Moran: %v, Gibbs: %v, Planck: %v, Luban: %v, Plato: %v, Hertz: %v Engine: %v}",
			c.ChainID,
			c.RamanujanBlock,
			c.NielsBlock,
			c.MirrorSyncBlock,
			c.BrunoBlock,
			c.EulerBlock,
			c.GibbsBlock,
			c.NanoBlock,
			c.MoranBlock,
			c.GibbsBlock,
			c.PlanckBlock,
			c.LubanBlock,
			c.PlatoBlock,
			c.HertzBlock,
			engine,
		)
	}

	return fmt.Sprintf("{ChainID: %v, Homestead: %v, DAO: %v, Tangerine Whistle: %v, Spurious Dragon: %v, Byzantium: %v, Constantinople: %v, Petersburg: %v, Istanbul: %v, Muir Glacier: %v, Berlin: %v, London: %v, Arrow Glacier: %v, Gray Glacier: %v, Terminal Total Difficulty: %v, Merge Netsplit: %v, Shanghai: %v, Cancun: %v, Prague: %v, Engine: %v}",
		c.ChainID,
		c.HomesteadBlock,
		c.DAOForkBlock,
		c.TangerineWhistleBlock,
		c.SpuriousDragonBlock,
		c.ByzantiumBlock,
		c.ConstantinopleBlock,
		c.PetersburgBlock,
		c.IstanbulBlock,
		c.MuirGlacierBlock,
		c.BerlinBlock,
		c.LondonBlock,
		c.ArrowGlacierBlock,
		c.GrayGlacierBlock,
		c.TerminalTotalDifficulty,
		c.MergeNetsplitBlock,
		c.ShanghaiTime,
		c.CancunTime,
		c.PragueTime,
		engine,
	)
}

func (c *Config) getEngine() string {
	switch {
	case c.Ethash != nil:
		return c.Ethash.String()
	case c.Clique != nil:
		return c.Clique.String()
	case c.Parlia != nil:
		return c.Parlia.String()
	case c.Bor != nil:
		return c.Bor.String()
	case c.Aura != nil:
		return c.Aura.String()
	default:
		return "unknown"
	}
}

// IsHomestead returns whether num is either equal to the homestead block or greater.
func (c *Config) IsHomestead(num uint64) bool {
	return isForked(c.HomesteadBlock, num)
}

// IsDAOFork returns whether num is either equal to the DAO fork block or greater.
func (c *Config) IsDAOFork(num uint64) bool {
	return isForked(c.DAOForkBlock, num)
}

// IsTangerineWhistle returns whether num is either equal to the Tangerine Whistle (EIP150) fork block or greater.
func (c *Config) IsTangerineWhistle(num uint64) bool {
	return isForked(c.TangerineWhistleBlock, num)
}

// IsSpuriousDragon returns whether num is either equal to the Spurious Dragon fork block or greater.
func (c *Config) IsSpuriousDragon(num uint64) bool {
	return isForked(c.SpuriousDragonBlock, num)
}

// IsByzantium returns whether num is either equal to the Byzantium fork block or greater.
func (c *Config) IsByzantium(num uint64) bool {
	return isForked(c.ByzantiumBlock, num)
}

// IsConstantinople returns whether num is either equal to the Constantinople fork block or greater.
func (c *Config) IsConstantinople(num uint64) bool {
	return isForked(c.ConstantinopleBlock, num)
}

// IsMuirGlacier returns whether num is either equal to the Muir Glacier (EIP-2384) fork block or greater.
func (c *Config) IsMuirGlacier(num uint64) bool {
	return isForked(c.MuirGlacierBlock, num)
}

// IsPetersburg returns whether num is either
// - equal to or greater than the PetersburgBlock fork block,
// - OR is nil, and Constantinople is active
func (c *Config) IsPetersburg(num uint64) bool {
	return isForked(c.PetersburgBlock, num) || c.PetersburgBlock == nil && isForked(c.ConstantinopleBlock, num)
}

// IsIstanbul returns whether num is either equal to the Istanbul fork block or greater.
func (c *Config) IsIstanbul(num uint64) bool {
	return isForked(c.IstanbulBlock, num)
}

// IsBerlin returns whether num is either equal to the Berlin fork block or greater.
func (c *Config) IsBerlin(num uint64) bool {
	return isForked(c.BerlinBlock, num)
}

// IsLondon returns whether num is either equal to the London fork block or greater.
func (c *Config) IsLondon(num uint64) bool {
	return isForked(c.LondonBlock, num)
}

// IsArrowGlacier returns whether num is either equal to the Arrow Glacier (EIP-4345) fork block or greater.
func (c *Config) IsArrowGlacier(num uint64) bool {
	return isForked(c.ArrowGlacierBlock, num)
}

// IsGrayGlacier returns whether num is either equal to the Gray Glacier (EIP-5133) fork block or greater.
func (c *Config) IsGrayGlacier(num uint64) bool {
	return isForked(c.GrayGlacierBlock, num)
}

// IsShanghai returns whether time is either equal to the Shanghai fork time or greater.
func (c *Config) IsShanghai(time uint64) bool {
	return isForked(c.ShanghaiTime, time)
}

// IsCancun returns whether time is either equal to the Cancun fork time or greater.
func (c *Config) IsCancun(time uint64) bool {
	return isForked(c.CancunTime, time)
}

// IsPrague returns whether time is either equal to the Prague fork time or greater.
func (c *Config) IsPrague(time uint64) bool {
	return isForked(c.PragueTime, time)
}

func (c *Config) IsEip1559FeeCollector(num uint64) bool {
	return c.Eip1559FeeCollector != nil && isForked(c.Eip1559FeeCollectorTransition, num)
}

// IsRamanujan returns whether num is either equal to the IsRamanujan fork block or greater.
func (c *Config) IsRamanujan(num uint64) bool {
	return isForked(c.RamanujanBlock, num)
}

// IsOnRamanujan returns whether num is equal to the Ramanujan fork block
func (c *Config) IsOnRamanujan(num *big.Int) bool {
	return numEqual(c.RamanujanBlock, num)
}

// IsNiels returns whether num is either equal to the Niels fork block or greater.
func (c *Config) IsNiels(num uint64) bool {
	return isForked(c.NielsBlock, num)
}

// IsOnNiels returns whether num is equal to the IsNiels fork block
func (c *Config) IsOnNiels(num *big.Int) bool {
	return numEqual(c.NielsBlock, num)
}

// IsMirrorSync returns whether num is either equal to the MirrorSync fork block or greater.
func (c *Config) IsMirrorSync(num uint64) bool {
	return isForked(c.MirrorSyncBlock, num)
}

// IsOnMirrorSync returns whether num is equal to the MirrorSync fork block
func (c *Config) IsOnMirrorSync(num *big.Int) bool {
	return numEqual(c.MirrorSyncBlock, num)
}

// IsBruno returns whether num is either equal to the Burn fork block or greater.
func (c *Config) IsBruno(num uint64) bool {
	return isForked(c.BrunoBlock, num)
}

// IsOnBruno returns whether num is equal to the Burn fork block
func (c *Config) IsOnBruno(num *big.Int) bool {
	return numEqual(c.BrunoBlock, num)
}

// IsEuler returns whether num is either equal to the euler fork block or greater.
func (c *Config) IsEuler(num *big.Int) bool {
	return isForked(c.EulerBlock, num.Uint64())
}

func (c *Config) IsOnEuler(num *big.Int) bool {
	return numEqual(c.EulerBlock, num)
}

// IsGibbs returns whether num is either equal to the euler fork block or greater.
func (c *Config) IsGibbs(num *big.Int) bool {
	return isForked(c.GibbsBlock, num.Uint64())
}

func (c *Config) IsOnGibbs(num *big.Int) bool {
	return numEqual(c.GibbsBlock, num)
}

func (c *Config) IsMoran(num uint64) bool {
	return isForked(c.MoranBlock, num)
}

func (c *Config) IsOnMoran(num *big.Int) bool {
	return numEqual(c.MoranBlock, num)
}

// IsNano returns whether num is either equal to the euler fork block or greater.
func (c *Config) IsNano(num uint64) bool {
	return isForked(c.NanoBlock, num)
}

func (c *Config) IsOnNano(num *big.Int) bool {
	return numEqual(c.NanoBlock, num)
}

func (c *Config) IsPlanck(num uint64) bool {
	return isForked(c.PlanckBlock, num)
}

func (c *Config) IsOnPlanck(num *big.Int) bool {
	return numEqual(c.PlanckBlock, num)
}

func (c *Config) IsLuban(num uint64) bool {
	return isForked(c.LubanBlock, num)
}

func (c *Config) IsOnLuban(num *big.Int) bool {
	return numEqual(c.LubanBlock, num)
}

func (c *Config) IsPlato(num uint64) bool {
	return isForked(c.PlatoBlock, num)
}

func (c *Config) IsOnPlato(num *big.Int) bool {
	return numEqual(c.PlatoBlock, num)
}

func (c *Config) IsHertz(num uint64) bool {
	return isForked(c.HertzBlock, num)
}

func (c *Config) IsOnHertz(num *big.Int) bool {
	return numEqual(c.HertzBlock, num)
}

// CheckCompatible checks whether scheduled fork transitions have been imported
// with a mismatching chain configuration.
func (c *Config) CheckCompatible(newcfg *Config, height uint64) *ConfigCompatError {
	bhead := height

	// Iterate checkCompatible to find the lowest conflict.
	var lasterr *ConfigCompatError
	for {
		err := c.checkCompatible(newcfg, bhead)
		if err == nil || (lasterr != nil && err.RewindTo == lasterr.RewindTo) {
			break
		}
		lasterr = err
		bhead = err.RewindTo
	}
	return lasterr
}

type forkBlockNumber struct {
	name        string
	blockNumber *big.Int
	optional    bool // if true, the fork may be nil and next fork is still allowed
}

func (c *Config) forkBlockNumbers() []forkBlockNumber {
	return []forkBlockNumber{
		{name: "homesteadBlock", blockNumber: c.HomesteadBlock},
		{name: "daoForkBlock", blockNumber: c.DAOForkBlock, optional: true},
		{name: "eip150Block", blockNumber: c.TangerineWhistleBlock},
		{name: "eip155Block", blockNumber: c.SpuriousDragonBlock},
		{name: "byzantiumBlock", blockNumber: c.ByzantiumBlock},
		{name: "constantinopleBlock", blockNumber: c.ConstantinopleBlock},
		{name: "petersburgBlock", blockNumber: c.PetersburgBlock},
		{name: "istanbulBlock", blockNumber: c.IstanbulBlock},
		{name: "muirGlacierBlock", blockNumber: c.MuirGlacierBlock, optional: true},
		{name: "eulerBlock", blockNumber: c.EulerBlock, optional: true},
		{name: "gibbsBlock", blockNumber: c.GibbsBlock},
		{name: "planckBlock", blockNumber: c.PlanckBlock},
		{name: "lubanBlock", blockNumber: c.LubanBlock},
		{name: "platoBlock", blockNumber: c.PlatoBlock},
		{name: "hertzBlock", blockNumber: c.HertzBlock},
		{name: "berlinBlock", blockNumber: c.BerlinBlock, optional: true},
		{name: "londonBlock", blockNumber: c.LondonBlock, optional: true},
		{name: "arrowGlacierBlock", blockNumber: c.ArrowGlacierBlock, optional: true},
		{name: "grayGlacierBlock", blockNumber: c.GrayGlacierBlock, optional: true},
		{name: "mergeNetsplitBlock", blockNumber: c.MergeNetsplitBlock, optional: true},
	}
}

// CheckConfigForkOrder checks that we don't "skip" any forks
func (c *Config) CheckConfigForkOrder() error {
	if c != nil && c.ChainID != nil && c.ChainID.Uint64() == 77 {
		return nil
	}

	var lastFork forkBlockNumber

	for _, fork := range c.forkBlockNumbers() {
		if lastFork.name != "" {
			// Next one must be higher number
			if lastFork.blockNumber == nil && fork.blockNumber != nil {
				return fmt.Errorf("unsupported fork ordering: %v not enabled, but %v enabled at %v",
					lastFork.name, fork.name, fork.blockNumber)
			}
			if lastFork.blockNumber != nil && fork.blockNumber != nil {
				if lastFork.blockNumber.Cmp(fork.blockNumber) > 0 {
					return fmt.Errorf("unsupported fork ordering: %v enabled at %v, but %v enabled at %v",
						lastFork.name, lastFork.blockNumber, fork.name, fork.blockNumber)
				}
			}
			// If it was optional and not set, then ignore it
		}
		if !fork.optional || fork.blockNumber != nil {
			lastFork = fork
		}
	}
	return nil
}

func (c *Config) checkCompatible(newcfg *Config, head uint64) *ConfigCompatError {
	// returns true if a fork scheduled at s1 cannot be rescheduled to block s2 because head is already past the fork.
	incompatible := func(s1, s2 *big.Int, head uint64) bool {
		return (isForked(s1, head) || isForked(s2, head)) && !numEqual(s1, s2)
	}

	// Ethereum mainnet forks
	if incompatible(c.HomesteadBlock, newcfg.HomesteadBlock, head) {
		return newCompatError("Homestead fork block", c.HomesteadBlock, newcfg.HomesteadBlock)
	}
	if incompatible(c.DAOForkBlock, newcfg.DAOForkBlock, head) {
		return newCompatError("DAO fork block", c.DAOForkBlock, newcfg.DAOForkBlock)
	}
	if incompatible(c.TangerineWhistleBlock, newcfg.TangerineWhistleBlock, head) {
		return newCompatError("Tangerine Whistle fork block", c.TangerineWhistleBlock, newcfg.TangerineWhistleBlock)
	}
	if incompatible(c.SpuriousDragonBlock, newcfg.SpuriousDragonBlock, head) {
		return newCompatError("Spurious Dragon fork block", c.SpuriousDragonBlock, newcfg.SpuriousDragonBlock)
	}
	if c.IsSpuriousDragon(head) && !numEqual(c.ChainID, newcfg.ChainID) {
		return newCompatError("EIP155 chain ID", c.SpuriousDragonBlock, newcfg.SpuriousDragonBlock)
	}
	if incompatible(c.ByzantiumBlock, newcfg.ByzantiumBlock, head) {
		return newCompatError("Byzantium fork block", c.ByzantiumBlock, newcfg.ByzantiumBlock)
	}
	if incompatible(c.ConstantinopleBlock, newcfg.ConstantinopleBlock, head) {
		return newCompatError("Constantinople fork block", c.ConstantinopleBlock, newcfg.ConstantinopleBlock)
	}
	if incompatible(c.PetersburgBlock, newcfg.PetersburgBlock, head) {
		// the only case where we allow Petersburg to be set in the past is if it is equal to Constantinople
		// mainly to satisfy fork ordering requirements which state that Petersburg fork be set if Constantinople fork is set
		if incompatible(c.ConstantinopleBlock, newcfg.PetersburgBlock, head) {
			return newCompatError("Petersburg fork block", c.PetersburgBlock, newcfg.PetersburgBlock)
		}
	}
	if incompatible(c.IstanbulBlock, newcfg.IstanbulBlock, head) {
		return newCompatError("Istanbul fork block", c.IstanbulBlock, newcfg.IstanbulBlock)
	}
	if incompatible(c.MuirGlacierBlock, newcfg.MuirGlacierBlock, head) {
		return newCompatError("Muir Glacier fork block", c.MuirGlacierBlock, newcfg.MuirGlacierBlock)
	}
	if incompatible(c.BerlinBlock, newcfg.BerlinBlock, head) {
		return newCompatError("Berlin fork block", c.BerlinBlock, newcfg.BerlinBlock)
	}
	if incompatible(c.LondonBlock, newcfg.LondonBlock, head) {
		return newCompatError("London fork block", c.LondonBlock, newcfg.LondonBlock)
	}
	if incompatible(c.ArrowGlacierBlock, newcfg.ArrowGlacierBlock, head) {
		return newCompatError("Arrow Glacier fork block", c.ArrowGlacierBlock, newcfg.ArrowGlacierBlock)
	}
	if incompatible(c.GrayGlacierBlock, newcfg.GrayGlacierBlock, head) {
		return newCompatError("Gray Glacier fork block", c.GrayGlacierBlock, newcfg.GrayGlacierBlock)
	}
	if incompatible(c.MergeNetsplitBlock, newcfg.MergeNetsplitBlock, head) {
		return newCompatError("Merge netsplit block", c.MergeNetsplitBlock, newcfg.MergeNetsplitBlock)
	}

	// Parlia forks
	if incompatible(c.RamanujanBlock, newcfg.RamanujanBlock, head) {
		return newCompatError("Ramanujan fork block", c.RamanujanBlock, newcfg.RamanujanBlock)
	}
	if incompatible(c.NielsBlock, newcfg.NielsBlock, head) {
		return newCompatError("Niels fork block", c.NielsBlock, newcfg.NielsBlock)
	}
	if incompatible(c.MirrorSyncBlock, newcfg.MirrorSyncBlock, head) {
		return newCompatError("MirrorSync fork block", c.MirrorSyncBlock, newcfg.MirrorSyncBlock)
	}
	if incompatible(c.BrunoBlock, newcfg.BrunoBlock, head) {
		return newCompatError("Bruno fork block", c.BrunoBlock, newcfg.BrunoBlock)
	}
	if incompatible(c.EulerBlock, newcfg.EulerBlock, head) {
		return newCompatError("Euler fork block", c.EulerBlock, newcfg.EulerBlock)
	}
	if incompatible(c.GibbsBlock, newcfg.GibbsBlock, head) {
		return newCompatError("Gibbs fork block", c.GibbsBlock, newcfg.GibbsBlock)
	}
	if incompatible(c.NanoBlock, newcfg.NanoBlock, head) {
		return newCompatError("Nano fork block", c.NanoBlock, newcfg.NanoBlock)
	}
	if incompatible(c.MoranBlock, newcfg.MoranBlock, head) {
		return newCompatError("moran fork block", c.MoranBlock, newcfg.MoranBlock)
	}
	if incompatible(c.PlanckBlock, newcfg.PlanckBlock, head) {
		return newCompatError("planck fork block", c.PlanckBlock, newcfg.PlanckBlock)
	}
	if incompatible(c.LubanBlock, newcfg.LubanBlock, head) {
		return newCompatError("luban fork block", c.LubanBlock, newcfg.LubanBlock)
	}
	if incompatible(c.PlatoBlock, newcfg.PlatoBlock, head) {
		return newCompatError("plato fork block", c.PlatoBlock, newcfg.PlatoBlock)
	}
	if incompatible(c.HertzBlock, newcfg.HertzBlock, head) {
		return newCompatError("hertz fork block", c.HertzBlock, newcfg.HertzBlock)
	}
	return nil
}

func numEqual(x, y *big.Int) bool {
	if x == nil {
		return y == nil
	}
	if y == nil {
		return x == nil
	}
	return x.Cmp(y) == 0
}

// ConfigCompatError is raised if the locally-stored blockchain is initialised with a
// ChainConfig that would alter the past.
type ConfigCompatError struct {
	What string
	// block numbers of the stored and new configurations
	StoredConfig, NewConfig *big.Int
	// the block number to which the local chain must be rewound to correct the error
	RewindTo uint64
}

func newCompatError(what string, storedblock, newblock *big.Int) *ConfigCompatError {
	var rew *big.Int
	switch {
	case storedblock == nil:
		rew = newblock
	case newblock == nil || storedblock.Cmp(newblock) < 0:
		rew = storedblock
	default:
		rew = newblock
	}
	err := &ConfigCompatError{what, storedblock, newblock, 0}
	if rew != nil && rew.Sign() > 0 {
		err.RewindTo = rew.Uint64() - 1
	}
	return err
}

func (err *ConfigCompatError) Error() string {
	return fmt.Sprintf("mismatching %s in database (have %d, want %d, rewindto %d)", err.What, err.StoredConfig, err.NewConfig, err.RewindTo)
}

// EthashConfig is the consensus engine configs for proof-of-work based sealing.
type EthashConfig struct{}

// String implements the stringer interface, returning the consensus engine details.
func (c *EthashConfig) String() string {
	return "ethash"
}

// CliqueConfig is the consensus engine configs for proof-of-authority based sealing.
type CliqueConfig struct {
	Period uint64 `json:"period"` // Number of seconds between blocks to enforce
	Epoch  uint64 `json:"epoch"`  // Epoch length to reset votes and checkpoint
}

// String implements the stringer interface, returning the consensus engine details.
func (c *CliqueConfig) String() string {
	return "clique"
}

// BorConfig is the consensus engine configs for Matic bor based sealing.
type BorConfig struct {
	Period                map[string]uint64 `json:"period"`                // Number of seconds between blocks to enforce
	ProducerDelay         map[string]uint64 `json:"producerDelay"`         // Number of seconds delay between two producer interval
	Sprint                map[string]uint64 `json:"sprint"`                // Epoch length to proposer
	BackupMultiplier      map[string]uint64 `json:"backupMultiplier"`      // Backup multiplier to determine the wiggle time
	ValidatorContract     string            `json:"validatorContract"`     // Validator set contract
	StateReceiverContract string            `json:"stateReceiverContract"` // State receiver contract

	OverrideStateSyncRecords map[string]int         `json:"overrideStateSyncRecords"` // override state records count
	BlockAlloc               map[string]interface{} `json:"blockAlloc"`

	CalcuttaBlock *big.Int `json:"calcuttaBlock"` // Calcutta switch block (nil = no fork, 0 = already on calcutta)
	JaipurBlock   *big.Int `json:"jaipurBlock"`   // Jaipur switch block (nil = no fork, 0 = already on jaipur)
	DelhiBlock    *big.Int `json:"delhiBlock"`    // Delhi switch block (nil = no fork, 0 = already on delhi)
}

// String implements the stringer interface, returning the consensus engine details.
func (b *BorConfig) String() string {
	return "bor"
}

func (c *BorConfig) CalculateProducerDelay(number uint64) uint64 {
	return c.sprintSize(c.ProducerDelay, number)
}

func (c *BorConfig) CalculateSprint(number uint64) uint64 {
	return c.sprintSize(c.Sprint, number)
}

func (c *BorConfig) CalculateBackupMultiplier(number uint64) uint64 {
	return c.calcConfig(c.BackupMultiplier, number)
}

func (c *BorConfig) CalculatePeriod(number uint64) uint64 {
	return c.calcConfig(c.Period, number)
}

func (c *BorConfig) IsJaipur(number uint64) bool {
	return isForked(c.JaipurBlock, number)
}

func (c *BorConfig) IsDelhi(number uint64) bool {
	return isForked(c.DelhiBlock, number)
}

func (c *BorConfig) IsCalcutta(number uint64) bool {
	return isForked(c.CalcuttaBlock, number)
}

func (c *BorConfig) IsOnCalcutta(number *big.Int) bool {
	return numEqual(c.CalcuttaBlock, number)
}

func (c *BorConfig) calcConfig(field map[string]uint64, number uint64) uint64 {
	keys := sortMapKeys(field)
	for i := 0; i < len(keys)-1; i++ {
		valUint, _ := strconv.ParseUint(keys[i], 10, 64)
		valUintNext, _ := strconv.ParseUint(keys[i+1], 10, 64)
		if number > valUint && number < valUintNext {
			return field[keys[i]]
		}
	}
	return field[keys[len(keys)-1]]
}

func (c *BorConfig) sprintSize(field map[string]uint64, number uint64) uint64 {
	keys := sortMapKeys(field)
	for i := 0; i < len(keys)-1; i++ {
		valUint, _ := strconv.ParseUint(keys[i], 10, 64)
		valUintNext, _ := strconv.ParseUint(keys[i+1], 10, 64)

		if number >= valUint && number < valUintNext {
			return field[keys[i]]
		}
	}

	return field[keys[len(keys)-1]]
}

type ParliaConfig struct {
	DBPath   string
	InMemory bool
	Period   uint64 `json:"period"` // Number of seconds between blocks to enforce
	Epoch    uint64 `json:"epoch"`  // Epoch length to update validatorSet
}

// String implements the stringer interface, returning the consensus engine details.
func (b *ParliaConfig) String() string {
	return "parlia"
}

func sortMapKeys(m map[string]uint64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}

// Rules is syntactic sugar over Config. It can be used for functions
// that do not have or require information about the block.
//
// Rules is a one time interface meaning that it shouldn't be used in between transition
// phases.
type Rules struct {
	ChainID                                                       *big.Int
	IsHomestead, IsTangerineWhistle, IsSpuriousDragon             bool
	IsByzantium, IsConstantinople, IsPetersburg, IsIstanbul       bool
	IsBerlin, IsLondon, IsShanghai, IsCancun                      bool
	IsSharding, IsPrague                                          bool
	IsNano, IsMoran, IsGibbs, IsPlanck, IsLuban, IsPlato, IsHertz bool
	IsEip1559FeeCollector                                         bool
	IsParlia, IsAura                                              bool
}

// Rules ensures c's ChainID is not nil and returns a new Rules instance
func (c *Config) Rules(num uint64, time uint64) *Rules {
	chainID := c.ChainID
	if chainID == nil {
		chainID = new(big.Int)
	}

	return &Rules{
		ChainID:               new(big.Int).Set(chainID),
		IsHomestead:           c.IsHomestead(num),
		IsTangerineWhistle:    c.IsTangerineWhistle(num),
		IsSpuriousDragon:      c.IsSpuriousDragon(num),
		IsByzantium:           c.IsByzantium(num),
		IsConstantinople:      c.IsConstantinople(num),
		IsPetersburg:          c.IsPetersburg(num),
		IsIstanbul:            c.IsIstanbul(num),
		IsBerlin:              c.IsBerlin(num),
		IsLondon:              c.IsLondon(num),
		IsShanghai:            c.IsShanghai(time),
		IsCancun:              c.IsCancun(time),
		IsPrague:              c.IsPrague(time),
		IsNano:                c.IsNano(num),
		IsMoran:               c.IsMoran(num),
		IsPlanck:              c.IsPlanck(num),
		IsLuban:               c.IsLuban(num),
		IsPlato:               c.IsPlato(num),
		IsHertz:               c.IsHertz(num),
		IsEip1559FeeCollector: c.IsEip1559FeeCollector(num),
		IsAura:                c.Aura != nil,
		IsParlia:              true,
	}
}

// isForked returns whether a fork scheduled at block s is active at the given head block.
func isForked(s *big.Int, head uint64) bool {
	if s == nil {
		return false
	}
	return s.Uint64() <= head
}
