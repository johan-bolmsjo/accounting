package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"
)

type DataStore struct {
	aliases             map[string]*Alias
	accounts            map[string]*Account
	transactions        []*Transaction
	currentDate         time.Time
	currentDateLineMeta LineMeta
}

func NewDataStore() *DataStore {
	return &DataStore{
		aliases:  make(map[string]*Alias),
		accounts: make(map[string]*Account),
	}
}

// Add alias to the data store.
// Returns an error if an alias with the same name already exists.
func (ds *DataStore) AddAlias(aliasName, accountName string, lineMeta *LineMeta) error {
	if oldAlias := ds.aliases[aliasName]; oldAlias != nil {
		return lineMeta.ErrorAt(fmt.Sprintf("alias '%s' redefined, first seen at '%s'",
			aliasName, &oldAlias.lineMeta))
	}
	ds.aliases[aliasName] = &Alias{name: aliasName, accountName: accountName, lineMeta: *lineMeta}
	return nil
}

// Lookup alias, nil is returned if none found.
func (ds *DataStore) GetAlias(aliasName string) *Alias {
	return ds.aliases[aliasName]
}

// Add account to data store.
// If an account already exist with the same name it will be overwritten.
// Use GetAccount() to check if the account exists before adding one.
func (ds *DataStore) AddAccount(accountName string) *Account {
	account := &Account{name: accountName}
	ds.accounts[accountName] = account
	return account
}

// Lookup account, nil is returned if none found.
func (ds *DataStore) GetAccount(accountName string) *Account {
	return ds.accounts[accountName]
}

func (ds *DataStore) AddTransaction(amount float64, dr *Account, cr *Account) *Transaction {
	transaction := &Transaction{
		date:   ds.currentDate,
		amount: amount,
		dr:     dr,
		cr:     cr,
	}
	ds.transactions = append(ds.transactions, transaction)
	return transaction
}

// Set date to be used for coming transactions.
// Returns an error if the date is set to an older date than the current one.
func (ds *DataStore) SetDate(date time.Time, lineMeta *LineMeta) error {
	if date.Before(ds.currentDate) {
		return lineMeta.ErrorAt(fmt.Sprintf("date set to an earlier date '%s' than previous date '%s' set at '%s'",
			date.Format(accountingDateFormat),
			ds.currentDate.Format(accountingDateFormat), &ds.currentDateLineMeta))
	}
	ds.currentDate = date
	ds.currentDateLineMeta = *lineMeta
	return nil
}

type Account struct {
	name string
	dr   float64
	cr   float64
}

// TODO(jb): Use this
func (account *Account) Balance() float64 {
	switch account.name[0] {
	case 'a', 'e':
		return account.dr - account.cr
	case 'd', 'i':
		return account.cr - account.dr
	}
	return 0
}

type Alias struct {
	name        string
	accountName string
	lineMeta    LineMeta // Alias defined at this line
}

type Transaction struct {
	date   time.Time
	amount float64
	dr     *Account
	cr     *Account
}

// Parse accounting file and store the information in the data store.
func (ds *DataStore) ParseAccountingFile(fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 1
	for scanner.Scan() {
		line := Line{
			meta: LineMeta{
				fileName:   fileName,
				lineNumber: lineNumber,
			},
			row: bytes.Fields(scanner.Bytes()),
		}
		lineNumber++

		if line.IsComment() {
		} else if line.IsAlias() {
			err = line.ParseAlias(ds)
		} else if line.IsDate() {
			err = line.ParseDate(ds)
		} else if line.IsTransaction() {
			err = line.ParseTransaction(ds)
		} else {
			return line.meta.ErrorAt("invalid syntax")
		}
		if err != nil {
			// Parsing errors
			return err
		}
	}
	if err = scanner.Err(); err != nil {
		// Scanning errors
		return err
	}

	return nil
}

const (
	accountingDateFormat = "2006-01-02"
)

var accountNameRegexp = regexp.MustCompile("^([adei]:[a-z.-]+)$")
var aliasNameRegexp = regexp.MustCompile("^([a-z.-]+)$")

type Line struct {
	meta LineMeta
	row  [][]byte
}

func (line *Line) IsComment() bool {
	if len(line.row) == 0 || line.row[0][0] == '#' {
		return true
	}
	return false
}

func (line *Line) IsAlias() bool {
	if len(line.row) == 3 && string(line.row[0]) == "alias" {
		return true
	}
	return false
}

func (line *Line) IsDate() bool {
	if len(line.row) == 1 && len(line.row[0]) == len(accountingDateFormat) {
		return true
	}
	return false
}

func (line *Line) IsTransaction() bool {
	if len(line.row) == 3 {
		return true
	}
	return false
}

func (line *Line) ParseAlias(ds *DataStore) error {
	aliasName := line.row[1]
	accountName := line.row[2]
	if !aliasNameRegexp.Match(aliasName) {
		return line.meta.ErrorAt("invalid alias name")
	}
	if !accountNameRegexp.Match(accountName) {
		return line.meta.ErrorAt("invalid account name reference by alias")
	}
	return ds.AddAlias(string(aliasName), string(accountName), &line.meta)
}

func (line *Line) ParseDate(ds *DataStore) error {
	date, err := time.Parse(accountingDateFormat, string(line.row[0]))
	if err != nil {
		return line.meta.ErrorAt("invalid date")
	}
	return ds.SetDate(date, &line.meta)
}

func (line *Line) ParseTransaction(ds *DataStore) error {
	amountText := line.row[0]
	for _, v := range line.row[0] {
		// Accept ',' as well as '.' as decimal point.
		if v == ',' {
			amountText = bytes.Replace(amountText, []byte(","), []byte("."), -1)
		}
	}
	amount, err := strconv.ParseFloat(string(amountText), 64)
	if err != nil {
		return line.meta.ErrorAt("invalid value")
	}
	var accounts [2]*Account
	for i, v := range line.row[1:] {
		accountName := string(v)
		if aliasNameRegexp.Match(v) {
			if alias := ds.GetAlias(accountName); alias != nil {
				accountName = alias.accountName
			} else {
				return line.meta.ErrorAt(fmt.Sprintf("referenced alias '%s' is undefined", accountName))
			}
		} else if !accountNameRegexp.Match(v) {
			return line.meta.ErrorAt(fmt.Sprintf("invalid account name '%s'", accountName))
		}
		account := ds.GetAccount(accountName)
		if account == nil {
			account = ds.AddAccount(accountName)
		}
		accounts[i] = account
	}

	transaction := ds.AddTransaction(amount, accounts[0], accounts[1])

	// TODO(jb): This is temporary
	transaction.dr.dr += amount
	transaction.cr.cr += amount

	// TODO(jb): Test logs
	fmt.Printf("%s: %.2f\n\t%s (%.2f)\n\t%s (%.2f)\n",
		transaction.date.Format(accountingDateFormat),
		transaction.amount,
		transaction.dr.name,
		transaction.dr.Balance(),
		transaction.cr.name,
		transaction.cr.Balance())

	return nil
}

type LineMeta struct {
	fileName   string
	lineNumber int
}

func (meta *LineMeta) String() string {
	return fmt.Sprintf("%s:%d", meta.fileName, meta.lineNumber)
}

func (meta *LineMeta) ErrorAt(desc string) error {
	if desc == "" {
		return fmt.Errorf("error: at '%s'", meta)
	} else {
		return fmt.Errorf("error: at '%s', %s", meta, desc)
	}
}
