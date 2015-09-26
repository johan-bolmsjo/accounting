package main

func main() {
	flags, err := ParseFlags()
	if err != nil {
		Fatal(err)
	}
	ds := NewDataStore()
	for _, fileName := range flags.AccountingFiles {
		if err = ds.ParseAccountingFile(fileName); err != nil {
			Fatal(err)
		}
	}
}
