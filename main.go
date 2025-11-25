package main

import (
	"fmt"
	"os"

	"github.com/AdityaKK0407/sentryvault/internal/app"
	"github.com/AdityaKK0407/sentryvault/internal/database"
)

func main() {
	fmt.Println(app.AsciiArt())

	// Get all  users present
	files, err := database.GetDBFiles()
	if err != nil {
		fmt.Printf("An error occurred: %+v\n", err)
		os.Exit(1)
	}

	// Run the Auth
	username, password, newUser, err := app.RunAuth(files)
	if err != nil {
		fmt.Printf("An error occurred: %+v\n", err)
		os.Exit(1)
	}

	// Create database instance
	db, err := database.Open(username)
	if err != nil {
		fmt.Printf("An error occurred: %+v\n", err)
		os.Exit(1)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			fmt.Printf("An error occurred: %+v\n", err)
			os.Exit(1)
		}
	}()

	// Run the encryption/decryption
	cipherKey32, err := app.RunCipher(db, username, password, newUser)
	if err != nil {
		fmt.Printf("An error occurred: %+v\n", err)
		os.Exit(1)
	}

	var cipherKey64 []byte

	// Run the Model
	if err = app.RunModel(db, cipherKey32, cipherKey64); err != nil {
		fmt.Printf("An error occurred: %+v\n", err)
		os.Exit(1)
	}
}
