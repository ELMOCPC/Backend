package enum

type BucketType uint

const ()

func (bt BucketType) String() string {
	return "unknown"
}

func GetAllBucketTypes() []BucketType {
	return []BucketType{}
}
