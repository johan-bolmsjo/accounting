A simple accounting program written in [Go](https://golang.org).

The program produces monthly and yearly text reports over assets, debts,
expenses and incomes read from transaction lists.

Usage
-----

accounting <-o output directory> <input files(s)>

The program takes one mandatory output directory specified using the -o flag and
one or more input text files.

Input format
------------

### Accounting

All transactions must be preceded by a date of the format "YYYY-MM-DD" on a
single line. The transaction format is "amount debits credits". The amount
format is a numeric floating point number with "." or "," as the decimal point.
The debits and credits format is "[adei]:[a-z.-]+". The first letter followed by
the colon indicates the account type. Only four types are supported.

a: Asset   (amount = debits - credits)
d: Debt    (amount = credits - debits)
e: Expense (amount = debite - credits)
i: Income  (amount = credits - debits)

Accounts are grouped by using dots in their names. For example "e:food.snacks"
and "e:food.takeout" will both be included in the "food" expense as well.

Aliases are expanded in the debits and credits columns. Empty lines are allowed
and removed while parsing.

Example:

    alias salary a:account.salary

    2014-12-01
    500  salary

    2014-12-03
    100  e:gifts         salary
    50   e:food.takeout  salary
    25   e:food.snacks   salary

### Aliases

Aliases can be used to shorten transaction account names. For example if the
salary account is used often a shorter alias can be created.

Example:
    alias  salary  a:account.salary

Alias names may not contain the account type prefix codes to avoid ambiguities.

Reports
-------

Four reports are generated, monthly and yearly into the output directory
specified on the command line.

The format shows the grouping used in the account names as a tree. Each node in
the tree shows the sum of all child nodes.

Example:

    expense            | amount
    -------------------+---------
    all                | 75.00
        food           | 75.00
            groceries  | 50.00
            snacks     | 25.00

The amount value in the expense and income reports is for the current report
period while the value in the asset and debt reports is derived from all
transactions.

The asset and debt reports also contains a *period delta* column that shows the
change since the last report period.

Example:

    asset              | amount  | preiod delta
    -------------------+---------+-------------
    all                |  425.00 | +200.00
        account        |  425.00 | +200.00
            salary     |  425.00 | +200.00

All transactions are listed at the end of the report.
