package main

import cli "github.com/jawher/mow.cli"

// initGlobalOptions defines some global CLI options, that are useful for most parts of the app.
// Before adding option to there, consider moving it into the actual Cmd.
func initGlobalOptions(
	envName **string,
	appLogLevel **string,
	svcWaitTimeout **string,
) {
	*envName = app.String(cli.StringOpt{
		Name:   "e env",
		Desc:   "The environment name this app runs in. Used for metrics and error reporting.",
		EnvVar: "LIQUIDATOR_ENV",
		Value:  "local",
	})

	*appLogLevel = app.String(cli.StringOpt{
		Name:   "l log-level",
		Desc:   "Available levels: error, warn, info, debug.",
		EnvVar: "LIQUIDATOR_LOG_LEVEL",
		Value:  "info",
	})

	*svcWaitTimeout = app.String(cli.StringOpt{
		Name:   "svc-wait-timeout",
		Desc:   "Standard wait timeout for external services (e.g. Cosmos daemon GRPC connection)",
		EnvVar: "LIQUIDATOR_SERVICE_WAIT_TIMEOUT",
		Value:  "1m",
	})
}

func initNetworkOptions(
	cmd *cli.Cmd,
	networkName **string,
	chainID **string,
	lcdEndpoint **string,
	tendermintEndpoint **string,
	chainGrpcEndpoint **string,
	chainStreamGrpcEndpoint **string,
	exchangeGrpcEndpoint **string,
	explorerGrpcEndpoint **string,
) {
	*networkName = cmd.String(cli.StringOpt{
		Name:   "network-name",
		Desc:   "Network to connect to (mainnet, testnet, custom)",
		EnvVar: "LIQUIDATOR_NETWORK_NAME",
		Value:  "mainnet",
	})

	*chainID = cmd.String(cli.StringOpt{
		Name:   "cosmos-chain-id",
		Desc:   "Specify Chain ID for custom network",
		EnvVar: "LIQUIDATOR_CHAIN_ID",
		Value:  "injective-1",
	})

	*lcdEndpoint = cmd.String(cli.StringOpt{
		Name:   "lcd-endpoint",
		Desc:   "LCD endpoint for custom network",
		EnvVar: "LIQUIDATOR_LCD_ENDPOINT",
		Value:  "http://localhost:10337",
	})

	*tendermintEndpoint = cmd.String(cli.StringOpt{
		Name:   "tendermint-endpoint",
		Desc:   "Tendermint endpoint for custom network",
		EnvVar: "LIQUIDATOR_TENDERMINT_ENDPOINT",
		Value:  "http://localhost:26657",
	})

	*chainGrpcEndpoint = cmd.String(cli.StringOpt{
		Name:   "chain-grpc-endpoint",
		Desc:   "Chain GRPC endpoint for custom network",
		EnvVar: "LIQUIDATOR_CHAIN_GRPC_ENDPOINT",
		Value:  "tcp://localhost:9900",
	})

	*chainStreamGrpcEndpoint = cmd.String(cli.StringOpt{
		Name:   "chain-stream-grpc-endpoint",
		Desc:   "ChainStream GRPC endpoint for custom network",
		EnvVar: "LIQUIDATOR_CHAIN_STREAM_GRPC_ENDPOINT",
		Value:  "tcp://localhost:9999",
	})

	*exchangeGrpcEndpoint = cmd.String(cli.StringOpt{
		Name:   "exchange-grpc-endpoint",
		Desc:   "Exchange GRPC endpoint for custom network",
		EnvVar: "LIQUIDATOR_EXCHANGE_GRPC_ENDPOINT",
		Value:  "tcp://localhost:9910",
	})

	*explorerGrpcEndpoint = cmd.String(cli.StringOpt{
		Name:   "explorer-grpc-endpoint",
		Desc:   "Explorer GRPC endpoint for custom network",
		EnvVar: "LIQUIDATOR_EXPLORER_GRPC_ENDPOINT",
		Value:  "tcp://localhost:9911",
	})
}

func initCosmosKeyOptions(
	cmd *cli.Cmd,
	cosmosKeyringDir **string,
	cosmosKeyringAppName **string,
	cosmosKeyringBackend **string,
	cosmosKeyFrom **string,
	cosmosKeyPassphrase **string,
	cosmosPrivKey **string,
	cosmosUseLedger **bool,
) {
	*cosmosKeyringBackend = cmd.String(cli.StringOpt{
		Name:   "cosmos-keyring",
		Desc:   "Specify Cosmos keyring backend (os|file|kwallet|pass|test)",
		EnvVar: "LIQUIDATOR_COSMOS_KEYRING",
		Value:  "file",
	})

	*cosmosKeyringDir = cmd.String(cli.StringOpt{
		Name:   "cosmos-keyring-dir",
		Desc:   "Specify Cosmos keyring dir, if using file keyring.",
		EnvVar: "LIQUIDATOR_COSMOS_KEYRING_DIR",
		Value:  "",
	})

	*cosmosKeyringAppName = cmd.String(cli.StringOpt{
		Name:   "cosmos-keyring-app",
		Desc:   "Specify Cosmos keyring app name.",
		EnvVar: "LIQUIDATOR_COSMOS_KEYRING_APP",
		Value:  "injectived",
	})

	*cosmosKeyFrom = cmd.String(cli.StringOpt{
		Name:   "cosmos-from",
		Desc:   "Specify the Cosmos validator key name or address. If specified, must exist in keyring, ledger or match the privkey.",
		EnvVar: "LIQUIDATOR_COSMOS_FROM",
	})

	*cosmosKeyPassphrase = cmd.String(cli.StringOpt{
		Name:   "cosmos-from-passphrase",
		Desc:   "Specify keyring passphrase, otherwise Stdin will be used.",
		EnvVar: "LIQUIDATOR_COSMOS_FROM_PASSPHRASE",
	})

	*cosmosPrivKey = cmd.String(cli.StringOpt{
		Name:   "cosmos-pk",
		Desc:   "Provide a raw Cosmos account private key of the validator in hex. USE FOR TESTING ONLY!",
		EnvVar: "LIQUIDATOR_COSMOS_PK",
	})

	*cosmosUseLedger = cmd.Bool(cli.BoolOpt{
		Name:   "cosmos-use-ledger",
		Desc:   "Use the Cosmos app on hardware ledger to sign transactions.",
		EnvVar: "LIQUIDATOR_COSMOS_USE_LEDGER",
		Value:  false,
	})
}

// initStatsdOptions sets options for StatsD metrics.
func initStatsdOptions(
	cmd *cli.Cmd,
	statsdAgent **string,
	statsdPrefix **string,
	statsdAddr **string,
	statsdStuckDur **string,
	statsdMocking **string,
	statsdDisabled **string,
) {
	*statsdAgent = cmd.String(cli.StringOpt{
		Name:   "statsd-agent",
		Desc:   "Specify StatsD agent.",
		EnvVar: "LIQUIDATOR_STATSD_AGENT",
		Value:  "telegraf",
	})

	*statsdPrefix = cmd.String(cli.StringOpt{
		Name:   "statsd-prefix",
		Desc:   "Specify StatsD compatible metrics prefix.",
		EnvVar: "LIQUIDATOR_STATSD_PREFIX",
		Value:  "liquidator",
	})

	*statsdAddr = cmd.String(cli.StringOpt{
		Name:   "statsd-addr",
		Desc:   "UDP address of a StatsD compatible metrics aggregator.",
		EnvVar: "LIQUIDATOR_STATSD_ADDR",
		Value:  "localhost:8125",
	})

	*statsdStuckDur = cmd.String(cli.StringOpt{
		Name:   "statsd-stuck-func",
		Desc:   "Sets a duration to consider a function to be stuck (e.g. in deadlock).",
		EnvVar: "LIQUIDATOR_STATSD_STUCK_DUR",
		Value:  "5m",
	})

	*statsdMocking = cmd.String(cli.StringOpt{
		Name:   "statsd-mocking",
		Desc:   "If enabled replaces statsd client with a mock one that simply logs values.",
		EnvVar: "LIQUIDATOR_STATSD_MOCKING",
		Value:  "false",
	})

	*statsdDisabled = cmd.String(cli.StringOpt{
		Name:   "statsd-disabled",
		Desc:   "Force disabling statsd reporting completely.",
		EnvVar: "LIQUIDATOR_STATSD_DISABLED",
		Value:  "true",
	})
}

func initLiquidationOptions(
	cmd *cli.Cmd,
	subaccountIndex **int,
	marketID **string,
) {
	*subaccountIndex = cmd.Int(cli.IntOpt{
		Name:   "subaccount-index",
		Desc:   "Subaccount number to use to create the liquidation orders",
		EnvVar: "LIQUIDATOR_SUBACCOUNT_INDEX",
		Value:  0,
	})

	*marketID = cmd.String(cli.StringOpt{
		Name:   "market-id",
		Desc:   "Market ID of the market to check liquidations for",
		EnvVar: "LIQUIDATOR_MARKET_ID",
		Value:  "",
	})
}
