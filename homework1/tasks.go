package lesson2

// FizzBuzz возвращает FizzBuzz, Fizz, Buzz или пустую строку для обычного числа.
// Для n <= 0 возвращает invalid.
func FizzBuzz(n int) string {
	switch {
	case n <= 0:
		return "invalid"
	case n%3 == 0 && n%5 != 0:
		return "Fizz"
	case n%5 == 0 && n%3 != 0:
		return "Buzz"
	case n%15 == 0:
		return "FizzBuzz"

	default:
		return ""
	}
}

// Sum возвращает сумму. Для пустого среза и nil возвращает 0.
func Sum(values []int64) int64 {
	if values == nil || len(values) == 0 {
		return 0
	}

	var sum int64 = 0
	for _, value := range values {
		sum += value
	}
	return sum
}

// Min возвращает минимум. Для пустого среза и nil возвращает 0 по условию задачи.
func Min(values []int64) int64 {
	if values == nil || len(values) == 0 {
		return 0
	}

	res := values[0]
	for _, value := range values {
		if value < res {
			res = value
		}
	}
	return res
}

// Max возвращает максимум. Для пустого среза и nil возвращает 0 по условию задачи.
func Max(values []int64) int64 {
	if values == nil || len(values) == 0 {
		return 0
	}

	res := values[0]
	for _, value := range values {
		if value > res {
			res = value
		}
	}
	return res
}

// Average возвращает среднее как float64. Для пустого среза и nil возвращает 0.
func Average(values []int64) float64 {
	if values == nil || len(values) == 0 {
		return 0
	}
	return float64(Sum(values)) / float64(len(values))
}

// Category возвращает invalid, small, regular или large по сумме в копейках.
func Category(amount int64) string {
	switch {
	case 0 <= amount && amount <= 9999:
		return "small"
	case 10000 <= amount && amount <= 49999:
		return "regular"
	case amount >= 50000:
		return "large"
	default:
		return "invalid"
	}
}

// RuneCount считает кодовые точки Unicode в строке через range.
func RuneCount(text string) int {
	cnt := 0
	for range text {
		cnt++
	}
	return cnt
}
