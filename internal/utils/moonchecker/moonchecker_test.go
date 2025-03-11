package moonchecker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMoonChecker(t *testing.T) {
	ordes := []struct {
		number  string
		isValid bool
	}{
		{number: "9278923470", isValid: true},
		{number: "1234567890", isValid: false},
	}

	for _, o := range ordes {
		t.Run(o.number, func(t *testing.T) {
			isValid := MoonChecker(o.number)
			assert.Equal(t, o.isValid, isValid)
		})
	}
}
