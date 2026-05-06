package obs

import "github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"

type StorageClassType = obs.StorageClassType

const (
	StorageClassStandard StorageClassType = obs.StorageClassStandard
	StorageClassWarm     StorageClassType = obs.StorageClassWarm
	StorageClassCold     StorageClassType = obs.StorageClassCold
)
