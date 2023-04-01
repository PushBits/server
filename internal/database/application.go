package database

import (
	"errors"

	"github.com/pushbits/server/internal/assert"
	"github.com/pushbits/server/internal/model"

	"gorm.io/gorm"
)

// CreateApplication creates an application.
func (d *Database) CreateApplication(application *model.Application) error {
	return d.gormdb.Create(application).Error
}

// DeleteApplication deletes an application.
func (d *Database) DeleteApplication(application *model.Application) error {
	return d.gormdb.Delete(application).Error
}

// UpdateApplication updates an application.
func (d *Database) UpdateApplication(application *model.Application) error {
	return d.gormdb.Save(application).Error
}

// GetApplicationByID returns the application with the given ID or nil.
func (d *Database) GetApplicationByID(id uint) (*model.Application, error) {
	var application model.Application

	err := d.gormdb.First(&application, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	assert.Assert(application.ID == id)

	return &application, err
}

// GetApplicationByToken returns the application with the given token or nil.
func (d *Database) GetApplicationByToken(token string) (*model.Application, error) {
	var application model.Application

	err := d.gormdb.Where("token = ?", token).First(&application).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	assert.Assert(application.Token == token)

	return &application, err
}
