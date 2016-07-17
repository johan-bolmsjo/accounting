/*
Package table provides a text table generator.

Exammple:

	t := new(Table)
	t.SetTitles(Row{
		{Content: "Fruit"},
		{Content: "Price"},
	})
	t.AddRow(Row{
		{Content: "Apple"},
		{"10", 0, 0, AlignRight},
	})
	t.AddRow(Row{
		{Content: "Banana"},
		{"10", 0, 0, AlignRight},
	})
	fmt.Printf("%s", t)

Output:
	Fruit  | Price
	-------+------
	Apple  |    10
	Banana |    12
*/
package table
