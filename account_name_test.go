package main

import (
	"fmt"
)

func ExampleAccountNameParent() {
	name := AccountName("e:food.snacks")
	for name != "" {
		fmt.Printf("%s ", name)
		name = name.Parent()
	}
	// Output:
	// e:food.snacks e:food e:
}

func ExampleAccountNameLeaf() {
	names := []AccountName{
		"e:food.snacks", "e:food", "e:",
	}
	for _, name := range names {
		fmt.Printf("%s ", name.Leaf())
	}
	// Output:
	// snacks food expense
}

func ExampleAccountNameType() {
	names := []AccountName{
		"a:", "d:", "e:", "i:", "z:", "",
	}
	for _, name := range names {
		fmt.Printf("%s ", name.Type())
	}
	// Output:
	// asset debt expense income none none
}

func ExampleAccountNameValid() {
	names := []AccountName{
		"a:", "a:account", "a:account.salary",
		"", "a", "a.", "a:.", "a: ", "a:account.", "a:.account", "z:",
	}
	for i, name := range names {
		fmt.Printf("%d:%v ", i, name.Valid())
	}
	// Output:
	// 0:true 1:true 2:true 3:false 4:false 5:false 6:false 7:false 8:false 9:false 10:false
}
