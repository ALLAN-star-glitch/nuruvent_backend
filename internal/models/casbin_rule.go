package models

type CasbinRule struct {
	ID    uint   `gorm:"primaryKey"`
	Ptype string `gorm:"size:100;index"`
	V0    string `gorm:"size:100"`
	V1    string `gorm:"size:100"`
	V2    string `gorm:"size:100"`
	V3    string `gorm:"size:100"`
	V4    string `gorm:"size:100"`
	V5    string `gorm:"size:100"`
}

func (CasbinRule) TableName() string {
	return "casbin_rule"
}