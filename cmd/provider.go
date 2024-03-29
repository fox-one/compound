package cmd

import (
	"compound/config"
	"compound/core"
	"compound/pkg/mtg"
	accountservice "compound/service/account"
	messageservice "compound/service/message"
	proposalservice "compound/service/proposal"
	walletservice "compound/service/wallet"
	"compound/store/borrow"
	"compound/store/market"
	"compound/store/message"
	"compound/store/oracle"
	"compound/store/proposal"
	"compound/store/supply"
	"compound/store/transaction"
	"compound/store/user"
	"compound/store/wallet"
	"compound/worker/cashier"
	"compound/worker/datadog"
	"fmt"
	_ "time/tzdata"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/fox-one/pkg/property"
	"github.com/fox-one/pkg/store/db"
	propertystore "github.com/fox-one/pkg/store/property"
)

// provide db instance
func provideDatabase() *db.DB {
	database := db.MustOpen(cfg.DB)
	if err := db.Migrate(database); err != nil {
		panic(err)
	}
	return database
}

// provide mixin dapp
func provideDapp() *core.Wallet {
	c, err := mixin.NewFromKeystore(&cfg.Dapp.Keystore)
	if err != nil {
		panic(err)
	}
	return &core.Wallet{
		Client: c,
		Pin:    cfg.Dapp.Pin,
	}
}

// provide system config info
func provideSystem() *core.System {
	members := make([]*core.Member, 0, len(cfg.Group.Members))
	memberIDs := make([]string, 0, len(cfg.Group.Members))
	for _, m := range cfg.Group.Members {
		verifyKey, err := mtg.DecodePublicKey(m.VerifyKey)
		if err != nil {
			panic(fmt.Errorf("decode verify key for member %s failed", m.ClientID))
		}

		members = append(members, &core.Member{
			ClientID:  m.ClientID,
			VerifyKey: verifyKey,
		})

		memberIDs = append(memberIDs, m.ClientID)
	}

	privateKey, err := mtg.DecodePrivateKey(cfg.Group.PrivateKey)
	if err != nil {
		panic(fmt.Errorf("base64 decode group private key failed: %w", err))
	}

	signKey, err := mtg.DecodePrivateKey(cfg.Group.SignKey)
	if err != nil {
		panic(fmt.Errorf("base64 decode group sign key failed: %w", err))
	}

	return &core.System{
		Admins:     cfg.Group.Admins,
		ClientID:   cfg.Dapp.ClientID,
		Members:    members,
		MemberIDs:  memberIDs,
		Threshold:  cfg.Group.Threshold,
		VoteAsset:  cfg.Group.Vote.Asset,
		VoteAmount: cfg.Group.Vote.Amount,
		PrivateKey: privateKey,
		SignKey:    signKey,
		Genesis:    cfg.Genesis,
	}
}

func provideCashierConfig() cashier.Config {
	return cashier.Config{
		Batch:    _flag.cashier.batch,
		Capacity: _flag.cashier.capacity,
	}
}

func provideDataDogConfig(cfg config.Config) datadog.Config {
	return datadog.Config{
		ConversationID: cfg.DataDog.ConversationID,
		Interval:       _flag.datadog.interval,
		Version:        rootCmd.Version,
	}
}

// ---------------store-----------------------------------------
func providePropertyStore(db *db.DB) property.Store {
	return propertystore.New(db)
}

func provideMarketStore(db *db.DB) core.IMarketStore {
	return market.New(db)
}

func provideSupplyStore(db *db.DB) core.ISupplyStore {
	return supply.New(db)
}

func provideBorrowStore(db *db.DB) core.IBorrowStore {
	return borrow.New(db)
}

func provideWalletStore(db *db.DB) core.WalletStore {
	return wallet.New(db)
}

func provideMessageStore(db *db.DB) core.MessageStore {
	return message.New(db)
}

func provideProposalStore(db *db.DB) core.ProposalStore {
	return proposal.New(db)
}

func provideUserStore(db *db.DB) core.UserStore {
	return user.New(db)
}

func provideTransactionStore(db *db.DB) core.TransactionStore {
	return transaction.New(db)
}

func provideOracleSignerStore(db *db.DB) core.OracleSignerStore {
	return oracle.NewSignerStore(db)
}

// ------------------service------------------------------------
func provideProposalService(client *mixin.Client, system *core.System, marketStore core.IMarketStore, messageStore core.MessageStore) core.ProposalService {
	return proposalservice.New(
		system,
		client,
		marketStore,
		messageStore,
	)
}

func provideMessageService(client *mixin.Client) core.MessageService {
	return messageservice.New(client)
}

func provideWalletService(client *mixin.Client) core.WalletService {
	members := make([]string, len(cfg.Group.Members))
	for i, member := range cfg.Group.Members {
		members[i] = member.ClientID
	}
	return walletservice.New(client, walletservice.Config{
		Pin:       cfg.Dapp.Pin,
		Members:   members,
		Threshold: cfg.Group.Threshold,
	})
}

func provideAccountService(
	marketStore core.IMarketStore,
	supplyStore core.ISupplyStore,
	borrowStore core.IBorrowStore,
) core.IAccountService {
	return accountservice.New(marketStore, supplyStore, borrowStore)
}
