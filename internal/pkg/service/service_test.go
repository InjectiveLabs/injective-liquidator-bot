package service

import (
	"context"
	"cosmossdk.io/math"
	"testing"

	exchangetypes "github.com/InjectiveLabs/sdk-go/chain/exchange/types"
	"github.com/InjectiveLabs/sdk-go/client/chain"
	"github.com/InjectiveLabs/sdk-go/client/exchange"
	derivativeExchangePB "github.com/InjectiveLabs/sdk-go/exchange/derivative_exchange_rpc/pb"
	spotExchangePB "github.com/InjectiveLabs/sdk-go/exchange/spot_exchange_rpc/pb"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	eth "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestLiquidatePositionMessageWhenNotUsingDelegatedAccount(t *testing.T) {
	granterPublicAddress := ""
	granteePublicAddress := "inj14au322k9munkmx5wrchz9q30juf5wjgz2cfqku"
	granteeSubaccountID := eth.HexToHash("0x00606da8ef76ca9c36616fa576d1c053bb0f7eb2000000000000000000000000")
	// granterSubaccountID := eth.HexToHash("0xbdaedec95d563fb05240d6e01821008454c24c36000000000000000000000000")
	granterSubaccountID := eth.HexToHash("")

	mockChain := LocalMockChainClient{}
	mockExchange := exchange.MockExchangeClient{}

	address, _ := types.AccAddressFromBech32(granteePublicAddress)
	mockChain.FromAddresses = append(mockChain.FromAddresses, address)

	var derivativeMarketInfos []*derivativeExchangePB.DerivativeMarketInfo
	btcUsdtDerivativeMarketInfo := createBTCUSDTDerivativeMarketInfo()

	derivativeMarketInfos = append(derivativeMarketInfos, btcUsdtDerivativeMarketInfo)
	mockExchange.SpotMarketsResponses = append(mockExchange.SpotMarketsResponses, &spotExchangePB.MarketsResponse{
		Markets: []*spotExchangePB.SpotMarketInfo{},
	})
	mockExchange.DerivativeMarketsResponses = append(mockExchange.DerivativeMarketsResponses, &derivativeExchangePB.MarketsResponse{
		Markets: derivativeMarketInfos,
	})

	ctx := context.Background()
	marketAssistant, err := chain.NewMarketsAssistantInitializedFromChain(ctx, &mockExchange)

	assert.NoError(t, err)

	liquidatorService := liquidatorSvc{
		chainClient:          &mockChain,
		exchangeClient:       &mockExchange,
		marketsAssistant:     marketAssistant,
		marketID:             btcUsdtDerivativeMarketInfo.MarketId,
		subaccountID:         granteeSubaccountID,
		granterPublicAddress: granterPublicAddress,
		granterSubaccountID:  granterSubaccountID,
		maxOrderAmount:       math.LegacyMaxSortableDec,
		maxOrderNotional:     math.LegacyMaxSortableDec,
	}

	market := marketAssistant.AllDerivativeMarkets()[btcUsdtDerivativeMarketInfo.MarketId]
	position := derivativeExchangePB.DerivativePosition{
		MarketId:     market.Id,
		SubaccountId: "positionSubaccountID",
		Direction:    "long",
		Quantity:     "1",
		MarkPrice:    "3400000000",
	}

	message := liquidatorService.createLiquidationMessage(&position, market)
	liquidationMessage := message.(*exchangetypes.MsgLiquidatePosition)

	assert.Equal(t, granteePublicAddress, liquidationMessage.Sender)
	assert.Equal(t, position.SubaccountId, liquidationMessage.SubaccountId)
	assert.Equal(t, market.Id, position.MarketId)
	assert.Equal(t, market.Id, liquidationMessage.Order.GetMarketId())
	assert.Equal(t, granteeSubaccountID.String(), liquidationMessage.Order.OrderInfo.SubaccountId)
	assert.Equal(t, granteePublicAddress, liquidationMessage.Order.OrderInfo.FeeRecipient)
	assert.Equal(t, math.LegacyMustNewDecFromStr(position.MarkPrice), liquidationMessage.Order.OrderInfo.Price)
	assert.Equal(t, math.LegacyMustNewDecFromStr(position.Quantity), liquidationMessage.Order.OrderInfo.Quantity)
	assert.NotEqual(t, "", liquidationMessage.Order.OrderInfo.GetCid())
	assert.Equal(t, exchangetypes.OrderType_BUY, liquidationMessage.Order.OrderType)
	assert.Equal(
		t,
		math.LegacyMustNewDecFromStr(position.Quantity).Mul(math.LegacyMustNewDecFromStr(position.MarkPrice)),
		liquidationMessage.Order.GetMargin(),
	)
}

func TestLiquidatePositionMessageWhenUsingDelegatedAccount(t *testing.T) {
	granterPublicAddress := "granterPublicAddress"
	granteePublicAddress := "inj14au322k9munkmx5wrchz9q30juf5wjgz2cfqku"
	granteeSubaccountID := eth.HexToHash("0x00606da8ef76ca9c36616fa576d1c053bb0f7eb2000000000000000000000000")
	granterSubaccountID := eth.HexToHash("0xbdaedec95d563fb05240d6e01821008454c24c36000000000000000000000000")

	mockChain := LocalMockChainClient{}
	mockExchange := exchange.MockExchangeClient{}

	address, _ := types.AccAddressFromBech32(granteePublicAddress)
	mockChain.FromAddresses = append(mockChain.FromAddresses, address)

	var derivativeMarketInfos []*derivativeExchangePB.DerivativeMarketInfo
	btcUsdtDerivativeMarketInfo := createBTCUSDTDerivativeMarketInfo()

	derivativeMarketInfos = append(derivativeMarketInfos, btcUsdtDerivativeMarketInfo)
	mockExchange.SpotMarketsResponses = append(mockExchange.SpotMarketsResponses, &spotExchangePB.MarketsResponse{
		Markets: []*spotExchangePB.SpotMarketInfo{},
	})
	mockExchange.DerivativeMarketsResponses = append(mockExchange.DerivativeMarketsResponses, &derivativeExchangePB.MarketsResponse{
		Markets: derivativeMarketInfos,
	})

	ctx := context.Background()
	marketAssistant, err := chain.NewMarketsAssistantInitializedFromChain(ctx, &mockExchange)

	assert.NoError(t, err)

	liquidatorService := liquidatorSvc{
		chainClient:          &mockChain,
		exchangeClient:       &mockExchange,
		marketsAssistant:     marketAssistant,
		marketID:             btcUsdtDerivativeMarketInfo.MarketId,
		subaccountID:         granteeSubaccountID,
		granterPublicAddress: granterPublicAddress,
		granterSubaccountID:  granterSubaccountID,
		maxOrderAmount:       math.LegacyMaxSortableDec,
		maxOrderNotional:     math.LegacyMaxSortableDec,
	}

	market := marketAssistant.AllDerivativeMarkets()[btcUsdtDerivativeMarketInfo.MarketId]
	position := derivativeExchangePB.DerivativePosition{
		MarketId:     market.Id,
		SubaccountId: "positionSubaccountID",
		Direction:    "long",
		Quantity:     "1",
		MarkPrice:    "3400000000",
	}

	message := liquidatorService.createLiquidationMessage(&position, market)
	execMessage := message.(*authz.MsgExec)

	assert.Equal(t, granteePublicAddress, execMessage.Grantee)

	liquidationMessage := exchangetypes.MsgLiquidatePosition{}
	unmarshallError := liquidationMessage.Unmarshal(execMessage.Msgs[0].GetValue())

	assert.NoError(t, unmarshallError)

	assert.Equal(t, granterPublicAddress, liquidationMessage.Sender)
	assert.Equal(t, position.SubaccountId, liquidationMessage.SubaccountId)
	assert.Equal(t, market.Id, position.MarketId)
	assert.Equal(t, market.Id, liquidationMessage.Order.GetMarketId())
	assert.Equal(t, granterSubaccountID.String(), liquidationMessage.Order.OrderInfo.SubaccountId)
	assert.Equal(t, granterPublicAddress, liquidationMessage.Order.OrderInfo.FeeRecipient)
	assert.Equal(t, math.LegacyMustNewDecFromStr(position.MarkPrice), liquidationMessage.Order.OrderInfo.Price)
	assert.Equal(t, math.LegacyMustNewDecFromStr(position.Quantity), liquidationMessage.Order.OrderInfo.Quantity)
	assert.NotEqual(t, "", liquidationMessage.Order.OrderInfo.GetCid())
	assert.Equal(t, exchangetypes.OrderType_BUY, liquidationMessage.Order.OrderType)
	assert.Equal(
		t,
		math.LegacyMustNewDecFromStr(position.Quantity).Mul(math.LegacyMustNewDecFromStr(position.MarkPrice)),
		liquidationMessage.Order.GetMargin(),
	)
}
