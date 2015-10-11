package main

import (
	"regexp"
	"strings"
)

type AccountName string

type AccountType int

const (
	AccountTypeNone AccountType = iota
	AccountTypeAsset
	AccountTypeDebt
	AccountTypeExpense
	AccountTypeIncome
)

func (typ AccountType) String() string {
	switch typ {
	case AccountTypeAsset:
		return "asset"
	case AccountTypeDebt:
		return "debt"
	case AccountTypeExpense:
		return "expense"
	case AccountTypeIncome:
		return "income"
	}
	return "none"
}

// Length of the account type character and following colon.
const AccountNameTypePrefixLen = 2

// Returns the parent account name or "" if there is none.
// The account name is assumed to be valid as indicated by Valid().
// Example: "e:food.snacks" -> "e:food" -> "e:" -> ""
func (name AccountName) Parent() AccountName {
	lastIndex := strings.LastIndexByte(string(name), '.')
	if lastIndex == -1 && len(name) > AccountNameTypePrefixLen {
		// Return the type prefix as the root node.
		// It's a vaild account according to the specification.
		lastIndex = AccountNameTypePrefixLen
	}
	if lastIndex == -1 {
		lastIndex = 0
	}
	return name[:lastIndex]
}

// Get the name of the leftmost account group.
// The account name is assumed to be valid as indicated by Valid().
// Example: "e:food.snacks" -> "snacks", "e:food" -> "food", "e:" -> "expense"
func (name AccountName) Leaf() string {
	lastIndex := strings.LastIndexByte(string(name), '.')
	if lastIndex != -1 {
		return string(name[lastIndex+1:])
	}
	lastIndex = strings.LastIndexByte(string(name), ':')
	if lastIndex != -1 && len(name) > AccountNameTypePrefixLen {
		return string(name[lastIndex+1:])
	}
	return name.Type().String()
}

// Get the depth of the account name grouping.
// The account name is assumed to be valid as indicated by Valid().
// Example: "e:food.snacks" -> 2, "e:food" -> 1, "e:" -> 0
func (name AccountName) Depth() int {
	if len(name) == AccountNameTypePrefixLen {
		return 0
	}
	return 1 + strings.Count(string(name), ".")
}

// Return a descriptive type string based on the first character in the account name.
func (name AccountName) Type() AccountType {
	if len(name) > 0 {
		switch name[0] {
		case 'a':
			return AccountTypeAsset
		case 'd':
			return AccountTypeDebt
		case 'e':
			return AccountTypeExpense
		case 'i':
			return AccountTypeIncome
		}
	}
	return AccountTypeNone
}

var accountNameRegexp = regexp.MustCompile("^[adei]:([\\pL-]+(\\.[\\pL-]+)*|)$")

// Check if the account name is valid according to the accounting file format
// specification.
func (name AccountName) Valid() bool {
	return accountNameRegexp.MatchString(string(name))
}
