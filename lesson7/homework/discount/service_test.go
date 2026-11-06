package discount

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePercent(t *testing.T) {
	cases := []struct {
		name  string
		input int
		err   error
	}{
		{"hundred", 100, nil},
		{"zero", 0, nil},
		{"ten", 10, nil},
		{"minus one", -1, ErrInvalidPercent},
		{"hundred and one", 101, ErrInvalidPercent},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if c.err == nil {
				assert.NoError(t, ValidatePercent(c.input))
			} else {
				assert.EqualError(t, ValidatePercent(c.input), c.err.Error())
			}
		})
	}
}

func TestApply(t *testing.T) {
	cases := []struct {
		name     string
		price    int
		percent  int
		expected int
		err      error
	}{
		{"hundred-twenty", 100, 20, 80, nil},
		{"zero-percent", 100, 0, 100, nil},
		{"hundred percent", 100, 100, 0, nil},
		{"negative price", -12, 67, 0, ErrNegativePrice},
		{"negative percent", 12, -12, 0, ErrInvalidPercent},
		{"more than hundred percent", 12, 1000, 0, ErrInvalidPercent},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			t.Parallel()

			price, err := Apply(c.price, c.percent)

			if c.err == nil {
				assert.Equal(t, c.expected, price)
			} else {
				assert.EqualError(t, c.err, err.Error())
				assert.Equal(t, c.expected, price)
			}
		})
	}
}

func TestFinalPrice(t *testing.T) {
	cases := []struct {
		name     string
		items    []Item
		expected int
		err      error
	}{
		{"empty", []Item{}, 0, nil},
		{"valid", []Item{{100, 12}, {200, 24}}, 240, nil},
		{"incorrect price", []Item{{100, 12}, {-12, 24}}, 0, ErrNegativePrice},
		{"incorrect percent", []Item{{100, 12}, {12, -12}}, 0, ErrInvalidPercent},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			t.Parallel()

			total, err := FinalPrice(c.items)

			if c.err == nil {
				assert.Equal(t, c.expected, total)
			} else {
				assert.EqualError(t, c.err, err.Error())
				assert.Equal(t, c.expected, total)
			}
		})
	}
}
