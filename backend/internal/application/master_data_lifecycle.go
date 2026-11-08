package application

import "railkeeper/backend/internal/inventory/masterdata"

var ErrMasterDataBundled = masterdata.ErrMasterDataBundled
var ErrMasterDataInUse = masterdata.ErrMasterDataInUse
var ErrMasterDataUsageUnknown = masterdata.ErrMasterDataUsageUnknown

const MaxMasterDataActiveBatchSize = masterdata.MaxMasterDataActiveBatchSize
