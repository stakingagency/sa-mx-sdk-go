# sa-mx-sdk-go
MultiversX Golang SDK by Staking Agency


This kit makes building on the MultiversX blockchain a walk in the park.

1. **[Accounts](https://github.com/stakingagency/sa-mx-sdk-go/tree/master/accounts)**
   - `GetAccountKeys` - retrieves all the keys associated with an account. Useful for reading a SC's storage.
   - `DNSResolve` - pass a herotag a get the associated address
   - `GetEgldBalance` - easily retrieve an account's eGLD balance
   - `GetTokensBalances` - same as above, but for an account's ESDTs
   - `GetTokenDecimals` - get an ESDT's number of decimals

   *Callbacks:* `EgldBalanceChanged` `TokenBalanceChanged`

2. **[Exchanges](https://github.com/stakingagency/sa-mx-sdk-go/tree/master/exchanges)**
   + [xExchange](https://github.com/stakingagency/sa-mx-sdk-go/tree/master/exchanges/xexchange)
      - `GetDexPairs` - reads all trading pairs listed on xExchange
      - `GetPairByTickers` - specify the pair's tickers and get all the pair details
      - `GetPairByContractAddress` - get a pair's details by its contract address

      *Callbacks:* `NewPair` `PairStateChanged` `DexStateChanged`

   + [OneDex](https://github.com/stakingagency/sa-mx-sdk-go/tree/master/exchanges/onedex)
      - `GetLiquidityPools` - get all liquidity pools
      - `GetFarms` - get all farms (simple and dual)
      - `GetStakes` - get all coins stakes
      - `GetLaunchpads` - get all launchpads

      *Callbacks:* `NewPair` `PairStateChanged` `NewStake` `NewFarm` `NewDualFarm` `NewLaunchpad` `LaunchpadEnded` `AnnualRewardChanged` `StakeAprChanged`

3. **[Network](https://github.com/stakingagency/sa-mx-sdk-go/tree/master/network)**
   - `SearchIndexer` - a powerful function to retrieve data from an ES indexed with MultiversX data (retrieves more than 10,000 records)
   - `GetTxInfo` - gets a transaction's details from ES
   - `GetTxLogs` - gets a transaction's logs from ES
   - `GetTxOperations` - gets a transaction's operations from ES
   - `GetTxResult` - after sending a tx, call this function to wait for the tx's result and get a detailed error if it fails
   - `GetNetworkConfig` - retrieves the network configuration from the proxy
   - `SendTransaction` - sends a tx with customizable gas limit, data field, nonce
   - `SendEsdtTransaction` - generates and sends an ESDT transfer

4. **[Staking](https://github.com/stakingagency/sa-mx-sdk-go/tree/master/staking)**
   - `GetAllProvidersAddresses` - returns all the staking providers contracts addresses
   - `GetMetaData` - get the name, website and identity for a specific provider
   - `GetUserStakeInfo` - gets a user staking details for a specific provider
   - `GetProviderConfig` - gets a provider configuration details
   - `GetProvidesConfigs` - gets all providers configurations

   *Callbacks:* `ProviderOwnerChanged` `ProviderNameChanged` `ProviderFeeChanged` `ProviderCapChanged` `ProviderSpaceAvailable` `NewProvider` `ProviderClosed`

5. **[Tokens](https://github.com/stakingagency/sa-mx-sdk-go/tree/master/tokens)**
   - `GetTokens` - retrieves all issued tokens (takes a while ...)
   - `IsTokenPaused` - returns true if the specified ESDT is paused
   - `GetTokenProperties` - retrieves an ESDT's properties, including all mint info

   *Callbacks:* `NewTokenIssued` `TokenStateChanged` `TokenSupplyChanged`

6. **[telegramBot](https://github.com/stakingagency/sa-mx-sdk-go/tree/master/telegramBot)**
   - `SendMessage` - sends a message to the specified user ID (can be a chat ID as well)
   - `SendFormattedMessage` - same as above, but you can specify the text format (markdown or html)

   *Callbacks:* `PrivateCommandReceived` `PublicCommandReceived` `PrivateMessageReceived` `PublicMessageReceived` `PrivateReplyReceived` `CallbackReceived`



**[ABI2GO](https://github.com/stakingagency/sa-mx-sdk-go/tree/master/abi2go)**
This is a very useful tool (still beta though) that generates Go language bindings to a MultiversX SC.


**[Examples](https://github.com/stakingagency/sa-mx-sdk-go/tree/master/examples)**
Here you will find examples for each of the packages presented above.
