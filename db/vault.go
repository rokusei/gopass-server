package db

import (
	"context"

	"gorm.io/gorm"
)

// a Vault contains a UUID and a list of vault entries
type Vault struct {
	gorm.Model
	Entries []VaultEntry
}

// a VaultEntry contains a UUID and an encrypted blob
type VaultEntry struct {
	gorm.Model
	UUID    string `gorm:"primaryKey"`
	VaultID uint64
	Blob    []byte
}

// GetVault fetches a User's Vault
func GetVault(ctx context.Context, db *gorm.DB, user User) (Vault, error) {
	if err := ctx.Err(); err != nil {
		return Vault{}, err
	}

	var vault Vault
	result := db.Model(&user).Association("Vault").DB.First(vault)
	if result.Error != nil {
		return Vault{}, result.Error
	}
	return vault, nil
}

// GetVaultEntry fetches a VaultEntry from a Vault
func GetVaultEntry(ctx context.Context, db *gorm.DB, vault Vault, entryID uint64) (VaultEntry, error) {
	if err := ctx.Err(); err != nil {
		return VaultEntry{}, err
	}

	var vaultEntry VaultEntry
	result := db.Model(&vault).Association("VaultEntry").DB.First(vaultEntry, entryID)
	if result.Error != nil {
		return VaultEntry{}, result.Error
	}
	return vaultEntry, nil
}
