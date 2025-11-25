package app

import (
	"strings"

	"github.com/charmbracelet/huh"
)

func CreateUser() (string, string, error) {
	var username, password string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Enter your username").
				Value(&username),

			huh.NewInput().
				EchoMode(huh.EchoModePassword).
				Title("Enter your password").
				Placeholder("Press enter to not give a password. (Recommended for security)").
				Value(&password),
		),
	)

	return username, password, form.Run()
}

func SelectUser(files []string) (string, string, bool, error) {
	for i := range len(files) {
		files[i] = strings.TrimSuffix(files[i], ".db")
	}
	newUser := "New User"
	files = append(files, newUser)

	var username, password string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Options(huh.NewOptions(files...)...).
				Title("Userid here").
				Value(&username),
		),
	)

	if err := form.Run(); err != nil {
		return "", "", false, err
	}

	if username == newUser {
		username, password, err := CreateUser()
		return username, password, true, err
	}

	form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				EchoMode(huh.EchoModePassword).
				Title("Enter your password").
				Value(&password),
		),
	)

	if err := form.Run(); err != nil {
		return "", "", false, err
	}

	return username, password, false, nil
}
