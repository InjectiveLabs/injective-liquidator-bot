package service

import (
	exchangetypes "github.com/InjectiveLabs/sdk-go/chain/exchange/types"
	"github.com/InjectiveLabs/sdk-go/client/chain"
	derivativeExchangePB "github.com/InjectiveLabs/sdk-go/exchange/derivative_exchange_rpc/pb"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	eth "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

func createUSDTPerpTokenMeta() derivativeExchangePB.TokenMeta {
	return derivativeExchangePB.TokenMeta{
		Name:      "Tether",
		Address:   "0xdAC17F958D2ee523a2206206994597C13D831ec7",
		Symbol:    "USDTPerp",
		Logo:      "https://static.alchemyapi.io/images/assets/825.png",
		Decimals:  6,
		UpdatedAt: 1683929869866,
	}
}

func createBTCUSDTDerivativeMarketInfo() *derivativeExchangePB.DerivativeMarketInfo {
	usdtPerpTokenMeta := createUSDTPerpTokenMeta()

	perpetualMarketInfo := derivativeExchangePB.PerpetualMarketInfo{
		HourlyFundingRateCap: "0.0000625",
		HourlyInterestRate:   "0.00000416666",
		NextFundingTimestamp: 1684764000,
		FundingInterval:      3600,
	}

	perpetualmarketFunding := derivativeExchangePB.PerpetualMarketFunding{
		CumulativeFunding: "6880500093.266083891331674194",
		CumulativePrice:   "-0.952642601240470199",
		LastTimestamp:     1684763442,
	}

	marketInfo := derivativeExchangePB.DerivativeMarketInfo{
		MarketId:               "0x4ca0f92fc28be0c9761326016b5a1a2177dd6375558365116b5bdda9abc229ce",
		MarketStatus:           "active",
		Ticker:                 "BTC/USDT PERP",
		OracleBase:             "BTC",
		OracleQuote:            "USDT",
		OracleType:             "bandibc",
		OracleScaleFactor:      6,
		InitialMarginRatio:     "0.095",
		MaintenanceMarginRatio: "0.025",
		QuoteDenom:             "peggy0xdAC17F958D2ee523a2206206994597C13D831ec7",
		QuoteTokenMeta:         &usdtPerpTokenMeta,
		MakerFeeRate:           "-0.0001",
		TakerFeeRate:           "0.001",
		ServiceProviderFee:     "0.4",
		IsPerpetual:            true,
		MinPriceTickSize:       "1000000",
		MinQuantityTickSize:    "0.0001",
		PerpetualMarketInfo:    &perpetualMarketInfo,
		PerpetualMarketFunding: &perpetualmarketFunding,
	}

	return &marketInfo
}

type LocalMockChainClient struct {
	chain.MockChainClient
	FromAddresses       []sdk.AccAddress
	BroadcastedMessages []sdk.Msg
}

func (c *LocalMockChainClient) FromAddress() sdk.AccAddress {
	address := c.FromAddresses[0]
	c.FromAddresses = c.FromAddresses[1:len(c.FromAddresses)]
	return address
}

func (c *LocalMockChainClient) CreateDerivativeOrder(defaultSubaccountID eth.Hash, d *chain.DerivativeOrderData, marketAssistant chain.MarketsAssistant) *exchangetypes.DerivativeOrder {
	market, isPresent := marketAssistant.AllDerivativeMarkets()[d.MarketId]
	if !isPresent {
		panic(errors.Errorf("Invalid derivative market id %s", d.MarketId))
	}

	orderSize := market.QuantityToChainFormat(d.Quantity)
	orderPrice := market.PriceToChainFormat(d.Price)
	orderMargin := sdk.MustNewDecFromStr("0")

	if !d.IsReduceOnly {
		orderMargin = market.CalculateMarginInChainFormat(d.Quantity, d.Price, d.Leverage)
	}

	return &exchangetypes.DerivativeOrder{
		MarketId:  d.MarketId,
		OrderType: d.OrderType,
		Margin:    orderMargin,
		OrderInfo: exchangetypes.OrderInfo{
			SubaccountId: defaultSubaccountID.Hex(),
			FeeRecipient: d.FeeRecipient,
			Price:        orderPrice,
			Quantity:     orderSize,
			Cid:          d.Cid,
		},
	}
}

func (c *LocalMockChainClient) SyncBroadcastMsg(msgs ...sdk.Msg) (*tx.BroadcastTxResponse, error) {
	c.BroadcastedMessages = append(c.BroadcastedMessages, msgs...)
	return &tx.BroadcastTxResponse{}, nil
}
