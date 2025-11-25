package app

import (
	"errors"

	"github.com/AdityaKK0407/sentryvault/internal/cipher"
	"github.com/AdityaKK0407/sentryvault/internal/database"
	"github.com/AdityaKK0407/sentryvault/internal/model"
	tea "github.com/charmbracelet/bubbletea"
	bolt "go.etcd.io/bbolt"
)

func RunAuth(files []string) (string, string, bool, error) {
	if len(files) == 0 {
		username, password, err := CreateUser()
		if err != nil {
			return "", "", false, err
		}
		return username, password, true, err
	} else {
		return SelectUser(files)
	}
}

func RunCipher(db *bolt.DB, username, password string, newUser bool) ([]byte, error) {
	if newUser {
		salt, err := cipher.GenerateRandomSalt()
		if err != nil {
			return nil, err
		}
		cipherKey32 := cipher.DeriveEncryptionKey32([]byte(password), salt)

		title := username + "'s Vault"
		combinedTitle, err := cipher.EncryptAESGCM(cipherKey32, []byte(title))
		if err != nil {
			return nil, err
		}

		if err = database.SetHeaders(db, combinedTitle, salt); err != nil {
			return nil, err
		}

		return cipherKey32, nil
	} else {
		combinedTitle, salt, err := database.GetHeaders(db)
		cipherKey32 := cipher.DeriveEncryptionKey32([]byte(password), salt)
		_, err = cipher.DecryptAESGCM(cipherKey32, combinedTitle)

		if err != nil {
			return nil, errors.New("invalid password")
		}

		return cipherKey32, nil
	}
}

func RunModel(db *bolt.DB, cipherKey32, cipherKey64 []byte) error {
	p := tea.NewProgram(model.InitialMainModel(db, cipherKey32, cipherKey64))
	m, err := p.Run()
	if err != nil {
		return err
	}
	if finalModel, ok := m.(model.MainModel); ok {
		if finalModel.Err != nil {
			return finalModel.Err
		}
	}
	return nil
}
