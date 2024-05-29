package models

import (
	"time"
	"github.com/jinzhu/gorm"
	"github.com/lib/pq" // for PostgreSQL JSONB
)

type Matrix struct {
	ID          uint64         `gorm:"primary_key;auto_increment" json:"id"`
	Coordinates pq.Int32Array  `gorm:"type:jsonb;not null" json:"coordinates"`
	UserID      uint32         `sql:"type:int REFERENCES users(id)" json:"user_id"`
	User        User           `json:"user"`
	CreatedAt   time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (m *Matrix) Prepare() {
	m.ID = 0
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
}

func (m *Matrix) Validate() error {
	if len(m.Coordinates) == 0 {
		return errors.New("Required Coordinates")
	}
	if m.UserID < 1 {
		return errors.New("Required UserID")
	}
	return nil
}

func (m *Matrix) SaveMatrix(db *gorm.DB) (*Matrix, error) {
	var err error
	err = db.Debug().Model(&Matrix{}).Create(&m).Error
	if err != nil {
		return &Matrix{}, err
	}
	if m.ID != 0 {
		err = db.Debug().Model(&User{}).Where("id = ?", m.UserID).Take(&m.User).Error
		if err != nil {
			return &Matrix{}, err
		}
	}
	return m, nil
}

func (m *Matrix) FindAllMatrices(db *gorm.DB) (*[]Matrix, error) {
	var err error
	matrices := []Matrix{}
	err = db.Debug().Model(&Matrix{}).Limit(100).Find(&matrices).Error
	if err != nil {
		return &[]Matrix{}, err
	}
	if len(matrices) > 0 {
		for i := range matrices {
			err := db.Debug().Model(&User{}).Where("id = ?", matrices[i].UserID).Take(&matrices[i].User).Error
			if err != nil {
				return &[]Matrix{}, err
			}
		}
	}
	return &matrices, nil
}

func (m *Matrix) FindMatrixByID(db *gorm.DB, mid uint64) (*Matrix, error) {
	var err error
	err = db.Debug().Model(&Matrix{}).Where("id = ?", mid).Take(&m).Error
	if err != nil {
		return &Matrix{}, err
	}
	if m.ID != 0 {
		err = db.Debug().Model(&User{}).Where("id = ?", m.UserID).Take(&m.User).Error
		if err != nil {
			return &Matrix{}, err
		}
	}
	return m, nil
}
