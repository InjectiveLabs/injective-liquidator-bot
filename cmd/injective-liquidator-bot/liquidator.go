package main

import (
	"context"
	"fmt"
	"github.com/InjectiveLabs/sdk-go/client"
	"github.com/InjectiveLabs/sdk-go/client/common"
	"os"
	"time"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	cli "github.com/jawher/mow.cli"
	"github.com/xlab/closer"
	log "github.com/xlab/suplog"

	chainclient "github.com/InjectiveLabs/sdk-go/client/chain"
	sdkCommon "github.com/InjectiveLabs/sdk-go/client/common"
	exchangeclient "github.com/InjectiveLabs/sdk-go/client/exchange"

	"github.com/InjectiveLabs/injective-liquidator-bot/liquidator"
)

// liquidatorCmd action runs the service
//
// $ injective-liquidator-bot start
func liquidatorCmd(cmd *cli.Cmd) {
	// orchestrator-specific CLI options
	var (
		// Network params
		networkName             *string
		chainID                 *string
		lcdEndpoint             *string
		tendermintEndpoint      *string
		chainGrpcEndpoint       *string
		chainStreamGrpcEndpoint *string
		exchangeGrpcEndpoint    *string
		explorerGrpcEndpoint    *string

		// Cosmos Key Management
		cosmosKeyringDir     *string
		cosmosKeyringAppName *string
		cosmosKeyringBackend *string

		cosmosKeyFrom       *string
		cosmosKeyPassphrase *string
		cosmosPrivKey       *string
		cosmosUseLedger     *bool

		// Metrics
		statsdAgent    *string
		statsdPrefix   *string
		statsdAddr     *string
		statsdStuckDur *string
		statsdMocking  *string
		statsdDisabled *string

		//Liquidation
		subaccountIndex *int
		marketID        *string
	)

	initNetworkOptions(
		cmd,
		&networkName,
		&chainID,
		&lcdEndpoint,
		&tendermintEndpoint,
		&chainGrpcEndpoint,
		&chainStreamGrpcEndpoint,
		&exchangeGrpcEndpoint,
		&explorerGrpcEndpoint,
	)

	initCosmosKeyOptions(
		cmd,
		&cosmosKeyringDir,
		&cosmosKeyringAppName,
		&cosmosKeyringBackend,
		&cosmosKeyFrom,
		&cosmosKeyPassphrase,
		&cosmosPrivKey,
		&cosmosUseLedger,
	)

	initStatsdOptions(
		cmd,
		&statsdAgent,
		&statsdPrefix,
		&statsdAddr,
		&statsdStuckDur,
		&statsdMocking,
		&statsdDisabled,
	)

	initLiquidationOptions(
		cmd,
		&subaccountIndex,
		&marketID,
	)

	cmd.Action = func() {
		// ensure a clean exit
		defer closer.Close()

		startMetricsGathering(
			statsdAgent,
			statsdPrefix,
			statsdAddr,
			statsdStuckDur,
			statsdMocking,
			statsdDisabled,
		)

		if *cosmosUseLedger {
			log.Fatalln("cannot really use Ledger for liquidator service loop, since signatures must be realtime")
		}

		senderAddress, cosmosKeyring, err := chainclient.InitCosmosKeyring(
			*cosmosKeyringDir,
			*cosmosKeyringAppName,
			*cosmosKeyringBackend,
			*cosmosKeyFrom,
			*cosmosKeyPassphrase,
			*cosmosPrivKey,
			*cosmosUseLedger,
		)
		if err != nil {
			log.WithError(err).Fatalln("failed to init Cosmos keyring")
		}

		log.Infoln("Using Cosmos Sender", senderAddress.String())

		network, err := createNetwork(
			*networkName,
			*chainID,
			*lcdEndpoint,
			*tendermintEndpoint,
			*chainGrpcEndpoint,
			*chainStreamGrpcEndpoint,
			*exchangeGrpcEndpoint,
			*explorerGrpcEndpoint,
		)
		if err != nil {
			log.WithError(err).Fatalln("failed to configure the network")
		}

		clientCtx, err := chainclient.NewClientContext(network.ChainId, senderAddress.String(), cosmosKeyring)
		if err != nil {
			log.WithError(err).Fatalln("failed to initialize cosmos client context")
		}

		tmClient, err := rpchttp.New(network.TmEndpoint, "/websocket")
		if err != nil {
			log.WithError(err).Fatalln("failed to connect to tendermint RPC")
		}
		clientCtx = clientCtx.WithNodeURI(network.TmEndpoint).WithClient(tmClient).WithFromAddress(senderAddress)

		daemonClient, err := chainclient.NewChainClient(
			clientCtx,
			network,
			common.OptionGasPrices(client.DefaultGasPriceWithDenom),
		)
		if err != nil {
			log.WithError(err).Fatalln("failed to connect chain client, is injectived running?")
		}
		closer.Bind(func() {
			daemonClient.Close()
		})

		exchangeClient, err := exchangeclient.NewExchangeClient(network)
		if err != nil {
			log.WithError(err).Fatalln("failed to connect exchange client, is indexer running?")
		}
		closer.Bind(func() {
			exchangeClient.Close()
		})

		log.Infoln("Waiting for GRPC services")
		time.Sleep(1 * time.Second)

		daemonWaitCtx, cancelWait := context.WithTimeout(context.Background(), time.Minute)
		daemonConn := daemonClient.QueryClient()
		waitForService(daemonWaitCtx, daemonConn)
		cancelWait()

		exchangeWaitCtx, cancelWait := context.WithTimeout(context.Background(), time.Minute)
		exchangeConn := exchangeClient.QueryClient()
		waitForService(exchangeWaitCtx, exchangeConn)
		cancelWait()

		marketsAssistant, err := chainclient.NewMarketsAssistantInitializedFromChain(context.Background(), exchangeClient)
		if err != nil {
			log.WithError(err).Fatalln("failed to initialize the markets assistant")
		}

		subaccountID := daemonClient.Subaccount(senderAddress, *subaccountIndex)

		svc := liquidator.NewService(
			daemonClient,
			exchangeClient,
			marketsAssistant,
			*marketID,
			subaccountID,
		)
		closer.Bind(func() {
			svc.Close()
		})

		go func() {
			if err := svc.Start(); err != nil {
				log.Errorln(err)

				// signal there that the app failed
				os.Exit(1)
			}
		}()

		closer.Hold()
	}
}

func createNetwork(
	networkName string,
	chainID string,
	lcdEndpoint string,
	tendermintEndpoint string,
	chainGrpcEndpoint string,
	chainStreamGrpcEndpoint string,
	exchangeGrpcEndpoint string,
	explorerGrpcEndpoint string,
) (sdkCommon.Network, error) {
	var network = sdkCommon.NewNetwork()
	var err error

	if networkName == "mainnet" || networkName == "testnet" {
		network = sdkCommon.LoadNetwork(networkName, "lb")
	} else {
		if networkName == "custom" {
			network.LcdEndpoint = lcdEndpoint
			network.TmEndpoint = tendermintEndpoint
			network.ChainGrpcEndpoint = chainGrpcEndpoint
			network.ChainStreamGrpcEndpoint = chainStreamGrpcEndpoint
			network.ExchangeGrpcEndpoint = exchangeGrpcEndpoint
			network.ExplorerGrpcEndpoint = explorerGrpcEndpoint
			network.ChainId = chainID
			network.Fee_denom = "inj"
		} else {
			err = fmt.Errorf("network name %s is not valid", networkName)
		}
	}

	return network, err
}
