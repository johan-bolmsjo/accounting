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

// Validated data from accounting files.
// Aliases referenced in transactions has been expanded, i.e. transactions only
// reference account names.
type AccountingData struct {
	aliases             map[string]*Alias
	transactions        []*Transaction
	currentDate         time.Time
	currentDateLineMeta LineMeta
}

type Alias struct {
	name     string
	account  AccountName
	lineMeta LineMeta
}

const (
	Dr   = 0
	Cr   = 1
	DrCr = 2
)

type Transaction struct {
	date     time.Time
	amount   float64
	accounts [DrCr]AccountName
}

func NewAccountingData() *AccountingData {
	return &AccountingData{
		aliases: make(map[string]*Alias),
	}
}

func (data *AccountingData) SetDate(date time.Time, lineMeta *LineMeta) error {
	if date.Before(data.currentDate) {
		return lineMeta.ErrorAt(fmt.Sprintf("date set to an earlier date '%s' than previous date '%s' set at '%s'",
			date.Format(transactionDateFormat),
			data.currentDate.Format(transactionDateFormat), &data.currentDateLineMeta))
	}
	data.currentDate = date
	data.currentDateLineMeta = *lineMeta
	return nil
}

func (data *AccountingData) GetDate() time.Time {
	return data.currentDate
}

func (data *AccountingData) AddAlias(alias string, account AccountName, lineMeta *LineMeta) error {
	if t := data.aliases[alias]; t != nil {
		return lineMeta.ErrorAt(fmt.Sprintf("alias '%s' redefined, first seen at '%s'",
			alias, &t.lineMeta))
	}
	data.aliases[alias] = &Alias{name: alias, account: account, lineMeta: *lineMeta}
	return nil
}

func (data *AccountingData) GetAlias(aliasName string) *Alias {
	return data.aliases[aliasName]
}

func (data *AccountingData) AddTransaction(transaction *Transaction) {
	data.transactions = append(data.transactions, transaction)
}

// Get all transactions parsed from accounting files.
// This is the sole data used by the report generator.
func (data *AccountingData) Transactions() []*Transaction {
	return data.transactions
}

// Read accounting file and store the information in data collection.
func (data *AccountingData) ReadFile(fileName string) error {
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
				file: fileName,
				line: lineNumber,
			},
			row: bytes.Fields(scanner.Bytes()),
		}
		lineNumber++

		if line.IsComment() {
		} else if line.IsAlias() {
			err = line.ParseAlias(data)
		} else if line.IsDate() {
			err = line.ParseDate(data)
		} else if line.IsTransaction() {
			err = line.ParseTransaction(data)
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
	transactionDateFormat = "2006-01-02"
)

var aliasNameRegexp = regexp.MustCompile(`^[\p{L}\p{N}-]+$`)

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
	if len(line.row) == 1 && len(line.row[0]) == len(transactionDateFormat) {
		return true
	}
	return false
}

func (line *Line) IsTransaction() bool {
	if len(line.row) == (1 + DrCr) {
		return true
	}
	return false
}

func (line *Line) ParseAlias(data *AccountingData) error {
	alias := line.row[1]
	account := AccountName(line.row[2])
	if !aliasNameRegexp.Match(alias) {
		return line.meta.ErrorAt("invalid alias name")
	}
	if !account.Valid() {
		return line.meta.ErrorAt("invalid account name reference by alias")
	}
	return data.AddAlias(string(alias), AccountName(account), &line.meta)
}

func (line *Line) ParseDate(data *AccountingData) error {
	date, err := time.Parse(transactionDateFormat, string(line.row[0]))
	if err != nil {
		return line.meta.ErrorAt("invalid date")
	}
	return data.SetDate(date, &line.meta)
}

func (line *Line) ParseTransaction(data *AccountingData) error {
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

	transaction := &Transaction{
		date:   data.GetDate(),
		amount: amount,
	}
	if transaction.date.IsZero() {
		return line.meta.ErrorAt("transaction without previous date")
	}

	for i, account := range line.row[1:] {
		transaction.accounts[i] = AccountName(account)
		if aliasNameRegexp.Match(account) {
			if alias := data.GetAlias(string(account)); alias != nil {
				transaction.accounts[i] = alias.account
			} else {
				return line.meta.ErrorAt(fmt.Sprintf("referenced alias '%s' is undefined", account))
			}
		} else if !transaction.accounts[i].Valid() {
			return line.meta.ErrorAt(fmt.Sprintf("invalid account name '%s'", account))
		}
	}

	data.AddTransaction(transaction)
	return nil
}

type LineMeta struct {
	file string
	line int
}

func (meta *LineMeta) String() string {
	return fmt.Sprintf("%s:%d", meta.file, meta.line)
}

func (meta *LineMeta) ErrorAt(desc string) error {
	if desc == "" {
		return fmt.Errorf("error: at '%s'", meta)
	} else {
		return fmt.Errorf("error: at '%s', %s", meta, desc)
	}
}
