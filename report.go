package main

import (
	"fmt"
	"time"
)

type reportPeriod int

const (
	reportPeriodYearly reportPeriod = iota
	reportPeriodMonthly
	reportPeriodCount
)

// Returns the number of months the report period corresponds to.
func (period reportPeriod) months() int {
	switch period {
	case reportPeriodYearly:
		return 12
	case reportPeriodMonthly:
		return 1
	}
	return 0
}

type Report struct {
	prev         *Report
	period       reportPeriod
	from, to     time.Time // Report interval [from, to)
	transactions []*Transaction
	accounts     map[AccountName]*Account
}

// Adds transaction to log and updates accounts.
func (report *Report) addTransaction(t *Transaction) {
	for i, accountName := range t.accounts {
		account := report.getAccount(accountName)
		account.flat[i] += t.amount

		for ; accountName != ""; accountName = accountName.Parent() {
			account := report.getAccount(accountName)
			account.cumulative[i] += t.amount
		}
	}
}

// Lookup existing or add new account to the report.
func (report *Report) getAccount(name AccountName) *Account {
	if account, ok := report.accounts[name]; ok {
		return account
	}
	account := &Account{name: name}
	report.accounts[name] = account
	return account
}

// Copy accounts of the specified type from another report.
func (report *Report) copyAccountsOfType(from *Report, typ AccountType) {
	for k, v := range from.accounts {
		if k.Type() == typ {
			accountCopy := *v
			report.accounts[k] = &accountCopy
		}
	}
}

// Get the report for the next period.
// The old report becomes the previous report of the new report.
func (report *Report) nextPeriod() *Report {
	nextToMonth := report.to.Month() + time.Month(report.period.months())
	nextReport := &Report{
		prev:     report,
		period:   report.period,
		from:     report.to,
		to:       time.Date(report.to.Year(), nextToMonth, 1, 0, 0, 0, 0, time.UTC),
		accounts: make(map[AccountName]*Account),
	}
	nextReport.copyAccountsOfType(report, AccountTypeAsset)
	nextReport.copyAccountsOfType(report, AccountTypeDebt)
	return nextReport
}

// Create reports until arriving on time.
// Returns the report corresponding to time.
func (report *Report) untilPeriod(curr time.Time) *Report {
	for curr.After(report.to) || curr.Equal(report.to) {
		report = report.nextPeriod()
	}
	return report
}

// Account data stored in reports.
type Account struct {
	name       AccountName
	flat       [drcr]float64 // Amounts accounted specifically to this account
	cumulative [drcr]float64 // Amounts accounted to this account and all child accounts.
}

func balance(typ AccountType, val *[drcr]float64) float64 {
	switch typ {
	case AccountTypeAsset, AccountTypeExpense:
		return val[dr] - val[cr]
	case AccountTypeDebt, AccountTypeIncome:
		return val[cr] - val[dr]
	}
	return 0
}

func (account *Account) flatBalance() float64 {
	return balance(account.name.Type(), &account.flat)
}

func (account *Account) cumulativeBalance() float64 {
	return balance(account.name.Type(), &account.cumulative)
}

// Prepare reports from accounting data.
func PrepareReports(data *AccountingData) []*Report {
	var periods [reportPeriodCount]*Report
	for i, t := range data.Transactions() {
		if i == 0 {
			periods[reportPeriodYearly] = &Report{
				period:   reportPeriodYearly,
				from:     time.Date(t.date.Year(), time.January, 1, 0, 0, 0, 0, time.UTC),
				to:       time.Date(t.date.Year()+1, time.January, 1, 0, 0, 0, 0, time.UTC),
				accounts: make(map[AccountName]*Account),
			}
			periods[reportPeriodMonthly] = &Report{
				period:   reportPeriodMonthly,
				from:     time.Date(t.date.Year(), t.date.Month(), 1, 0, 0, 0, 0, time.UTC),
				to:       time.Date(t.date.Year(), t.date.Month()+1, 1, 0, 0, 0, 0, time.UTC),
				accounts: make(map[AccountName]*Account),
			}
		}
		for j, _ := range periods {
			periods[j] = periods[j].untilPeriod(t.date)
			periods[j].addTransaction(t)
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
		fmt.Printf("%s  %f\n", account.name, account.cumulativeBalance())
	}

	//date.Format(dateFormatYear)
	//date.Format(dateFormatYearMonth)

	return nil
}
