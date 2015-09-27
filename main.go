package main

func main() {
	flags, err := ParseFlags()
	if err != nil {
		Fatal(err)
	}
	data := NewAccountingData()
	for _, fileName := range flags.AccountingFiles {
		if err = data.ReadFile(fileName); err != nil {
			Fatal(err)
		}
	}
}
