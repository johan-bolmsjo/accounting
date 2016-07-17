package table

import (
	"fmt"
)

func createTestTable(titles, rows, emptyRow bool) *Table {
	t := new(Table)
	if titles {
		t.SetTitles(Row{
			{"Fruit", 0, 0, AlignLeft},
			{"Price", 0, 0, AlignLeft},
			{"Notes", 0, 0, AlignLeft},
		})
	}
	if rows {
		t.AddRow(Row{
			{"Äpple", 0, 0, AlignRight}, // Use UTF-8 character to test padding and alignment
			{"10", 0, 5, AlignRight},
		})
		if emptyRow {
			t.AddRow(Row{})
		} else {
			t.AddRow(Row{
				{"Banana", 0, 0, AlignCenter},
				{"5", 5, 0, AlignRight},
				{"Overripe", 0, 0, AlignLeft},
			})
		}
		t.AddRow(Row{
			{"Red Grapes", 0, 0, AlignLeft},
			{"17", 0, 0, AlignCenter},
			{"", 0, 0, AlignLeft},
			{"Extra column", 0, 0, AlignLeft},
		})
	}

	return t
}

func ExampleTitlesAndRows() {
	fmt.Printf("%s", createTestTable(true, true, false))
	// Output:
	// Fruit      | Price   | Notes    |
	// -----------+---------+----------+-------------
	//      Äpple | 10      |          |
	//   Banana   |       5 | Overripe |
	// Red Grapes |   17    |          | Extra column
}

func ExampleTitlesOnly() {
	fmt.Printf("%s", createTestTable(true, false, false))
	// Output:
	// Fruit | Price | Notes
	// ------+-------+------
}

func ExampleRowsOnly() {
	fmt.Printf("!\n%s", createTestTable(false, true, false))
	// Output:
	// !
	//      Äpple | 10      |          |
	//   Banana   |       5 | Overripe |
	// Red Grapes |   17    |          | Extra column
}

func ExampleEmptyRow() {
	fmt.Printf("%s", createTestTable(true, true, true))
	// Output:
	// Fruit      | Price   | Notes |
	// -----------+---------+-------+-------------
	//      Äpple | 10      |       |
	//            |         |       |
	// Red Grapes |   17    |       | Extra column
}

func ExampleSingleColumn() {
	t := new(Table)
	t.SetTitles(Row{{Content: "Fruit"}})
	t.AddRow(Row{{Content: "Äpple"}})
	t.AddRow(Row{{Content: "Banana"}})
	fmt.Printf("%s", t)
	// Output:
	// Fruit
	// ------
	// Äpple
	// Banana
}
