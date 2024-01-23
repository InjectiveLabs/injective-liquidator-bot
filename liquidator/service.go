package liquidator

import (
	"context"
	sdkCommon "github.com/InjectiveLabs/sdk-go/client/common"
	"github.com/InjectiveLabs/sdk-go/client/exchange"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"runtime/debug"
	"time"

	"github.com/pkg/errors"
	log "github.com/xlab/suplog"

	"github.com/InjectiveLabs/metrics"
	exchangetypes "github.com/InjectiveLabs/sdk-go/chain/exchange/types"
	chainclient "github.com/InjectiveLabs/sdk-go/client/chain"
	derivativeExchangePB "github.com/InjectiveLabs/sdk-go/exchange/derivative_exchange_rpc/pb"
	metaPB "github.com/InjectiveLabs/sdk-go/exchange/meta_rpc/pb"
)

type Service interface {
	Start() error
	Close()
}

type liquidatorSvc struct {
	chainClient      chainclient.ChainClient
	exchangeClient   exchange.ExchangeClient
	marketsAssistant chainclient.MarketsAssistant
	marketID         string
	subaccountID     common.Hash

	logger  log.Logger
	svcTags metrics.Tags
}

func NewService(
	chainClient chainclient.ChainClient,
	exchangeClient exchange.ExchangeClient,
	marketsAssistant chainclient.MarketsAssistant,
	marketID string,
	subaccountID common.Hash,
) Service {
	return &liquidatorSvc{
		logger: log.WithField("svc", "liquidator"),
		svcTags: metrics.Tags{
			"svc": "liquidator_bot",
		},

		chainClient:      chainClient,
		exchangeClient:   exchangeClient,
		marketsAssistant: marketsAssistant,
		marketID:         marketID,
		subaccountID:     subaccountID,
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
				orderType := exchangetypes.OrderType_BUY
				if position.Direction == "short" {
					orderType = exchangetypes.OrderType_SELL
				}

				order := s.chainClient.CreateDerivativeOrder(
					s.subaccountID,
					sdkCommon.NewNetwork(),
					&chainclient.DerivativeOrderData{
						OrderType:    orderType,
						Quantity:     market.QuantityFromChainFormat(types.MustNewDecFromStr(position.Quantity)),
						Price:        market.PriceFromChainFormat(types.MustNewDecFromStr(position.MarkPrice)),
						Leverage:     decimal.RequireFromString("1"),
						FeeRecipient: s.chainClient.FromAddress().String(),
						MarketId:     "0x17ef48032cb24375ba7c2e39f384e56433bcab20cbee9a7357e4cba2eb00abe6",
						Cid:          uuid.NewString(),
					},
					s.marketsAssistant,
				)

				msg := &exchangetypes.MsgLiquidatePosition{
					Sender:       s.chainClient.FromAddress().String(),
					SubaccountId: position.SubaccountId,
					MarketId:     position.MarketId,
					Order:        order,
				}

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