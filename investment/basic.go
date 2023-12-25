package investment

import "fmt"

type ProfitReport struct {
	Year                     int
	ProfitGainedInPeriod     float32
	ProfitToTopUpRatioPeriod float32
	TotalProfitGained        float32
	ProfitToTopUpRatio       float32
	TotalMoney               float32
}

func CalcProfit(years int, yearlyTopUp, avgInterest float32) []ProfitReport {
	periodLength := 5
	numOutput := int(float32(years/periodLength) + 0.99)
	accumulatedPeriod := make([]ProfitReport, 0, numOutput)
	var money, moneyWithoutProfitInPeriod, moneyWithoutProfit float32
	var profitToTopUpRatioPeriod float32 = 1
	var profitToTopUpRatio float32 = 1
	for yearCount := 1; yearCount <= years; yearCount++ {
		money = money*(1+avgInterest) + yearlyTopUp
		moneyWithoutProfitInPeriod += yearlyTopUp
		moneyWithoutProfit += yearlyTopUp
		profitToTopUpRatioPeriod = (money - moneyWithoutProfitInPeriod) / moneyWithoutProfitInPeriod
		profitToTopUpRatio = (money - moneyWithoutProfit) / moneyWithoutProfit
		if yearCount%periodLength == 0 || yearCount == years {
			pr := ProfitReport{
				Year:                     yearCount,
				ProfitGainedInPeriod:     money - moneyWithoutProfitInPeriod,
				ProfitToTopUpRatioPeriod: profitToTopUpRatioPeriod,
				TotalProfitGained:        money - moneyWithoutProfit,
				ProfitToTopUpRatio:       profitToTopUpRatio,
				TotalMoney:               money,
			}
			accumulatedPeriod = append(accumulatedPeriod, pr)
			moneyWithoutProfitInPeriod = money
		}
	}
	return accumulatedPeriod
}

func PrintReport(yearsOfInvestment int, yearlyTopUp, fixedInterestRate float32) {
	if fixedInterestRate > 1 {
		fixedInterestRate /= 100
	}
	accumulated := CalcProfit(yearsOfInvestment, yearlyTopUp, fixedInterestRate)
	fmt.Printf("If you top up %v every year for %v years to RRSP, considering %v%% fixed interest rate \n", yearlyTopUp, yearsOfInvestment, float32(int(fixedInterestRate*1000+0.5))/10)
	for _, profitReport := range accumulated {
		fmt.Printf("year: %v, profit in period: $%v, pr 2 top up ratio period: %v%%, total profit to date: $%v, pr 2 top up ratio total: %v%%, total money: $%v\n",
			profitReport.Year, profitReport.ProfitGainedInPeriod, toPercentage(profitReport.ProfitToTopUpRatioPeriod),
			profitReport.TotalProfitGained, toPercentage(profitReport.ProfitToTopUpRatio), profitReport.TotalMoney)
	}
}

func toPercentage(value float32) float32 {
	return float32(int(value*1000+0.5)) / 10
}
