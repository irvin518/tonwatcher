package tonwatcher

import "context"

type AddressCache interface {
	AddAddressWatch(ctx context.Context, address string, precessed uint64) error

	QueryLastProcessed(ctx context.Context, address string) (uint64, error)

	UpdateLastProcessed(ctx context.Context, address string, processed uint64) error
}
