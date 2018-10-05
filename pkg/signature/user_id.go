package signature

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// UserID models a user ID
//
// In OpenPGP, this is usually structured as `NAME (COMMENT) <EMAIL>`
type UserID struct {
	Name, Comment, Email string
}

// String reprsents the UserID as an OpenPGP user string
func (u UserID) String() string {
	// TODO: Handle case where comment is empty
	comment := ""
	if u.Comment != "" {
		comment = fmt.Sprintf(" (%s)", u.Comment)
	}
	return fmt.Sprintf("%s%s <%s>", u.Name, comment, u.Email)
}

// Not sure about using [[:print]], so I'll leave this here in case that one does not work for some edge case.
//var userIDrx = regexp.MustCompile(`^([\w\s\.\+\-\_@]+)(?:\s*?\(([\w\s\.\+\-\_@]*)\))?(?:\s+\<([a-zA-Z0-9\.\+\-\_]+@[a-zA-Z0-9\.\+\-\_]+)\>)?$`)
var userIDrx = regexp.MustCompile(`^([[:print:]]+?)(?:\s*?\(([[:print:]]*?)\))?(?:\s+\<([a-zA-Z0-9\.\+\-\_]+@[a-zA-Z0-9\.\+\-\_]+)\>)?$`)
var emailish = regexp.MustCompile(`^[a-zA-Z0-9\.\+\-\_]+@[a-zA-Z0-9\.\+\-\_]+$`)

// ParseUserID attempts to parse the format `NAME (COMMENT) <EMAIL>` into three fields.
//
//  - If name is empty, this will return an error
// 	- If comment is omitted or empty, the empty string is returned.
// 	- If email is omitted, this will check the name to see if it looks like an email address, and use it or else error out
func ParseUserID(id string) (UserID, error) {
	matches := userIDrx.FindStringSubmatch(id)
	ret := UserID{}

	if len(matches) != 4 {
		return ret, errors.New("invalid ID format")
	}

	if matches[1] == "" {
		return ret, errors.New("name field is required")
	}
	ret.Name = strings.TrimSpace(matches[1])
	ret.Comment = strings.TrimSpace(matches[2])

	if matches[3] == "" {
		if !emailish.MatchString(ret.Name) {
			return ret, errors.New("email is required")
		}
		ret.Email = ret.Name
	} else {
		ret.Email = strings.TrimSpace(matches[3])
	}

	return ret, nil
}
