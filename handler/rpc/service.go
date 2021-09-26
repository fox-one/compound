package rpc

import (
	"compound/core"
	"compound/pkg/mtg"
	context "context"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/uuid"
	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type RPCService struct {
	System            *core.System
	Dapp              *core.Wallet
	TransactionStore  core.TransactionStore
	MarketStore       core.IMarketStore
	OracleSignerStore core.OracleSignerStore
	SupplyStore       core.ISupplyStore
	BorrowStore       core.IBorrowStore
	MarketService     core.IMarketService
}

func NewServiceImpl(system *core.System,
	dapp *core.Wallet,
	transactionStr core.TransactionStore,
	marketStr core.IMarketStore,
	oracleSignerStr core.OracleSignerStore,
	supplyStr core.ISupplyStore,
	borrowStr core.IBorrowStore,
	marketSrv core.IMarketService) *RPCService {
	return &RPCService{
		System:            system,
		Dapp:              dapp,
		TransactionStore:  transactionStr,
		MarketStore:       marketStr,
		OracleSignerStore: oracleSignerStr,
		SupplyStore:       supplyStr,
		BorrowStore:       borrowStr,
		MarketService:     marketSrv,
	}
}

func (s *RPCService) AllMarkets(ctx context.Context, req *MarketReq) (*MarketListResp, error) {
	markets, e := s.MarketStore.All(ctx)
	if e != nil {
		return nil, e
	}

	marketViews := make([]*Market, 0)
	for _, m := range markets {
		supplyRate, e := s.MarketService.CurSupplyRate(ctx, m)
		if e != nil {
			supplyRate = decimal.Zero
		}
		borrowRate, e := s.MarketService.CurBorrowRate(ctx, m)
		if e != nil {
			borrowRate = decimal.Zero
		}

		countOfSupplies, e := s.SupplyStore.CountOfSuppliers(ctx, m.CTokenAssetID)
		if e != nil {
			countOfSupplies = 0
		}

		countOfBorrows, e := s.BorrowStore.CountOfBorrowers(ctx, m.AssetID)
		if e != nil {
			countOfBorrows = 0
		}
		marketView := Market{
			Id:                   int64(m.ID),
			AssetId:              m.AssetID,
			Symbol:               m.Symbol,
			CtokenAssetId:        m.CTokenAssetID,
			TotalCash:            m.TotalCash.String(),
			TotalBorrows:         m.TotalBorrows.String(),
			Reserves:             m.Reserves.String(),
			Ctokens:              m.CTokens.String(),
			InitExchangeRate:     m.InitExchangeRate.String(),
			ReserveFactor:        m.ReserveFactor.String(),
			LiquidationIncentive: m.LiquidationIncentive.String(),
			BorrowCap:            m.BorrowCap.String(),
			CollateralFactor:     m.CollateralFactor.String(),
			CloseFactor:          m.CloseFactor.String(),
			BaseRate:             m.BaseRate.String(),
			Multiplier:           m.Multiplier.String(),
			JumpMultiplier:       m.JumpMultiplier.String(),
			Kink:                 m.Kink.String(),
			BlockNumber:          m.BlockNumber,
			UtilizationRate:      m.UtilizationRate.String(),
			ExchangeRate:         m.ExchangeRate.String(),
			SupplyRatePerBlock:   m.SupplyRatePerBlock.String(),
			BorrowRatePerBlock:   m.BorrowRatePerBlock.String(),
			Price:                m.Price.String(),
			PriceUpdateAt:        timestamppb.New(m.PriceUpdatedAt),
			BorrowIndex:          m.BorrowIndex.String(),
			Status:               int32(m.Status),
			CreatedAt:            timestamppb.New(m.CreatedAt),
			UpdatedAt:            timestamppb.New(m.UpdatedAt),
			Suppliers:            countOfSupplies,
			Borrowers:            countOfBorrows,
			SupplyApy:            supplyRate.String(),
			BorrowApy:            borrowRate.String(),
		}
		marketViews = append(marketViews, &marketView)
	}

	resp := MarketListResp{
		Data: marketViews,
	}

	return &resp, nil
}

func (s *RPCService) PriceRequest(ctx context.Context, req *PriceReq) (*PriceRequestResp, error) {
	markets, err := s.MarketStore.All(ctx)
	if err != nil {
		return nil, err
	}

	// members
	var members []string
	for _, m := range s.System.Members {
		members = append(members, m.ClientID)
	}

	// signers
	ss, err := s.OracleSignerStore.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	signers := make([]*PriceSigner, len(ss))
	for idx, s := range ss {
		signers[idx] = &PriceSigner{
			Index:     int32(idx) + 1,
			VerifyKey: s.PublicKey,
		}
	}

	requests := make([]*Price, 0)
	for _, m := range markets {
		if m.PriceThreshold > 0 && time.Now().After(m.PriceUpdatedAt.Add(10*time.Minute)) {
			requests = append(requests, &Price{
				TraceId: uuid.Modify(m.AssetID, fmt.Sprintf("price-request:%s:%d", s.System.ClientID, time.Now().Unix()/600)),
				AssetId: m.AssetID,
				Symbol:  m.Symbol,
				Receiver: &PriceReceiver{
					Threshold: int32(s.System.Threshold),
					Members:   members,
				},
				Signers:   signers,
				Threshold: int32(m.PriceThreshold),
			})
		}
	}

	resp := PriceRequestResp{
		Data: requests,
	}

	return &resp, nil
}

func (s *RPCService) Transactions(ctx context.Context, req *TransactionReq) (*TransactionListResp, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 500
	}

	offset := req.Offset.AsTime()

	fmt.Println("offset:", offset, ":limit:", limit)

	transactions, e := s.TransactionStore.List(ctx, offset, int(limit))
	if e != nil {
		return nil, e
	}

	fmt.Println("transactions.len:", len(transactions))

	items := make([]*Transaction, 0)
	for _, t := range transactions {
		i := Transaction{
			Id:              t.ID,
			Action:          int32(t.Action),
			TraceId:         t.TraceID,
			UserId:          t.UserID,
			FollowId:        t.FollowID,
			SnapshotTraceId: t.SnapshotTraceID,
			AssetId:         t.AssetID,
			Amount:          t.Amount.String(),
			Data:            t.Data,
			CreatedAt:       timestamppb.New(t.CreatedAt),
		}
		items = append(items, &i)
	}
	resp := TransactionListResp{
		Data: items,
	}
	return &resp, nil
}

func (s *RPCService) PayRequest(ctx context.Context, req *PayReq) (*PayResp, error) {
	memoBytes, err := base64.StdEncoding.DecodeString(req.MemoBase64)
	if err != nil {
		return nil, err
	}

	key := mixin.GenerateEd25519Key()
	pub := s.System.PrivateKey.Public().(ed25519.PublicKey)

	memoEncrypt, err := mtg.Encrypt(memoBytes, key, pub)
	if err != nil {
		return nil, err
	}

	assetID := req.AssetId
	amount, _ := decimal.NewFromString(req.Amount)

	if req.WithGas {
		assetID = s.System.VoteAsset
		amount = s.System.VoteAmount
	}

	input := mixin.TransferInput{
		AssetID: assetID,
		Amount:  amount,
		TraceID: req.TraceId,
		Memo:    base64.StdEncoding.EncodeToString(memoEncrypt),
	}
	input.OpponentMultisig.Receivers = s.System.MemberIDs
	input.OpponentMultisig.Threshold = s.System.Threshold

	payment, err := s.Dapp.Client.VerifyPayment(ctx, input)
	if err != nil {
		return nil, err
	}

	url := mixin.URL.Codes(payment.CodeID)

	var response struct {
		URL string `json:"url"`
	}

	response.URL = url

	payResp := PayResp{
		Url: url,
		TransferInput: &TransferInput{
			AssetId: input.AssetID,
			Amount:  input.Amount.String(),
			TraceId: input.TraceID,
			Memo:    input.Memo,
			OpponentMultisig: &OpponentMultiSig{
				Receivers: input.OpponentMultisig.Receivers,
				Threshold: int32(input.OpponentMultisig.Threshold),
			},
		},
	}

	return &payResp, nil
}
