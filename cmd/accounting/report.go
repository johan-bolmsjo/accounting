package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"table"
	"time"
)

type ReportPeriod int

const (
	ReportPeriodYearly ReportPeriod = iota
	ReportPeriodMonthly
	ReportPeriodCount
)

// Returns the number of months the report period corresponds to.
func (period ReportPeriod) Months() int {
	switch period {
	case ReportPeriodYearly:
		return 12
	case ReportPeriodMonthly:
		return 1
	}
	return 0
}

type Report struct {
	prev         *Report
	period       ReportPeriod
	from, to     time.Time // Report interval [from, to)
	transactions []*Transaction
	accounts     map[AccountName]*Account
}

// Adds transaction to log and updates accounts.
func (report *Report) AddTransaction(tr *Transaction) {
	for i, accountName := range tr.accounts {
		account := report.GetAccount(accountName)
		account.flat[i] += tr.amount

		for ; accountName != ""; accountName = accountName.Parent() {
			account := report.GetAccount(accountName)
			account.cumulative[i] += tr.amount
		}
	}
	report.transactions = append(report.transactions, tr)
}

// Lookup existing or add new account to the report.
func (report *Report) GetAccount(name AccountName) *Account {
	if account, ok := report.accounts[name]; ok {
		return account
	}
	account := &Account{name: name}
	report.accounts[name] = account
	return account
}

// Copy accounts of the specified type from another report.
func (report *Report) CopyAccountsOfType(from *Report, typ AccountType) {
	for k, v := range from.accounts {
		if k.Type() == typ {
			accountCopy := *v
			report.accounts[k] = &accountCopy
		}
	}
}

// Get the report for the next period.
// The old report becomes the previous report of the new report.
func (report *Report) NextPeriod() *Report {
	nextToMonth := report.to.Month() + time.Month(report.period.Months())
	nextReport := &Report{
		prev:     report,
		period:   report.period,
		from:     report.to,
		to:       time.Date(report.to.Year(), nextToMonth, 1, 0, 0, 0, 0, time.UTC),
		accounts: make(map[AccountName]*Account),
	}
	nextReport.CopyAccountsOfType(report, AccountTypeAsset)
	nextReport.CopyAccountsOfType(report, AccountTypeDebt)
	return nextReport
}

// Create reports until arriving on time.
// Returns the report corresponding to time.
func (report *Report) UntilPeriod(curr time.Time) *Report {
	for curr.After(report.to) || curr.Equal(report.to) {
		report = report.NextPeriod()
	}
	return report
}

// Return the account delta computed from the cumulative balance of the current
// and previous report.
func (report *Report) AccountDelta(name AccountName) float64 {
	var curr, prev float64

	if account := report.accounts[name]; account != nil {
		curr = account.CumulativeBalance()
	}
	if report.prev != nil {
		if account := report.prev.accounts[name]; account != nil {
			prev = account.CumulativeBalance()
		}
	}
	return curr - prev
}

// Account data stored in reports.
type Account struct {
	name       AccountName
	flat       [DrCr]float64 // Amounts accounted specifically to this account
	cumulative [DrCr]float64 // Amounts accounted to this account and all child accounts.
}

func balance(typ AccountType, val *[DrCr]float64) float64 {
	switch typ {
	case AccountTypeAsset, AccountTypeExpense:
		return val[Dr] - val[Cr]
	case AccountTypeDebt, AccountTypeIncome:
		return val[Cr] - val[Dr]
	}
	return 0
}

func (account *Account) FlatBalance() float64 {
	return balance(account.name.Type(), &account.flat)
}

func (account *Account) CumulativeBalance() float64 {
	return balance(account.name.Type(), &account.cumulative)
}

// Prepare reports from accounting data.
func PrepareReports(data *AccountingData) []*Report {
	var periods [ReportPeriodCount]*Report
	for i, tr := range data.Transactions() {
		if i == 0 {
			periods[ReportPeriodYearly] = &Report{
				period:   ReportPeriodYearly,
				from:     time.Date(tr.date.Year(), time.January, 1, 0, 0, 0, 0, time.UTC),
				to:       time.Date(tr.date.Year()+1, time.January, 1, 0, 0, 0, 0, time.UTC),
				accounts: make(map[AccountName]*Account),
			}
			periods[ReportPeriodMonthly] = &Report{
				period:   ReportPeriodMonthly,
				from:     time.Date(tr.date.Year(), tr.date.Month(), 1, 0, 0, 0, 0, time.UTC),
				to:       time.Date(tr.date.Year(), tr.date.Month()+1, 1, 0, 0, 0, 0, time.UTC),
				accounts: make(map[AccountName]*Account),
			}
		}
		for j, _ := range periods {
			periods[j] = periods[j].UntilPeriod(tr.date)
			periods[j].AddTransaction(tr)
		}
	}

	var reports []*Report
	for _, report := range periods {
		for ; report != nil; report = report.prev {
			reports = append(reports, report)
		}
	}
	return reports
}

// Sort interface to sort by account name.
type sortByAccountName []*Account

func (a sortByAccountName) Len() int      { return len(a) }
func (a sortByAccountName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a sortByAccountName) Less(i, j int) bool {
	// Substitute '.' with ' ' when sorting to get the account grouping right.
	return strings.Replace(string(a[i].name), ".", " ", -1) < strings.Replace(string(a[j].name), ".", " ", -1)
}

const indentAmount = 4

func balanceToString(v float64) string {
	s := fmt.Sprintf("%.2f", v)
	if s == "0.00" {
		s = "-"
	}
	return s
}

// Generate reports to output directory.
func (report *Report) Generate(outputDir string) error {
	var buf bytes.Buffer
	var filename string

	switch report.period {
	case ReportPeriodMonthly:
		fmt.Fprintf(&buf, "%s %d\n\n", report.from.Month(), report.from.Year())
		filename = filepath.Join(outputDir, fmt.Sprintf("%d-%02d.txt", report.from.Year(), report.from.Month()))
	case ReportPeriodYearly:
		fmt.Fprintf(&buf, "%d\n\n", report.from.Year())
		filename = filepath.Join(outputDir, fmt.Sprintf("%d.txt", report.from.Year()))
	default:
		return fmt.Errorf("unsupported report period %d", report.period)
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var accounts []*Account
	for _, account := range report.accounts {
		accounts = append(accounts, account)
	}
	sort.Sort(sortByAccountName(accounts))

	// Account summary
	t := new(table.Table)
	t.SetTitles(table.Row{
		{Content: "account"},
		{Content: "amount"},
		{Content: "cumulative"},
		{Content: "delta"},
	})

	for _, account := range accounts {
		cumulativeStr := balanceToString(account.CumulativeBalance())
		deltaStr := fmt.Sprintf("%+.2f", report.AccountDelta(account.name))

		if cumulativeStr != "-" || deltaStr != "+0.00" {
			t.AddRow(table.Row{
				{Content: account.name.Leaf(), PadLeft: uint(indentAmount * account.name.Depth())},
				{Content: balanceToString(account.FlatBalance()), Align: table.AlignRight},
				{Content: cumulativeStr, Align: table.AlignRight},
				{Content: deltaStr, Align: table.AlignRight},
			})
		}
	}
	buf.Write(t.RenderText())

	// Transaction log
	fmt.Fprintf(&buf, "\nTransactions\n\n")

	t = new(table.Table)
	t.SetTitles(table.Row{
		{Content: "date"},
		{Content: "account"},
		{Content: "debit"},
		{Content: "credit"},
	})

	var prevDate time.Time
	for _, tr := range report.transactions {
		var dateStr string
		if tr.date != prevDate {
			dateStr = tr.date.Format(transactionDateFormat)
		}
		prevDate = tr.date

		t.AddRow(table.Row{
			{Content: dateStr},
			{Content: string(tr.accounts[Dr])},
			{Content: fmt.Sprintf("%.2f", tr.amount), Align: table.AlignRight},
			{Content: ""},
		})
		t.AddRow(table.Row{
			{Content: ""},
			{Content: string(tr.accounts[Cr])},
			{Content: ""},
			{Content: fmt.Sprintf("%.2f", tr.amount), Align: table.AlignRight},
		})
	}
	buf.Write(t.RenderText())

	fmt.Fprintf(file, "%s", buf.Bytes())
	return nil
}
