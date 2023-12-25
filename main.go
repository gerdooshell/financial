package main

import (
	"fmt"
	"github.com/gerdooshell/financial/environment"
	"github.com/gerdooshell/financial/investment"
	"github.com/gerdooshell/financial/tax/canada/canada_region"
	"github.com/gerdooshell/financial/tax/canada/rrsp"
	"github.com/gerdooshell/financial/tax/canada/total_tax"
	"runtime"
	"sync"
	"time"
)

func main() {
	investment.PrintReport(10, 100000, 0.2)
	//runBulkTaxReport()
	//getRRspProcessed()
	//pm := password_manager.NewPasswordManager()
	//fmt.Println(pm.GeneratePassword(61)) // wfCxfLUkZ_d9BMg@LNJX
}

func getRRspProcessed() {
	if err := environment.SetEnvironment(environment.Dev); err != nil {
		panic(err)
	}
	salaries := []float32{400_000, 350_000, 300_000, 270_000, 220_000, 180_000, 145_000, 120_000, 100_000}
	inputs := make([]*rrsp.OptimalReturnsInputParameters, 0, len(salaries))
	for _, salary := range salaries {
		inputs = append(inputs,
			&rrsp.OptimalReturnsInputParameters{Year: 2023, Province: canada_region.BritishColumbia, Salary: salary})
	}
	rrSavingPlan := rrsp.NewRRetirementSavingPlan()
	outChan, err := rrSavingPlan.GetOptimalReturns(inputs)
	if err != nil {
		panic(err)
	}
	for out := range outChan {
		for _, param := range out {
			fmt.Println("Year:", param.Year)
			fmt.Println("Province:", param.Province)
			fmt.Println("Salary:", param.Salary)
			fmt.Println("Tax on Salary:", param.TaxOnSalary)
			fmt.Println("IsEmployer:", param.IsEmployer)
			fmt.Println("Net Remained:", param.NetRemained)
			fmt.Println("EdgeSalary:", param.EdgeSalary)
			fmt.Println("Tax on Edge:", param.TaxOnEdge)
			fmt.Println("Contribution:", param.Contribution)
			fmt.Println("Tax Return:", param.TaxReturn)
			fmt.Println("-------------------------------")
		}
		fmt.Println("+++++++++++++++++++++++++++++++")
		fmt.Println("+++++++++++++++++++++++++++++++")
	}
}

func applyRateLimit(channel <-chan []*total_tax.InputParameters) {
	waitingTime := time.NewTicker(time.Microsecond * 10)
	wg := sync.WaitGroup{}
	count := 0
	rateCount := 0
	rate := 0
	t0 := time.Now()
	for inputs := range channel {
		count++
		rateCount++
		if time.Since(t0) > time.Second {
			t0 = time.Now()
			rate = rateCount
			rateCount = 0
			fmt.Println("num goroutines", runtime.NumGoroutine(), "; total num reqs processed:", count, "; num req per second (rate):", rate)
		}
		for i := 0; i < runtime.NumGoroutine()/1000; i++ {
			<-waitingTime.C
		}
		wg.Add(1)
		go getTaxReport(&wg, inputs)
	}
	wg.Wait()
}

func runBulkTaxReport() {
	tStart := time.Now()
	if err := environment.SetEnvironment(environment.Dev); err != nil {
		panic(err)
	}
	count := 50_000
	channel := make(chan []*total_tax.InputParameters, 10_000)
	reqInputs := make([]*total_tax.InputParameters, 0, 100)
	for i := 0; i < 5; i++ {
		reqInputs = append(reqInputs, []*total_tax.InputParameters{
			{Year: 2023, Salary: 50000, Province: canada_region.BritishColumbia},
			{Year: 2023, Salary: 100000, Province: canada_region.BritishColumbia},
			{Year: 2023, Salary: 200000, Province: canada_region.BritishColumbia},
			{Year: 2023, Salary: 300000, Province: canada_region.BritishColumbia},
			{Year: 2023, Salary: 400000, Province: canada_region.BritishColumbia},
			{Year: 2023, Salary: 500000, Province: canada_region.BritishColumbia},
			{Year: 2023, Salary: 600000, Province: canada_region.BritishColumbia},
			{Year: 2023, Salary: 700000, Province: canada_region.BritishColumbia},
			{Year: 2023, Salary: 800000, Province: canada_region.BritishColumbia},
			{Year: 2023, Salary: 900000, Province: canada_region.BritishColumbia},
		}...)
	}
	go func() {
		defer close(channel)
		for i := 0; i < count; i++ {
			channel <- reqInputs
		}
	}()
	applyRateLimit(channel)

	fmt.Println("processing time:", time.Since(tStart))
}

func getTaxReport(wg *sync.WaitGroup, inputs []*total_tax.InputParameters) <-chan *total_tax.OutputParameters {
	defer wg.Done()
	totalTaxCalculator := total_tax.NewTotalTaxCalculator()

	taxChan, err := totalTaxCalculator.CalculateTotalTax(inputs)
	if err != nil {
		panic(err)
	}

	for _ = range taxChan {
		time.Sleep(time.Millisecond * 1)
		//fmt.Println("Year:", taxReport.Year)
		//fmt.Println("Province:", taxReport.Province)
		//fmt.Println("Salary:", taxReport.Salary)
		//fmt.Println("EI:", taxReport.EI)
		//fmt.Println("CPP:", taxReport.CPP)
		//fmt.Println("ProvincialTax:", taxReport.ProvincialTax)
		//fmt.Println("ProvincialTaxRate:", taxReport.ProvincialTaxRate)
		//fmt.Println("Provincial BPA:", taxReport.ProvincialBPA)
		//fmt.Println("FederalTax:", taxReport.FederalTax)
		//fmt.Println("FederalTaxRate:", taxReport.FederalTaxRate)
		//fmt.Println("Federal BPA:", taxReport.FederalBPA)
		//fmt.Println("TotalTax:", taxReport.TotalTax)
		//fmt.Println("Total tax rate:", taxReport.TaxRate)
		//fmt.Println("NetIncome:", taxReport.NetIncome)
		//fmt.Println("===========")
	}
	return taxChan
}
