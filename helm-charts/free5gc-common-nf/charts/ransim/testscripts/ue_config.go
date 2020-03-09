package test

import (
	"gofree5gc/lib/nas/nasType"
)

type UeData struct {
	Supi              string `yaml:"supi"`
	RanUeNgapId       int64  `yaml:"ranUeNgapId"`
	AmfUeNgapId       int64  `yaml:"amfUeNgapId"`
	Sst               int32  `yaml:"sst"`
	Sd                string `yaml:"sd"`
	MobileIdentity5GS nasType.MobileIdentity5GS
}
