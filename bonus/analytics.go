package analytics

type Sale struct {
	Product          string
	Category         string
	Quantity         int
	UnitPriceKopecks int64
	Cancelled        bool
}

type Report struct {
	RevenueKopecks    int64
	UnitsSold         int
	RevenueByCategory map[string]int64
	TopProduct        string
	InvalidSales      int
}

func uncancelledSales(sales []Sale) []Sale {
	var result = make([]Sale, 0)

	for _, sale := range sales {
		if !sale.Cancelled {
			result = append(result, sale)
		}
	}

	return result
}

func validSales(sales []Sale) []Sale {
	var result = make([]Sale, 0)

	for _, sale := range sales {
		if sale.Category != "" && sale.Product != "" && sale.Quantity > 0 && sale.UnitPriceKopecks > 0 {
			result = append(result, sale)
		}
	}

	return result
}

func compareRunes(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	for i := 0; i < len(ra) && i < len(rb); i++ {
		if ra[i] != rb[i] {
			if ra[i] < rb[i] {
				return -1
			}
			return 1
		}
	}

	return len(ra) - len(rb)
}

func AnalyzeSales(sales []Sale) Report {
	var uncancelled = uncancelledSales(sales)
	var validSales = validSales(uncancelled)

	var revenueKopecks int64 = 0
	var unitsSold = 0

	for _, sale := range validSales {
		unitsSold += sale.Quantity
		revenueKopecks += sale.UnitPriceKopecks * int64(sale.Quantity)
	}

	var revenueByCategory = map[string]int64{}
	var revenueByProduct = map[string]int64{}
	var quantityByProduct = map[string]int{}

	for _, sale := range validSales {
		revenueByCategory[sale.Category] += sale.UnitPriceKopecks * int64(sale.Quantity)
		revenueByProduct[sale.Product] += sale.UnitPriceKopecks * int64(sale.Quantity)
		quantityByProduct[sale.Product] += sale.Quantity
	}

	var topProduct string
	var maxSoled = 0
	for name, quantity := range quantityByProduct {
		if quantity > maxSoled || quantity == maxSoled && compareRunes(name, topProduct) < 0 {
			maxSoled = quantity
			topProduct = name
		}
	}

	return Report{
		RevenueKopecks:    revenueKopecks,
		UnitsSold:         unitsSold,
		RevenueByCategory: revenueByCategory,
		TopProduct:        topProduct,
		InvalidSales:      len(uncancelled) - len(validSales),
	}
}
