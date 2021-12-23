package sql

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type BaseGorm struct {
	ID        int64          `gorm:"primaryKey" sql:"auto_increment:true"`
	CreatedAt time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:current_timestamp(6)"`
	UpdatedAt time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:current_timestamp(6)"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;type:timestamp;index"`
}

func (b *BaseGorm) AfterCreate(tx *gorm.DB) (err error) {
	return tx.Exec(fmt.Sprintf(`SELECT setval('%[1]s_id_seq', (SELECT MAX(id) from "%[1]s"));`, tx.Statement.Table)).Debug().Error
}
