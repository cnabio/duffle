package main

import (
	"fmt"
	"os"
	"os/user"

	"github.com/deis/duffle/pkg/signature"
)

// defaultUserID returns the default user name.
func defaultUserID() signature.UserID {
	// TODO: I am not sure how reliable this is on Windows.
	domain, err := os.Hostname()
	if err != nil {
		domain = "localhost.localdomain"
	}
	var name, username string

	if account, err := user.Current(); err != nil {
		name = "user"
	} else {
		name = account.Name
		username = account.Username
	}
	email := fmt.Sprintf("%s@%s", username, domain)
	return signature.UserID{
		Name:  name,
		Email: email,
	}
}
