package tonwatcher

import (
	"context"
	"log"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
)

type TonWatcher struct {
	ctx               context.Context
	api               ton.APIClientWrapped
	master            *ton.BlockIDExt
	WatcherTableModle AddressCache
}

type AddressCallback func(*tlb.Transaction, string)

func NewTonWatcher(ctx context.Context, api ton.APIClientWrapped, addressCache AddressCache) *TonWatcher {
	log.Printf("mastert proof checks since config init  block, it mai yake near a minute...")
	master, err := api.CurrentMasterchainInfo(ctx)
	if err != nil {
		log.Printf("get masterchain info err: %s", err.Error())
		return nil
	}

	log.Print("mastert proof checks are conmleted, now 100% safe!")

	return &TonWatcher{
		ctx:               ctx,
		api:               api,
		master:            master,
		WatcherTableModle: addressCache,
	}
}

func (l *TonWatcher) subscribe(treasuryAddress *address.Address, cb func(*tlb.Transaction) error, ctx context.Context, cancle context.CancelFunc) error {

	lastProcessedLT, err := l.WatcherTableModle.QueryLastProcessed(context.Background(), treasuryAddress.String())
	if err != nil {
		log.Printf("QueryLastProcessed %s, err %s", treasuryAddress.String(), err)
		return err
	}
	if lastProcessedLT == 0 { //new address
		acc, err := l.api.GetAccount(context.Background(), l.master, treasuryAddress)
		if err != nil {
			return err
		}
		// if !acc.IsActive {
		// 	log.Printf("%s is not active", treasuryAddress.String())
		// 	return fmt.Errorf("%s is not active", treasuryAddress.String())
		// }

		// Cursor of processed transaction, save it to your db
		// We start from last transaction, will not process transactions older than we started from.
		// After each processed transaction, save lt to your db, to continue after restart
		lastProcessedLT = acc.LastTxLT
		err = l.WatcherTableModle.AddAddressWatch(context.Background(), treasuryAddress.String(), lastProcessedLT)
		if err != nil {
			log.Printf("add watcher address error %s", err)
			return err
		}
	}

	go func() {
		// channel with new transactions
		// make sure call cancle
		defer func() {
			if cancle != nil {
				cancle()
			}
		}()
		transactions := make(chan *tlb.Transaction)
		go l.api.SubscribeOnTransactions(ctx, treasuryAddress, lastProcessedLT, transactions)
		// listen for new transactions from channel
		for tx := range transactions {
			err = cb(tx)
			if err != nil {
				log.Printf("parse tx error %s ", err)
			}
			// update last processed lt and save it in db
			lastProcessedLT = tx.LT

			err = l.WatcherTableModle.UpdateLastProcessed(context.Background(), treasuryAddress.String(), lastProcessedLT)
			if err != nil {
				log.Printf("add watcher address error %s", err)
				continue
			}
		}
	}()

	return nil
}

func (l *TonWatcher) WatchAddress(addr string, callback AddressCallback, callbackUrl string, ctx context.Context) error {
	treasuryAddress, err := address.ParseAddr(addr)
	if err != nil {
		return err
	}
	err = l.subscribe(treasuryAddress, func(tx *tlb.Transaction) error {
		// process transaction here
		if tx.IO.In != nil {
			callback(tx, callbackUrl)
		}

		return nil
	}, ctx, nil)
	return err
}
