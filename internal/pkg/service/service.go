package service

import (
	"context"
	"cosmossdk.io/math"
	"runtime/debug"
	"time"

	"github.com/InjectiveLabs/sdk-go/client/core"
	"github.com/InjectiveLabs/sdk-go/client/exchange"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/pkg/errors"
	log "github.com/xlab/suplog"

	"github.com/InjectiveLabs/metrics"
	exchangetypes "github.com/InjectiveLabs/sdk-go/chain/exchange/types"
	chainclient "github.com/InjectiveLabs/sdk-go/client/chain"
	derivativeExchangePB "github.com/InjectiveLabs/sdk-go/exchange/derivative_exchange_rpc/pb"
	metaPB "github.com/InjectiveLabs/sdk-go/exchange/meta_rpc/pb"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

type Service interface {
	Start() error
	Close()
}

type liquidatorSvc struct {
	chainClient          chainclient.ChainClient
	exchangeClient       exchange.ExchangeClient
	marketsAssistant     chainclient.MarketsAssistant
	marketID             string
	subaccountID         common.Hash
	granterPublicAddress string
	granterSubaccountID  common.Hash
	maxOrderAmount       math.LegacyDec
	maxOrderNotional     math.LegacyDec

	logger  log.Logger
	svcTags metrics.Tags
}

func NewService(
	chainClient chainclient.ChainClient,
	exchangeClient exchange.ExchangeClient,
	marketsAssistant chainclient.MarketsAssistant,
	marketID string,
	subaccountID common.Hash,
	granterPublicAddress string,
	granterSubaccountID common.Hash,
	maxOrderAmount math.LegacyDec,
	maxOrderNotional math.LegacyDec,

) Service {
	return &liquidatorSvc{
		logger: log.WithField("svc", "liquidator"),
		svcTags: metrics.Tags{
			"svc": "liquidator_bot",
		},

		chainClient:          chainClient,
		exchangeClient:       exchangeClient,
		marketsAssistant:     marketsAssistant,
		marketID:             marketID,
		subaccountID:         subaccountID,
		granterPublicAddress: granterPublicAddress,
		granterSubaccountID:  granterSubaccountID,
		maxOrderAmount:       maxOrderAmount,
		maxOrderNotional:     maxOrderNotional,
	}
}

func (s *liquidatorSvc) Start() (err error) {
	defer s.panicRecover(&err)

	s.logger.Infoln("Service starts")

	// main bot loop

	ctx := context.Background()
	resp, err := s.exchangeClient.GetVersion(ctx, &metaPB.VersionRequest{})
	s.logger.Infof("Connected to Exchange API %s (build %s)", resp.Version, resp.Build["BuildDate"])

	for {
		req := derivativeExchangePB.LiquidablePositionsRequest{
			MarketId: s.marketID,
		}
		resp, err := s.exchangeClient.GetDerivativeLiquidablePositions(ctx, &req)
		metrics.ReportClosureFuncCall("LiquidablePositions", s.svcTags)

		if err != nil {
			metrics.ReportClosureFuncError("LiquidablePositions", s.svcTags)
			s.logger.Warning("Failed to get liquidable positions")

			metrics.ReportClosureFuncTiming("LiquidablePositions", s.svcTags)
			time.Sleep(10 * time.Second)
			continue
		}

		positions := resp.Positions
		market := s.marketsAssistant.AllDerivativeMarkets()[s.marketID]

		if len(positions) > 0 {
			for _, position := range positions {
				msg := s.createLiquidationMessage(position, market)

				if _, err := s.chainClient.SyncBroadcastMsg(msg); err != nil {
					metrics.ReportClosureFuncError("SyncBroadcastMsg", s.svcTags)
					s.logger.Errorf("Failed liquidating position %s with error %s", position.String(), err.Error())
				}
			}
		}

		metrics.ReportClosureFuncTiming("LiquidablePositions", s.svcTags)
		time.Sleep(10 * time.Second)
	}
}

func (s *liquidatorSvc) panicRecover(err *error) {
	if r := recover(); r != nil {
		*err = errors.Errorf("%v", r)

		if e, ok := r.(error); ok {
			s.logger.WithError(e).Errorln("service main loop panicked with an error")
			s.logger.Debugln(string(debug.Stack()))
		} else {
			s.logger.Errorln(r)
		}
	}
}

func (s *liquidatorSvc) Close() {
	// graceful shutdown if needed
}

func (s *liquidatorSvc) createLiquidationMessage(position *derivativeExchangePB.DerivativePosition, market core.DerivativeMarket) sdktypes.Msg {
	var msg sdktypes.Msg
	if s.granterPublicAddress == "" {
		liquidatePositionMessage := s.createLiquidatePositionMessage(position, market, s.chainClient.FromAddress().String(), s.subaccountID)
		msg = &liquidatePositionMessage
	} else {
		liquidatePositionMessage := s.createLiquidatePositionMessage(position, market, s.granterPublicAddress, s.granterSubaccountID)
		liquidationMessageBytes, _ := liquidatePositionMessage.Marshal()
		liquidationMsgAsAny := &codectypes.Any{
			TypeUrl: sdktypes.MsgTypeURL(&liquidatePositionMessage),
			Value:   liquidationMessageBytes,
		}
		msg = &authz.MsgExec{
			Grantee: s.chainClient.FromAddress().String(),
			Msgs:    []*codectypes.Any{liquidationMsgAsAny},
		}
	}

	return msg
}

func (s *liquidatorSvc) createLiquidatePositionMessage(
	position *derivativeExchangePB.DerivativePosition,
	market core.DerivativeMarket,
	senderAddress string,
	senderSubaccountID common.Hash,
) exchangetypes.MsgLiquidatePosition {

	orderType := exchangetypes.OrderType_BUY
	if position.Direction == "short" {
		orderType = exchangetypes.OrderType_SELL
	}

	candidateOrderAmountFromMaxNotional := s.maxOrderNotional.Quo(math.LegacyMustNewDecFromStr(position.MarkPrice))
	fullLiquidationOrderAmount := math.LegacyMustNewDecFromStr(position.Quantity)
	orderAmount := math.LegacyMinDec(candidateOrderAmountFromMaxNotional, fullLiquidationOrderAmount)
	orderAmount = math.LegacyMinDec(orderAmount, s.maxOrderAmount)

	order := s.chainClient.CreateDerivativeOrder(
		senderSubaccountID,
		&chainclient.DerivativeOrderData{
			OrderType:    orderType,
			Quantity:     market.QuantityFromChainFormat(orderAmount),
			Price:        market.PriceFromChainFormat(math.LegacyMustNewDecFromStr(position.MarkPrice)),
			Leverage:     decimal.RequireFromString("1"),
			FeeRecipient: senderAddress,
			MarketId:     market.Id,
			Cid:          uuid.NewString(),
		},
		s.marketsAssistant,
	)

	msg := exchangetypes.MsgLiquidatePosition{
		Sender:       senderAddress,
		SubaccountId: position.SubaccountId,
		MarketId:     position.MarketId,
		Order:        order,
	}

	return msg
}
