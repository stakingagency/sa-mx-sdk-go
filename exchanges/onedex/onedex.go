package onedex

import (
	"encoding/hex"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stakingagency/sa-mx-sdk-go/accounts"
	"github.com/stakingagency/sa-mx-sdk-go/data"
	"github.com/stakingagency/sa-mx-sdk-go/network"
	"github.com/stakingagency/sa-mx-sdk-go/tokens"
	"github.com/stakingagency/sa-mx-sdk-go/utils"
)

const (
	OneToken     = "ONE-f9954f"
	RoyalMember  = "ONEDEX-df4fac"
	SeedInvestor = "SEEDINV-9e0663"

	liquidityPoolSC = "erd1qqqqqqqqqqqqqpgqqz6vp9y50ep867vnr296mqf3dduh6guvmvlsu3sujc"
	stakingSC       = "erd1qqqqqqqqqqqqqpgql9z9vm8d599ya2r9seklpkcas6qmude4mvlsgrj7hv"
	farmSC          = "erd1qqqqqqqqqqqqqpgq5774jcntdqkzv62tlvvhfn2y7eevpty6mvlszk3dla"
	launchpadSC     = "erd1qqqqqqqqqqqqqpgqxjj6tyrnrdegga4j66s20wql5e9ksq0hmvlssnf6j2"
)

type (
	NewPairCallbackFunc             func(ticker1 string, ticker2 string)
	PairStateChangedCallbackFunc    func(ticker1 string, ticker2 string, newState bool)
	NewStakeCallbackFunc            func(ticker string)
	NewFarmCallbackFunc             func(lpTicker string, rewardTicker string)
	NewDualFarmCallbackFunc         func(lpTicker string, rewardTicker1 string, rewardTicker2 string)
	NewLaunchpadCallbackFunc        func(ticker string)
	LaunchpadEndedCallbackFunc      func(ticker string)
	AnnualRewardChangedCallbackFunc func(farmID uint32, oldReward float64, newReward float64)
	StakeAprChangedCallbackFunc     func(stakeID uint32, oldAPR float64, newAPR float64)
)

type OneDex struct {
	netMan             *network.NetworkManager
	liquidityScAccount *accounts.Account
	stakingScAccount   *accounts.Account
	farmScAccount      *accounts.Account
	launchpadScAccount *accounts.Account
	mxTokens           *tokens.Tokens
	refreshInterval    time.Duration

	liquidityPools    map[uint32]*LiquidityPool
	liquidityPoolsMut sync.Mutex
	farms             map[uint32]*Farm
	farmsMut          sync.Mutex
	stakes            map[uint32]*Stake
	stakesMut         sync.Mutex
	launchpads        map[uint32]*Launchpad
	launchpadsMut     sync.Mutex

	newPairCallback              NewPairCallbackFunc
	pairStateChangedCallback     PairStateChangedCallbackFunc
	newStakeCallback             NewStakeCallbackFunc
	newFarmCallback              NewFarmCallbackFunc
	newDualFarmCallback          NewDualFarmCallbackFunc
	newLaunchpadCallback         NewLaunchpadCallbackFunc
	launchpadEndedCallback       LaunchpadEndedCallbackFunc
	annualReward1ChangedCallback AnnualRewardChangedCallbackFunc
	annualReward2ChangedCallback AnnualRewardChangedCallbackFunc
	stakeAprChangedCallback      StakeAprChangedCallbackFunc
}

var log = logger.GetOrCreate("onedex")

func NewOneDex(netMan *network.NetworkManager, refreshInterval time.Duration) (*OneDex, error) {
	liquidityScAccount, err := accounts.NewAccount(liquidityPoolSC, netMan, 0)
	if err != nil {
		return nil, err
	}

	stakingScAccount, err := accounts.NewAccount(stakingSC, netMan, 0)
	if err != nil {
		return nil, err
	}

	farmScAccount, err := accounts.NewAccount(farmSC, netMan, 0)
	if err != nil {
		return nil, err
	}

	launchpadScAccount, err := accounts.NewAccount(launchpadSC, netMan, 0)
	if err != nil {
		return nil, err
	}

	mxTokens, err := tokens.NewTokens(netMan, refreshInterval)
	if err != nil {
		return nil, err
	}

	one := &OneDex{
		netMan:             netMan,
		liquidityScAccount: liquidityScAccount,
		stakingScAccount:   stakingScAccount,
		farmScAccount:      farmScAccount,
		launchpadScAccount: launchpadScAccount,

		mxTokens:        mxTokens,
		refreshInterval: refreshInterval,

		liquidityPools: make(map[uint32]*LiquidityPool),
		farms:          make(map[uint32]*Farm),
		stakes:         make(map[uint32]*Stake),
		launchpads:     make(map[uint32]*Launchpad),

		newPairCallback:              nil,
		pairStateChangedCallback:     nil,
		newStakeCallback:             nil,
		newFarmCallback:              nil,
		newDualFarmCallback:          nil,
		newLaunchpadCallback:         nil,
		launchpadEndedCallback:       nil,
		annualReward1ChangedCallback: nil,
		annualReward2ChangedCallback: nil,
		stakeAprChangedCallback:      nil,
	}
	one.startTasks()

	return one, nil
}

func (one *OneDex) SetNewPairCallback(f NewPairCallbackFunc) {
	one.newPairCallback = f
}

func (one *OneDex) SetPairStateChangedCallback(f PairStateChangedCallbackFunc) {
	one.pairStateChangedCallback = f
}

func (one *OneDex) SetNewStakeCallback(f NewStakeCallbackFunc) {
	one.newStakeCallback = f
}

func (one *OneDex) SetNewFarmCallback(f NewFarmCallbackFunc) {
	one.newFarmCallback = f
}

func (one *OneDex) SetNewDualFarmCallback(f NewDualFarmCallbackFunc) {
	one.newDualFarmCallback = f
}

func (one *OneDex) SetNewLaunchpadCallback(f NewLaunchpadCallbackFunc) {
	one.newLaunchpadCallback = f
}

func (one *OneDex) SetLaunchpadEndedCallback(f LaunchpadEndedCallbackFunc) {
	one.launchpadEndedCallback = f
}

func (one *OneDex) SetAnnualReward1ChangedCallback(f AnnualRewardChangedCallbackFunc) {
	one.annualReward1ChangedCallback = f
}

func (one *OneDex) SetAnnualReward2ChangedCallback(f AnnualRewardChangedCallbackFunc) {
	one.annualReward2ChangedCallback = f
}

func (one *OneDex) SetStakeAprChangedCallback(f StakeAprChangedCallbackFunc) {
	one.stakeAprChangedCallback = f
}

func (one *OneDex) GetLiquidityPools() (map[uint32]*LiquidityPool, error) {
	keys, err := one.liquidityScAccount.GetAccountKeys("")
	if err != nil {
		return nil, err
	}

	lps := make(map[uint32]*LiquidityPool)

	for key, value := range keys {
		prefix := hex.EncodeToString([]byte("pair_ids.value"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshLiquidityPools", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			lpID, _, ok := utils.ParseUint32(bytes, 0)
			if !ok {
				log.Debug("refreshLiquidityPools", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			idx := 0
			ticker1, idx, ok := utils.ParseString(value, idx)
			allOk := ok
			ticker2, _, ok := utils.ParseString(value, idx)
			allOk = allOk && ok
			if !allOk {
				log.Debug("refreshLiquidityPools", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			lp := lps[lpID]
			if lp == nil {
				lp = &LiquidityPool{
					ID: lpID,
				}
				lps[lpID] = lp
			}

			lp.Token1, err = one.getToken(ticker1)
			if err != nil {
				log.Debug("refreshLiquidityPools", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			lp.Token2, err = one.getToken(ticker2)
			if err != nil {
				log.Debug("refreshLiquidityPools", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}
		}

		prefix = hex.EncodeToString([]byte("pair_lp_token_id"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshLiquidityPools", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			lpID, _, ok := utils.ParseUint32(bytes, 0)
			if !ok {
				log.Debug("refreshLiquidityPools", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			lpToken := string(value)

			lp := lps[lpID]
			if lp == nil {
				lp = &LiquidityPool{
					ID: lpID,
				}
				lps[lpID] = lp
			}
			lp.LpToken, err = one.getToken(lpToken)
			if err != nil {
				log.Debug("refreshLiquidityPools", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}
		}
	}

	for key, value := range keys {
		prefix := hex.EncodeToString([]byte("pair_first_token_reserve"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshLiquidityPools", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			lpID, _, ok := utils.ParseUint32(bytes, 0)
			if !ok {
				log.Debug("refreshLiquidityPools", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			lp := lps[lpID]
			iReserve := big.NewInt(0).SetBytes(value)
			lp.Token1Reserve = utils.Denominate(iReserve, int(lp.Token1.Decimals))
		}

		prefix = hex.EncodeToString([]byte("pair_second_token_reserve"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshLiquidityPools", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			lpID, _, ok := utils.ParseUint32(bytes, 0)
			if !ok {
				log.Debug("refreshLiquidityPools", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			lp := lps[lpID]
			iReserve := big.NewInt(0).SetBytes(value)
			lp.Token2Reserve = utils.Denominate(iReserve, int(lp.Token2.Decimals))
		}

		prefix = hex.EncodeToString([]byte("pair_lp_token_supply"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshLiquidityPools", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			lpID, _, ok := utils.ParseUint32(bytes, 0)
			if !ok {
				log.Debug("refreshLiquidityPools", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			lp := lps[lpID]
			iSupply := big.NewInt(0).SetBytes(value)
			lp.LpTokenSupply = utils.Denominate(iSupply, int(lp.LpToken.Decimals))
		}

		prefix = hex.EncodeToString([]byte("pair_enabled"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshLiquidityPools", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			lpID, _, ok := utils.ParseUint32(bytes, 0)
			if !ok {
				log.Debug("refreshLiquidityPools", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			lp := lps[lpID]
			enabled, _, ok := utils.ParseByte(value, 0)
			if !ok {
				log.Debug("refreshLiquidityPools", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			lp.Enabled = enabled == 1

			prefix = hex.EncodeToString([]byte("pair_state"))
			if strings.HasPrefix(key, prefix) {
				bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
				if err != nil {
					log.Debug("refreshLiquidityPools", "step", "parse keys", "error", "can not decode key", "key", key)
					continue
				}

				lpID, _, ok := utils.ParseUint32(bytes, 0)
				if !ok {
					log.Debug("refreshLiquidityPools", "step", "parse keys", "error", "can not decode key", "key", key)
					continue
				}

				lp := lps[lpID]
				state, _, ok := utils.ParseByte(value, 0)
				if !ok {
					log.Debug("refreshLiquidityPools", "step", "parse keys", "error", "can not decode key", "key", key)
					continue
				}

				lp.State = state
			}
		}
	}

	for _, lp := range lps {
		lp.Token1Price = lp.Token2Reserve / lp.Token1Reserve
	}

	return lps, nil
}

func (one *OneDex) GetFarms() (map[uint32]*Farm, error) {
	keys, err := one.farmScAccount.GetAccountKeys("")
	if err != nil {
		return nil, err
	}

	farms := make(map[uint32]*Farm)

	for key, value := range keys {
		prefix := hex.EncodeToString([]byte("pool_stake_token_id"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshFarms", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			farmID, _, ok := utils.ParseUint32(bytes, 0)
			if !ok {
				log.Debug("refreshFarms", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			lpTicker := string(value)
			farm := farms[farmID]
			if farm == nil {
				farm = &Farm{
					ID:      farmID,
					Farmers: make([]*Farmer, 0),
				}
				farms[farmID] = farm
			}
			farm.LpToken, err = one.getToken(lpTicker)
			if err != nil {
				log.Debug("refreshFarms", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}
		}
	}

	conv, _ := pubkeyConverter.NewBech32PubkeyConverter(32, "erd")
	for key, value := range keys {
		prefix := hex.EncodeToString([]byte("pool_user_stake_amount"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshFarms", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			idx := 0
			farmID, idx, ok := utils.ParseUint32(bytes, 0)
			allOk := ok
			pubkey, _, ok := utils.ParsePubkey(bytes, idx)
			allOk = allOk && ok
			if !allOk {
				log.Debug("refreshFarms", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			farm := farms[farmID]
			address, _ := conv.Encode(pubkey)
			iAmount := big.NewInt(0).SetBytes(value)
			fAmount := utils.Denominate(iAmount, int(farm.LpToken.Decimals))
			farm.Farmers = append(farm.Farmers, &Farmer{
				Address: address,
				Amount:  fAmount,
			})
		}

		prefix = hex.EncodeToString([]byte("pool_reward_token_id"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshFarms", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			farmID, _, ok := utils.ParseUint32(bytes, 0)
			if !ok {
				log.Debug("refreshFarms", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			farm := farms[farmID]
			farm.RewardToken1, err = one.getToken(string(value))
			if err != nil {
				log.Debug("refreshFarms", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}
		}

		prefix = hex.EncodeToString([]byte("pool_second_reward_token_id"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshFarms", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			farmID, _, ok := utils.ParseUint32(bytes, 0)
			if !ok {
				log.Debug("refreshFarms", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			secondRewardToken, _, ok := utils.ParseString(value, 1)
			if !ok {
				log.Debug("refreshFarms", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			farm := farms[farmID]
			farm.RewardToken2, err = one.getToken(secondRewardToken)
			if err != nil {
				log.Debug("refreshFarms", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}
		}
	}

	lastItemIDs := make(map[uint32]uint32)

	for key, value := range keys {
		prefix := hex.EncodeToString([]byte("pool_reward_deposit_amount"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshFarms", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			farmID, _, ok := utils.ParseUint32(bytes, 0)
			if !ok {
				log.Debug("refreshFarms", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			farm := farms[farmID]
			iDepositAmount := big.NewInt(0).SetBytes(value)
			farm.RewardPool1 = utils.Denominate(iDepositAmount, int(farm.RewardToken1.Decimals))
		}

		prefix = hex.EncodeToString([]byte("pool_second_reward_deposit_amount"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshFarms", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			farmID, _, ok := utils.ParseUint32(bytes, 0)
			if !ok {
				log.Debug("refreshFarms", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			farm := farms[farmID]
			iDepositAmount := big.NewInt(0).SetBytes(value)
			farm.RewardPool2 = utils.Denominate(iDepositAmount, int(farm.RewardToken2.Decimals))
		}

		prefix = hex.EncodeToString([]byte("pool_total_stake_amount"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshFarms", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			farmID, _, ok := utils.ParseUint32(bytes, 0)
			if !ok {
				log.Debug("refreshFarms", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			farm := farms[farmID]
			iTotalStaked := big.NewInt(0).SetBytes(value)
			farm.TotalStake = utils.Denominate(iTotalStaked, int(farm.LpToken.Decimals))
		}

		prefix = hex.EncodeToString([]byte("pool_reward_infos"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshFarms", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			farmID, _, ok := utils.ParseUint32(bytes, 0)
			if !ok {
				log.Debug("refreshFarms", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			if !strings.HasPrefix(string(bytes[4:]), ".item") {
				continue
			}

			itemID, _, ok := utils.ParseUint32(bytes, 9)
			if itemID < lastItemIDs[farmID] || !ok {
				continue
			}

			farm := farms[farmID]
			idx := 0
			annualRewardPerLP1, idx, ok := utils.ParseBigInt(value, idx)
			allOk := ok
			var annualRewardPerLP2 *big.Int = nil
			if farm.RewardToken2 != nil {
				annualRewardPerLP2, _, ok = utils.ParseBigInt(value, idx)
				allOk = allOk && ok
			}
			if !allOk {
				log.Debug("refreshFarms", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			lastItemIDs[farmID] = itemID
			farm.AnnualRewardPerLP1 = utils.Denominate(annualRewardPerLP1, int(farm.RewardToken1.Decimals))
			if annualRewardPerLP2 != nil {
				farm.AnnualRewardPerLP2 = utils.Denominate(annualRewardPerLP2, int(farm.RewardToken2.Decimals))
			}
		}

		prefix = hex.EncodeToString([]byte("pool_user_last_update_timestamp"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshFarms", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			idx := 0
			farmID, idx, ok := utils.ParseUint32(bytes, 0)
			allOk := ok
			pubkey, _, ok := utils.ParsePubkey(bytes, idx)
			allOk = allOk && ok
			if !allOk {
				log.Debug("refreshFarms", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			farm := farms[farmID]
			address, _ := conv.Encode(pubkey)
			lastUpdate := big.NewInt(0).SetBytes(value)
			for _, farmer := range farm.Farmers {
				if farmer.Address == address {
					farmer.LastUpdate = lastUpdate.Int64()
					break
				}
			}
		}
	}

	return farms, nil
}

func (one *OneDex) GetStakes() (map[uint32]*Stake, error) {
	keys, err := one.stakingScAccount.GetAccountKeys("")
	if err != nil {
		return nil, err
	}

	stakes := make(map[uint32]*Stake)

	for key, value := range keys {
		prefix := hex.EncodeToString([]byte("pool_stake_token_id"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshStakes", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			stakeID, _, ok := utils.ParseUint32(bytes, 0)
			if !ok {
				log.Debug("refreshStakes", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			ticker := string(value)

			stake := stakes[stakeID]
			if stake == nil {
				stake = &Stake{
					ID:      stakeID,
					Stakers: make([]*Staker, 0),
				}
				stakes[stakeID] = stake
			}
			stake.Token, err = one.getToken(ticker)
			if err != nil {
				log.Debug("refreshStakes", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}
		}
	}

	conv, _ := pubkeyConverter.NewBech32PubkeyConverter(32, "erd")

	oneToken, err := one.getToken(OneToken)
	if err != nil {
		return nil, err
	}

	lastItemIDs := make(map[uint32]uint32)

	for key, value := range keys {
		prefix := hex.EncodeToString([]byte("pool_user_stake_amount"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshStakes", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			idx := 0
			stakeID, idx, ok := utils.ParseUint32(bytes, 0)
			allOk := ok
			pubkey, _, ok := utils.ParsePubkey(bytes, idx)
			allOk = allOk && ok
			if !allOk {
				log.Debug("refreshStakes", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			stake := stakes[stakeID]
			address, _ := conv.Encode(pubkey)
			iAmount := big.NewInt(0).SetBytes(value)
			fAmount := utils.Denominate(iAmount, int(stake.Token.Decimals))
			stake.Stakers = append(stake.Stakers, &Staker{
				Address: address,
				Amount:  fAmount,
			})
		}

		prefix = hex.EncodeToString([]byte("pool_total_stake_amount"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshStakes", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			stakeID, _, ok := utils.ParseUint32(bytes, 0)
			if !ok {
				log.Debug("refreshStakes", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			iTotalStake := big.NewInt(0).SetBytes(value)
			stake := stakes[stakeID]
			stake.TotalStake = utils.Denominate(iTotalStake, int(oneToken.Decimals))
		}

		prefix = hex.EncodeToString([]byte("pool_reward_deposit_amount"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshStakes", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			stakeID, _, ok := utils.ParseUint32(bytes, 0)
			if !ok {
				log.Debug("refreshStakes", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			iRewardPool := big.NewInt(0).SetBytes(value)
			stake := stakes[stakeID]
			stake.RewardPool = utils.Denominate(iRewardPool, int(oneToken.Decimals))
		}

		prefix = hex.EncodeToString([]byte("pool_reward_infos"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshStakes", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			stakeID, _, ok := utils.ParseUint32(bytes, 0)
			if !ok {
				log.Debug("refreshStakes", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			if !strings.HasPrefix(string(bytes[4:]), ".item") {
				continue
			}

			itemID, _, ok := utils.ParseUint32(bytes, 9)
			if itemID < lastItemIDs[stakeID] || !ok {
				continue
			}

			apr, _, ok := utils.ParseBigInt(value, 0)
			if !ok {
				log.Debug("refreshStakes", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			lastItemIDs[stakeID] = itemID
			stakes[stakeID].APR = float64(apr.Uint64()) / 100
		}
	}

	for key, value := range keys {
		prefix := hex.EncodeToString([]byte("pool_user_reward_amount"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshStakes", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			idx := 0
			stakeID, idx, ok := utils.ParseUint32(bytes, 0)
			allOk := ok
			pubkey, _, ok := utils.ParsePubkey(bytes, idx)
			allOk = allOk && ok
			if !allOk {
				log.Debug("refreshStakes", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			stake := stakes[stakeID]
			address, _ := conv.Encode(pubkey)
			iAmount := big.NewInt(0).SetBytes(value)
			fAmount := utils.Denominate(iAmount, int(stake.Token.Decimals))
			for _, staker := range stake.Stakers {
				if staker.Address == address {
					staker.Reward = fAmount
					break
				}
			}
		}

		prefix = hex.EncodeToString([]byte("pool_user_last_update_timestamp"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshStakes", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			idx := 0
			stakeID, idx, ok := utils.ParseUint32(bytes, 0)
			allOk := ok
			pubkey, _, ok := utils.ParsePubkey(bytes, idx)
			allOk = allOk && ok
			if !allOk {
				log.Debug("refreshStakes", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			stake := stakes[stakeID]
			address, _ := conv.Encode(pubkey)
			lastUpdate := big.NewInt(0).SetBytes(value)
			for _, staker := range stake.Stakers {
				if staker.Address == address {
					staker.LastUpdate = lastUpdate.Int64()
					break
				}
			}
		}
	}

	return stakes, nil
}

func (one *OneDex) GetLaunchpads() (map[uint32]*Launchpad, error) {
	keys, err := one.launchpadScAccount.GetAccountKeys("")
	if err != nil {
		return nil, err
	}

	launchpads := make(map[uint32]*Launchpad)
	conv, _ := pubkeyConverter.NewBech32PubkeyConverter(32, "erd")
	for key, value := range keys {
		prefix := hex.EncodeToString([]byte("project_is_lived"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			launchpadID, _, ok := utils.ParseUint32(bytes, 0)
			if !ok {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			isLive, _, ok := utils.ParseByte(value, 0)
			if !ok {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			launchpads[launchpadID] = &Launchpad{
				ID:     launchpadID,
				IsLive: isLive == 1,
				Buyers: make([]*LaunchpadBuyer, 0),
			}
		}
	}

	for key, value := range keys {
		prefix := hex.EncodeToString([]byte("project_create_time"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			launchpadID, _, ok := utils.ParseUint32(bytes, 0)
			if !ok {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			createTime := big.NewInt(0).SetBytes(value)
			launchpads[launchpadID].CreateTime = createTime.Int64()
		}

		prefix = hex.EncodeToString([]byte("project_presale_start_time"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			launchpadID, _, ok := utils.ParseUint32(bytes, 0)
			if !ok {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			presaleStart := big.NewInt(0).SetBytes(value)
			launchpads[launchpadID].StartTime = presaleStart.Int64()
		}

		prefix = hex.EncodeToString([]byte("project_presale_end_time"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			launchpadID, _, ok := utils.ParseUint32(bytes, 0)
			if !ok {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			presaleEnd := big.NewInt(0).SetBytes(value)
			launchpads[launchpadID].EndTime = presaleEnd.Int64()
		}

		prefix = hex.EncodeToString([]byte("project_description"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			launchpadID, _, ok := utils.ParseUint32(bytes, 0)
			if !ok {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			launchpads[launchpadID].Description = string(value)
		}

		prefix = hex.EncodeToString([]byte("project_social_telegram"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			launchpadID, _, ok := utils.ParseUint32(bytes, 0)
			if !ok {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			launchpads[launchpadID].Telegram = string(value)
		}

		prefix = hex.EncodeToString([]byte("project_social_twitter"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			launchpadID, _, ok := utils.ParseUint32(bytes, 0)
			if !ok {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			launchpads[launchpadID].Twitter = string(value)
		}

		prefix = hex.EncodeToString([]byte("project_social_website"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			launchpadID, _, ok := utils.ParseUint32(bytes, 0)
			if !ok {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			launchpads[launchpadID].Website = string(value)
		}

		prefix = hex.EncodeToString([]byte("project_presale_token_identifier"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			launchpadID, _, ok := utils.ParseUint32(bytes, 0)
			if !ok {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			launchpads[launchpadID].Token = string(value)
		}

		prefix = hex.EncodeToString([]byte("project_fund_token_identifier"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			launchpadID, _, ok := utils.ParseUint32(bytes, 0)
			if !ok {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			launchpads[launchpadID].FundToken = string(value)
		}
	}

	for key, value := range keys {
		prefix := hex.EncodeToString([]byte("project_hard_cap"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			launchpadID, _, ok := utils.ParseUint32(bytes, 0)
			if !ok {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			launchpad := launchpads[launchpadID]
			iHardCap := big.NewInt(0).SetBytes(value)
			tokenName := launchpad.FundToken
			if tokenName == "EGLD" {
				tokenName = utils.WEGLD
			}
			token, err := one.getToken(tokenName)
			if err != nil {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}
			launchpad.HardCap = utils.Denominate(iHardCap, int(token.Decimals))
		}

		prefix = hex.EncodeToString([]byte("project_presale_token_rate"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			launchpadID, _, ok := utils.ParseUint32(bytes, 0)
			if !ok {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			launchpad := launchpads[launchpadID]
			iTokenRate := big.NewInt(0).SetBytes(value)
			token, err := one.getToken(launchpad.Token)
			if err != nil {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}
			launchpad.Rate = utils.Denominate(iTokenRate, int(token.Decimals))
		}

		prefix = hex.EncodeToString([]byte("project_total_bought_amount"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			launchpadID, _, ok := utils.ParseUint32(bytes, 0)
			if !ok {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			launchpad := launchpads[launchpadID]
			iBoughtAmount := big.NewInt(0).SetBytes(value)
			token, err := one.getToken(launchpad.Token)
			if err != nil {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}
			launchpad.TotalBought = utils.Denominate(iBoughtAmount, int(token.Decimals))
		}

		prefix = hex.EncodeToString([]byte("project_total_fund_amount"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			launchpadID, _, ok := utils.ParseUint32(bytes, 0)
			if !ok {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			launchpad := launchpads[launchpadID]
			iFundAmount := big.NewInt(0).SetBytes(value)
			tokenName := launchpad.FundToken
			if tokenName == "EGLD" {
				tokenName = utils.WEGLD
			}
			token, err := one.getToken(tokenName)
			if err != nil {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}
			launchpad.FundAmount = utils.Denominate(iFundAmount, int(token.Decimals))
		}

		prefix = hex.EncodeToString([]byte("project_user_bought_amount"))
		if strings.HasPrefix(key, prefix) {
			bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
			if err != nil {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			idx := 0
			launchpadID, idx, ok := utils.ParseUint32(bytes, 0)
			allOk := ok
			pubkey, _, ok := utils.ParsePubkey(bytes, idx)
			allOk = allOk && ok
			if !allOk {
				log.Debug("refreshFarms", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}

			launchpad := launchpads[launchpadID]
			address, _ := conv.Encode(pubkey)
			iAmount := big.NewInt(0).SetBytes(value)
			token, err := one.getToken(launchpad.Token)
			if err != nil {
				log.Debug("refreshLaunchpads", "step", "parse keys", "error", "can not decode key", "key", key)
				continue
			}
			fAmount := utils.Denominate(iAmount, int(token.Decimals))
			launchpad.Buyers = append(launchpad.Buyers, &LaunchpadBuyer{
				Address: address,
				Amount:  fAmount,
			})
		}
	}

	return launchpads, nil
}

func (one *OneDex) getToken(ticker string) (*data.ESDT, error) {
	if one.refreshInterval == utils.NoRefresh {
		return one.mxTokens.GetTokenProperties(ticker)
	} else {
		return one.mxTokens.GetCachedTokenProperties(ticker)
	}
}

func (one *OneDex) GetCachedLiquidityPools() (map[uint32]*LiquidityPool, error) {
	if one.refreshInterval == utils.NoRefresh {
		return nil, utils.ErrRefreshIntervalNotSet
	}

	res := make(map[uint32]*LiquidityPool)
	one.liquidityPoolsMut.Lock()
	for k, v := range one.liquidityPools {
		res[k] = v
	}
	one.liquidityPoolsMut.Unlock()

	return res, nil
}

func (one *OneDex) GetCachedFarms() (map[uint32]*Farm, error) {
	if one.refreshInterval == utils.NoRefresh {
		return nil, utils.ErrRefreshIntervalNotSet
	}

	res := make(map[uint32]*Farm)
	one.farmsMut.Lock()
	for k, v := range one.farms {
		res[k] = v
	}
	one.farmsMut.Unlock()

	return res, nil
}

func (one *OneDex) GetCachedStakes() (map[uint32]*Stake, error) {
	if one.refreshInterval == utils.NoRefresh {
		return nil, utils.ErrRefreshIntervalNotSet
	}

	res := make(map[uint32]*Stake)
	one.stakesMut.Lock()
	for k, v := range one.stakes {
		res[k] = v
	}
	one.stakesMut.Unlock()

	return res, nil
}

func (one *OneDex) GetCachedLaunchpads() (map[uint32]*Launchpad, error) {
	if one.refreshInterval == utils.NoRefresh {
		return nil, utils.ErrRefreshIntervalNotSet
	}

	res := make(map[uint32]*Launchpad)
	one.launchpadsMut.Lock()
	for k, v := range one.launchpads {
		res[k] = v
	}
	one.launchpadsMut.Unlock()

	return res, nil
}

func (one *OneDex) GetCachedUserFarms(address string) []*UserFarm {
	userFarms := make([]*UserFarm, 0)

	one.farmsMut.Lock()
	for _, farm := range one.farms {
		for _, farmer := range farm.Farmers {
			if farmer.Address == address {
				days := float64(time.Now().Unix()-farmer.LastUpdate) / 86400
				reward1 := farmer.Amount * farm.AnnualRewardPerLP1 / 365 * days
				userFarm := &UserFarm{
					Farm:    farm,
					Amount:  farmer.Amount,
					Reward1: reward1,
				}
				userFarms = append(userFarms, userFarm)
				if farm.RewardToken2 != nil {
					reward2 := farmer.Amount * farm.AnnualRewardPerLP2 / 365 * days
					userFarm.Reward2 = reward2
				}
			}
		}
	}
	one.farmsMut.Unlock()

	return userFarms
}

func (one *OneDex) GetCachedUserStakes(address string) []*UserStake {
	userStakes := make([]*UserStake, 0)

	one.stakesMut.Lock()
	for _, stake := range one.stakes {
		for _, staker := range stake.Stakers {
			if staker.Address == address {
				days := float64(time.Now().Unix()-staker.LastUpdate) / 86400
				userStakes = append(userStakes, &UserStake{
					Token:  stake.Token,
					Amount: staker.Amount,
					Reward: staker.Amount * stake.APR / 100 * days / 365,
				})
			}
		}
	}
	one.stakesMut.Unlock()

	return userStakes
}

func (one *OneDex) GetCachedTokenPrice(ticker string, egldPrice float64) float64 {
	if ticker == utils.USDC || ticker == utils.BUSD || ticker == utils.USDT {
		return 1
	}

	if ticker == utils.WEGLD {
		return egldPrice
	}

	one.liquidityPoolsMut.Lock()
	defer one.liquidityPoolsMut.Unlock()

	for _, lp := range one.liquidityPools {
		if lp.Token1Reserve == 0 {
			continue
		}

		if lp.Token1.Ticker == ticker {
			price := lp.Token2Reserve / lp.Token1Reserve
			if lp.Token2.Ticker == utils.WEGLD {
				price *= egldPrice
			}
			if lp.Token2.Ticker == OneToken {
				price *= one.GetCachedTokenPrice(OneToken, egldPrice)
			}
			return price
		}

		if lp.LpToken != nil && lp.LpToken.Ticker == ticker {
			if lp.LpTokenSupply == 0 {
				continue
			}

			price := lp.Token2Reserve * 2 / lp.LpTokenSupply
			if lp.Token2.Ticker == utils.WEGLD {
				price *= egldPrice
			}
			if lp.Token2.Ticker == OneToken {
				price *= one.GetCachedTokenPrice(OneToken, egldPrice)
			}
			return price
		}
	}

	return 0
}
