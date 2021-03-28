package db

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

var ErrVaultNotFound = errors.New("vault not found")
var ErrEntryNotFound = errors.New("entry not found")

// a Vault contains a list of vault entries
type Vault struct {
	gorm.Model
	ID           uint   `gorm:"primarykey" json:"-"`
	UUID         string `json:"ID"`
	VaultEntries []VaultEntry
}

// a VaultEntry contains a UUID and an encrypted blob
type VaultEntry struct {
	gorm.Model
	ID             uint   `gorm:"primarykey" json:"-"`
	UUID           string `json:"ID"`
	VaultID        uint   `json:"-"`
	EncryptedEntry []byte
}

// GetVault fetches a User's Vault
func GetVault(ctx context.Context, db *gorm.DB, user *User) (*Vault, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	/*var vault Vault
	err := db.Debug().Model(user).Association("Vault").Find(&vault)
	if err != nil {
		return nil, err
	}*/

	err := db.Debug().Model(&user).Association("VaultEntries").Find(&user.Vault.VaultEntries)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Vault: %#v\n", user.Vault)
	fmt.Printf("VaultEntries: %#v\n", user.Vault.VaultEntries)
	return &user.Vault, nil
}

// GetVaultEntry fetches a VaultEntry from a Vault
func GetVaultEntry(ctx context.Context, db *gorm.DB, user *User, entryUUID string) (*VaultEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var entries []VaultEntry
	err := db.Debug().Find(user).Model(&user.Vault).Association("VaultEntries").Find(&entries)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Vault: %#v\n", user.Vault)
	for _, entry := range entries {
		fmt.Printf("Vault Entry: %#v\n", entry)
		if entry.UUID == entryUUID && entry.VaultID == user.Vault.ID {
			return &entry, nil
		}
	}

	return nil, ErrEntryNotFound
}

// GetVaultEntry fetches a VaultEntry from a Vault
func CreateVaultEntry(ctx context.Context, db *gorm.DB, user *User, encryptedEntry []byte) (*VaultEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	uuid, err := GenerateUUID()
	if err != nil {
		return nil, err
	}

	entry := VaultEntry{
		UUID:           uuid,
		VaultID:        user.Vault.ID,
		EncryptedEntry: encryptedEntry,
	}
	result := db.Debug().Create(&entry)
	if result.Error != nil {
		return nil, result.Error
	}

	fmt.Printf("CreateVaultEntry: Vault: %#v", user.Vault)

	err = db.Debug().Find(user).Model(&user.Vault).Association("VaultEntries").Append(&entry)
	if err != nil {
		return nil, err
	}
	db.Debug().Save(&entry)
	return &entry, nil
}

// GetVaultEntry fetches a VaultEntry from a Vault
func UpdateVaultEntry(ctx context.Context, db *gorm.DB, user *User, entryUUID string, encryptedEntry []byte) (*VaultEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var entries []VaultEntry
	err := db.Debug().Find(user).Model(&user.Vault).Association("VaultEntries").Find(&entries)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Vault: %#v\n", user.Vault)
	for i, entry := range entries {
		if entry.UUID == entryUUID {
			entries[i].EncryptedEntry = encryptedEntry
			db.Debug().Updates(entries)
			db.Debug().Save(&entry)
			return &entries[i], nil
		}
	}
	return nil, ErrEntryNotFound
}
