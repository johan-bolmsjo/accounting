A simple accounting program written in [Go](https://golang.org).

The program produces monthly and yearly text reports over assets, debts,
expenses and incomes read from transaction lists.

Please see the LICENSE file for details about licensing.

Building
========

[Go](https://golang.org) must be installed to build the software.

Once that is taken care of the program can be installed in `~/bin` as follows.

    go get github.com/johan-bolmsjo/accounting/...
        OR
    go install

Usage
=====

`accounting -o output-dir accounting-files`

The program takes one mandatory output directory specified using the `-o` flag and
one or more input text files.

Input format
============

Accounting
----------

All transactions must be preceded by a date of the format `"YYYY-MM-DD"` on a
single line. The transaction format is `"amount debit-account credit-account"`.
The amount format is a numeric floating point number with `"."` or `","` as the
decimal point. The account format is specified by the regexp
`"^[adei]:([\p{L}\p{N}-]+(\.[\p{L}\p{N}-]+)*|)$"`. The first letter followed by
the colon indicates the account type. Only four types are supported.

    a: Asset   (amount = debits - credits)
    d: Debt    (amount = credits - debits)
    e: Expense (amount = debite - credits)
    i: Income  (amount = credits - debits)

Accounts are grouped by using dots in their names. For example `"e:food.snacks"`
and `"e:food.takeout"` will both be included in the cumulative value of the
`"food"` expense.

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

Aliases
-------

Aliases can be used to shorten (often used) account names. The alias name is
specified by the regexp `"^[\p{L}\p{N}-]+$"`.

Example:

    alias  salary  a:account.salary

Alias names may not contain the account type prefix codes to avoid ambiguities.

Reports
=======

Reports are generated, monthly and yearly into the output directory specified on
the command line. Each report consist of two parts. A table over all accounts
followed by the transactions for the period.

The account table shows the grouping used in the account names as a tree. The
cumulative column includes the amount of child nodes.

Example:

    January 2012

    account            | amount | cumulative | delta
    -------------------+--------+------------+--------
    asset              |      - |     425.00 | +200.00
        account        |      - |     425.00 | +200.00
            salary     | 425.00 |     425.00 | +200.00
    expense            |      - |      92.00 |  +15.00
        food           |  15.00 |      82.00 |  +10.00
            groceries  |  50.00 |      50.00 |  +25.00
            snacks     |  25.00 |      25.00 |    0.00
            takeout    |   7.00 |       7.00 |  -20.00
        drinks         |  10.00 |      10.00 |    0.00

Values in expense and income accounts are based on transactions from the current
report period while values in asset and debt accounts are based on all previous
transactions. The delta column show change of the cumulative value compared to
the previous report period. At the moment delta values are not shown for expense
and income accounts without transactions for the current period. This was tried
but was found to be too cluttered. Accounts with zero cumulative and delta
values are suppressed.

All transactions are listed at the end of the report.

Example:

    date       | account          | debit  | credit
    -----------+------------------+--------+-------
    2012-01-01 | e:food.snacks    |  25.00 |
               | a:account.salary |        |  25.00
