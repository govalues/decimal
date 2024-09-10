package decimal

import "fmt"

// MustAdd is like [Add] but panics if computing error.
func (d Decimal) MustAdd(e Decimal) Decimal {
	f, err := d.Add(e)
	if err != nil {
		panic(fmt.Sprintf("MustAdd(%v) failed: %v", d, err))
	}
	return f
}

// MustSub is like [Sub] but panics if computing error.
func (d Decimal) MustSub(e Decimal) Decimal {
	f, err := d.Sub(e)
	if err != nil {
		panic(fmt.Sprintf("MustSub(%v) failed: %v", d, err))
	}
	return f
}

// MustMul is like [Mul] but panics if computing error.
func (d Decimal) MustMul(e Decimal) Decimal {
	f, err := d.Mul(e)
	if err != nil {
		panic(fmt.Sprintf("MustMul(%v) failed: %v", d, err))
	}
	return f
}

// MustQuo is like [Quo] but panics if computing error.
func (d Decimal) MustQuo(e Decimal) Decimal {
	f, err := d.Quo(e)
	if err != nil {
		panic(fmt.Sprintf("MustQuo(%v) failed: %v", d, err))
	}
	return f
}
