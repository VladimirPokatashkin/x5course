package homework

// PairInput соответствует объекту nums/target из условия контеста.
type PairInput struct {
	Nums   []int
	Target int
}

// longestPrefix возвращает самый длинный общий префикс строк.
func longestPrefix(words []string) string {
	if words == nil || len(words) == 0 {
		return ""
	}

	var first = []rune(words[0])
	var right = 0

	for range first {
		var stopped = false

		for i := 1; i < len(words); i++ {
			var word = []rune(words[i])

			if right >= len(word) || word[right] != first[right] {
				stopped = true
				break
			}
		}

		if stopped {
			break
		}
		right++

	}

	return string(first[:right])
}

// pairIndices возвращает индексы двух разных элементов по возрастанию.
func pairIndices(input PairInput) [2]int {
	var m = make(map[int]int)

	for i, num := range input.Nums {
		if _, ok := m[input.Target-num]; ok {
			return [2]int{m[input.Target-num], i}
		}
		m[num] = i
	}

	return [2]int{0, 0}
}

// hasDuplicate сообщает, встречается ли какое-либо значение повторно.
func hasDuplicate(nums []int) bool {
	var m = make(map[int]struct{})

	for _, num := range nums {
		if _, ok := m[num]; ok {
			return true
		}
		m[num] = struct{}{}
	}

	return false
}

// compactSorted создаёт независимый слайс уникальных значений.
func compactSorted(nums []int) []int {
	if nums == nil || len(nums) == 0 {
		return []int{}
	}

	var res = make([]int, len(nums))
	res[0] = nums[0]
	var left = 0

	for right := 1; right < len(nums); right++ {
		if nums[right] == res[left] {
			continue
		}

		left++
		res[left] = nums[right]
	}

	return res[:left+1]
}

// balancedBrackets проверяет пары и порядок закрытия скобок.
func balancedBrackets(text string) bool {
	var stack []rune

	for _, c := range text {
		switch c {
		case '(', '[', '{':
			stack = append(stack, c)
		case ')':
			if len(stack) == 0 || stack[len(stack)-1] != '(' {
				return false
			}
			stack = stack[:len(stack)-1]
		case ']':
			if len(stack) == 0 || stack[len(stack)-1] != '[' {
				return false
			}
			stack = stack[:len(stack)-1]
		case '}':
			if len(stack) == 0 || stack[len(stack)-1] != '{' {
				return false
			}
			stack = stack[:len(stack)-1]
		}
	}

	return len(stack) == 0
}
