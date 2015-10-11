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
	for _, report := range PrepareReports(data) {
		if err = report.Generate(flags.OutputDir); err != nil {
			Fatal(err)
		}
	}
}
