package enum

type BucketType uint

const (
	BucketTypeReceipts BucketType = iota
)

func (bt BucketType) String() string {
	switch bt {
	case BucketTypeReceipts:
		return "receipts"
	default:
		return "unknown"
	}
}

func GetAllBucketTypes() []BucketType {
	return []BucketType{BucketTypeReceipts}
}
