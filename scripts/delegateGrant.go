package main

import (
	"fmt"
	"os"
	"time"

	"github.com/InjectiveLabs/sdk-go/client"
	chainclient "github.com/InjectiveLabs/sdk-go/client/chain"
	"github.com/InjectiveLabs/sdk-go/client/common"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
)

// Configure the granter account private key
var granterPrivateKey = ""

// Configure the grantee account public injective address
var granteePublicAddress = ""

// Configure the time frame the grant should remain valid (you will have to renew it after this time)
var expireIn = time.Now().AddDate(1, 0, 0) // years months days

// Configure the network to execute the script in: mainnet or testnet
var networkName = "mainnet"

func main() {
	network := common.LoadNetwork(networkName, "lb")
	tmClient, err := rpchttp.New(network.TmEndpoint, "/websocket")
	if err != nil {
		panic(err)
	}

	senderAddress, cosmosKeyring, err := chainclient.InitCosmosKeyring(
		os.Getenv("HOME")+"/.injectived",
		"",
		"",
		"",
		"",
		granterPrivateKey,
		false,
	)

	if err != nil {
		panic(err)
	}

	clientCtx, err := chainclient.NewClientContext(
		network.ChainId,
		senderAddress.String(),
		cosmosKeyring,
	)

	if err != nil {
		panic(err)
	}

	clientCtx = clientCtx.WithNodeURI(network.TmEndpoint).WithClient(tmClient)

	chainClient, err := chainclient.NewChainClient(
		clientCtx,
		network,
		common.OptionGasPrices(client.DefaultGasPriceWithDenom),
	)

	if err != nil {
		panic(err)
	}

	granter := senderAddress.String()

	msg := chainClient.BuildGenericAuthz(
		granter,
		granteePublicAddress,
		"/injective.exchange.v1beta1.MsgLiquidatePosition",
		expireIn,
	)

	//AsyncBroadcastMsg, SyncBroadcastMsg, QueueBroadcastMsg
	response, err := chainClient.SyncBroadcastMsg(msg)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Broadcast result: %s\n", response)
}
