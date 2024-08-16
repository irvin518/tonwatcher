package toninterface

import (
	"context"
	"log"
	"math/big"
	"time"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/ton/nft"
)

type NftScanner struct {
	nftAddress string
	IdMaps     map[int64]*address.Address
	cancle     chan any
}

func NewScanner(addr string) *NftScanner {
	return &NftScanner{
		nftAddress: addr,
		IdMaps:     make(map[int64]*address.Address),
		cancle:     make(chan any),
	}
}

func (l *NftScanner) Begin(interval int64) {
	api := Api()
	nftAddress := address.MustParseAddr(l.nftAddress)
	client := nft.NewCollectionClient(api, nftAddress)
	scanFunc := func() {
		temp := make(map[int64]*address.Address)

		ctx := context.Background()
		data, err := client.GetCollectionData(ctx)
		if err != nil {
			log.Printf("get nft collection data error %s", err.Error())
			return
		}

		nextIndex := data.NextItemIndex
		var i int64 = 0
		for ; i < nextIndex.Int64(); i++ {
			addr, err := client.GetNFTAddressByIndex(ctx, big.NewInt(i))
			if err != nil {
				log.Printf("get nft item address error %s", err.Error())
				continue
			}
			item := nft.NewItemClient(api, addr)
			d, err := item.GetNFTData(ctx)
			if err != nil {
				log.Printf("get nft item data error %s", err.Error())
				continue
			}
			temp[d.Index.Int64()] = d.OwnerAddress
		}
		l.IdMaps = temp
	}
	go func() {
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		defer ticker.Stop() // 确保在不再需要时停止Ticker

		for {
			select {
			case <-l.cancle:
				return
			case <-ticker.C: // 等待Ticker通道发送时间
				scanFunc()
			}
		}
	}()

}

func (l *NftScanner) Stop() {
	close(l.cancle)
}
