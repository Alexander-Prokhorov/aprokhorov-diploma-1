package verificator

type Luhn struct{}

func NewLuhn() (*Luhn, error) {
	return &Luhn{}, nil
}

// Valid check number is valid or not based on Luhn algorithm
func (l *Luhn) Valid(number int64) bool {
	return (number%10+l.checksum(number/10))%10 == 0
}

func (l *Luhn) checksum(number int64) int64 {
	var luhn int64

	for i := 0; number > 0; i++ {
		cur := number % 10

		if i%2 == 0 { // even
			cur = cur * 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}

		luhn += cur
		number = number / 10
	}
	return luhn % 10
}
