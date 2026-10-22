package secure

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/viper"
	"github.com/zalando/go-keyring"
	"golang.org/x/term"
)

func SetPassword(email string) (string, error) {
	fmt.Print("Enter password > ")
	inputPass, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Println()

	password := strings.TrimSpace(string(inputPass))

	if password == "" {
		return "", errors.New("Password cannot be empty")
	}

	err = keyring.Set("EmailDodeyo", email, password)
	if err != nil {
		fmt.Println("Password couldnt be saved")
		return password, nil
	}

	return password, nil
}

func GetPassword(email string) (string, error) {
	password, err := keyring.Get("EmailDodeyo", email)
	if err != nil {
		fmt.Printf("Password for %s not found in keyring\n", email)
		password, err = SetPassword(email)
		if err != nil {
			return "", err
		}
	}
	return password, nil
}

func NewEmailPassword() (string, string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter email > ")

	inputMail, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}
	email := strings.TrimSpace(inputMail)

	viper.Set("userMail", email)
	err = viper.WriteConfig()
	if err != nil {
		return "", "", err
	}

	password, err := SetPassword(email)
	if err != nil {
		return "", "", err
	}

	return email, password, nil
}

func SetKindleEmail() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter kindle email > ")

	inputMail, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	email := strings.TrimSpace(inputMail)

	viper.Set("kindlemail", email)
	err = viper.WriteConfig()
	if err != nil {
		return "", err
	}
	return email, nil
}
