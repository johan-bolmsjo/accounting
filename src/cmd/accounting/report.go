package main

import (
	"fmt"
	"time"
)

//const (
//	dateFormatYear         = "2006"
//	dateFormatYearMonth    = "2006-01"
//	dateFormatYearMonthDay = "2006-01-02"
//)

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
func (report *Report) AddTransaction(t *Transaction) {
	for i, accountName := range t.accounts {
		account := report.GetAccount(accountName)
		account.flat[i] += t.amount

		for ; accountName != ""; accountName = accountName.Parent() {
			account := report.GetAccount(accountName)
			account.cumulative[i] += t.amount
		}
	}
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
	for i, t := range data.Transactions() {
		if i == 0 {
			periods[ReportPeriodYearly] = &Report{
				period:   ReportPeriodYearly,
				from:     time.Date(t.date.Year(), time.January, 1, 0, 0, 0, 0, time.UTC),
				to:       time.Date(t.date.Year()+1, time.January, 1, 0, 0, 0, 0, time.UTC),
				accounts: make(map[AccountName]*Account),
			}
			periods[ReportPeriodMonthly] = &Report{
				period:   ReportPeriodMonthly,
				from:     time.Date(t.date.Year(), t.date.Month(), 1, 0, 0, 0, 0, time.UTC),
				to:       time.Date(t.date.Year(), t.date.Month()+1, 1, 0, 0, 0, 0, time.UTC),
				accounts: make(map[AccountName]*Account),
			}
		}
		for j, _ := range periods {
			periods[j] = periods[j].UntilPeriod(t.date)
			periods[j].AddTransaction(t)
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

// Generate reports to output directory.
func (report *Report) Generate(outputDir string) error {

	// TODO(jb) These are just test logs.

	fmt.Printf("\n\n*** %v\n\n", report.from)

	for _, account := range report.accounts {
		fmt.Printf("%s  %f\n", account.name, account.CumulativeBalance())
	}

	//date.Format(dateFormatYear)
	//date.Format(dateFormatYearMonth)

	return nil
}
