# Injective Liquidator Bot


The Injective Liquidator Bot is a simple liquidation bot for executing liquidations on derivative markets on Injective. 
This bot operates by continuously monitoring for positions eligible for liquidation using the `LiquidablePositions` API. Upon identifying a liquidable position, the bot proceeds to initiate the liquidation process by dispatching a `MsgLiquidatedPosition` to the chain which executes the liquidation while also simultaneously creating the necessary order to facilitate an efficient liquidation.

The liquidation order is created with:
- The liquidable position's quantity
- The liquidable position's mark price
- 1x leverage
- A random UUID CID

Once running the bot activates the liquidable positions lookup process every 10 seconds.

## Running the bot

The bot can be started using the `restart.sh` script.

Additionally, the user can build an executable using the `make build` command. This will generate an executable called `injective-labs-liquidator`

### Bot commands

| Command | Description                                   |
|---------|-----------------------------------------------|
| start   | Start the bot to execute liquidable positions |
| version | Show the bot version information              |

### Configuration

The configuration is done with a `.env` file. You can create a copy of the `.env.example` file, and complete all the required information.

If the created configuration env file has a name different than `.env` you need to use the parameter `-e` or `-env` to specify the file name (this allows the user to prepare different configurations for different markets)

**General Configuration Options**

| Option                                | Description                                                                                                                                           |
|---------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------|
| LIQUIDATOR_SUBACCOUNT_INDEX           | The number of the subaccount the bot will use to send the liquidation requests to the chain (the account is determined by the configured credentials) |
| LIQUIDATOR_MARKET_ID                  | ID of the market the bot will use to find liquidable positions and execute the liquidations                                                           |


**Network Configuration options**

| Option                                | Description                                                                                                                                                               |
|---------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| LIQUIDATOR_NETWORK_NAME               | Network the bot will connect to. Possible balues are `mainnet`, `testnet` or `custom`. The next values are only required when using `custom` to connect to a private node |
| LIQUIDATOR_CHAIN_ID                   | The network chain id                                                                                                                                                      |
| LIQUIDATOR_LCD_ENDPOINT               | Network LDC endpoint                                                                                                                                                      |
| LIQUIDATOR_TENDERMINT_ENDPOINT        | Network Tendermint endpoint                                                                                                                                               |
| LIQUIDATOR_CHAIN_GRPC_ENDPOINT        | Network Chain GRPC endpoint                                                                                                                                               |
| LIQUIDATOR_CHAIN_STREAM_GRPC_ENDPOINT | Network Chain Stream GRPC endpoint                                                                                                                                        |
| LIQUIDATOR_EXCHANGE_GRPC_ENDPOINT     | Network Indexer Exchange GRPC endpoint                                                                                                                                    |
| LIQUIDATOR_EXPLORER_GRPC_ENDPOINT     | Network Indexer Explorer GRPC endpoint                                                                                                                                    |

When using a `custom` network you can check the endpoints in your private node's configuration (in general the values in the .env.example file are correct).

**Credentials Configuration Options**

The following configuration options can be specified for the bot to use a Cosmos Keyring to obtain the account private key (the user can provide directly the account private key using only the LIQUIDATOR_COSMOS_PK option)

- LIQUIDATOR_COSMOS_KEYRING
- LIQUIDATOR_COSMOS_KEYRING_DIR
- LIQUIDATOR_COSMOS_KEYRING_APP
- LIQUIDATOR_COSMOS_FROM
- LIQUIDATOR_COSMOS_FROM_PASSPHRASE
- LIQUIDATOR_COSMOS_PK
- LIQUIDATOR_COSMOS_USE_LEDGER