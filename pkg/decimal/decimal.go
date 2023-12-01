package decimal

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cockroachdb/apd/v3"
	jsoniter "github.com/json-iterator/go"
)

// Decimal is a wrapper around apd.Decimal that implements the json.Marshaler and json.Unmarshaler interfaces.
type Decimal struct {
	apd.Decimal
}

const (
	MAX_PRECISION    = 21 // total number of digits, before and after decimal points
	MAX_EXPONENT     = 13 // total number of digits, after decimal points
	MIN_EXPONENT     = -8 // total number of digits, before decimal points
	DEFAULT_ROUNDING = apd.RoundDown
)

var (
	apdCtx = &apd.Context{
		Precision:   MAX_PRECISION,
		MaxExponent: MAX_EXPONENT,
		MinExponent: MIN_EXPONENT,
		Rounding:    DEFAULT_ROUNDING,
		Traps:       apd.DefaultTraps,
	}
	Ctx     = &context{apdCtx}
	jsonSTD = jsoniter.ConfigCompatibleWithStandardLibrary
)

func NewFromString(numString string) *Decimal {
	dec := &Decimal{}
	decimal, _, err := dec.SetString(numString)
	if err != nil {
		fmt.Println(err.Error())
	}

	return &Decimal{*decimal}
}

func NewFromFloat(numFloat float64) *Decimal {
	dec := &Decimal{}
	decimal, err := dec.SetFloat64(numFloat)
	if err != nil {
		fmt.Println(err.Error())
	}

	return &Decimal{*decimal}
}

func (d *Decimal) UnmarshalJSON(b []byte) error {
	_, _, err := d.SetString(string(b))

	return err
}

// MarshalJSON returns d as the JSON encoding of d which is json.Number.
func (d *Decimal) MarshalJSON() ([]byte, error) {
	return jsonSTD.Marshal(json.Number(d.Truncate(MAX_PRECISION, MIN_EXPONENT, DEFAULT_ROUNDING).Text('f')))
}

// IsEqual returns true if d and x are equal.
func (d *Decimal) IsEqual(x *Decimal) bool {
	return d.Cmp(&x.Decimal) == 0
}

// IsGreaterThan returns true if d is greater than x.
func (d *Decimal) IsGreaterThan(x *Decimal) bool {
	return d.Cmp(&x.Decimal) > 0
}

// IsLessThan returns true if d is less than x.
func (d *Decimal) IsLessThan(x *Decimal) bool {
	return d.Cmp(&x.Decimal) < 0
}

// IsGreaterThanOrEqual returns true if d is greater than or equal to x.
func (d *Decimal) IsGreaterThanOrEqual(x *Decimal) bool {
	return d.Cmp(&x.Decimal) >= 0
}

// IsLessThanOrEqual returns true if d is less than or equal to x.
func (d *Decimal) IsLessThanOrEqual(x *Decimal) bool {
	return d.Cmp(&x.Decimal) <= 0
}

// Truncate returns a new Decimal with the precision and exponent truncated.
func (d *Decimal) Truncate(precision uint32, minExponent int32, rounding apd.Rounder) *Decimal {
	obj := &apd.Decimal{}
	ctx := apd.BaseContext.WithPrecision(precision)
	ctx.Rounding = rounding

	// if the exponent of the number to be truncated is less than -8, assign minExponent to numExponent.
	numExponent := d.Exponent
	if numExponent < minExponent {
		numExponent = minExponent
	}
	ctx.Quantize(obj, &d.Decimal, numExponent)

	/*
		will remove trailing zeros and then to remove the decimal point if there are no more digits after it.
		For example, the number 10.500 is formatted as a string and then the decimal zeros are stripped, resulting in 10.5
	*/
	formatDecimal := d.Decimal.String()
	if d.Exponent < 0 {
		formatDecimal = strings.TrimRight(formatDecimal, "0")
		formatDecimal = strings.TrimRight(formatDecimal, ".")
	}

	return NewFromString(formatDecimal)
}

// Truncate returns a new Decimal with the precision and exponent truncated.
func (d *Decimal) TrimZerosRight() *Decimal {
	/*
		will remove trailing zeros and then to remove the decimal point if there are no more digits after it.
		For example, the number 10.500 is formatted as a string and then the decimal zeros are stripped, resulting in 10.5
	*/
	formatDecimal := d.Decimal.String()
	if d.Exponent < 0 {
		formatDecimal = strings.TrimRight(formatDecimal, "0")
		formatDecimal = strings.TrimRight(formatDecimal, ".")
	}

	return NewFromString(formatDecimal)
}

// Context library.
type context struct {
	*apd.Context
}

// Abs sets d to |x| (the absolute value of x).
func (c *context) Abs(d, x *Decimal) error {
	_, err := apdCtx.Abs(&d.Decimal, &x.Decimal)

	return err
}

// Add sets d to the sum x+y.
func (c *context) Add(d, x, y *Decimal) error {
	_, err := apdCtx.Add(&d.Decimal, &x.Decimal, &y.Decimal)

	return err
}

// Sub sets d to the difference x-y.
func (c *context) Sub(d, x, y *Decimal) error {
	_, err := apdCtx.Sub(&d.Decimal, &x.Decimal, &y.Decimal)

	return err
}

// Mul sets d to the product x*y.
func (c *context) Mul(d, x, y *Decimal) error {
	_, err := apdCtx.Mul(&d.Decimal, &x.Decimal, &y.Decimal)

	return err
}

// Quo sets d to the quotient x/y for y != 0. c.Precision must be > 0.
// If an exact division is required, use a context with high precision and verify it was exact by checking the Inexact flag on the return Condition.
func (c *context) Quo(d, x, y *Decimal) error {
	_, err := apdCtx.Quo(&d.Decimal, &x.Decimal, &y.Decimal)

	return err
}

// Neg sets d to -x.
func (c *context) Neg(d, x *Decimal) error {
	_, err := apdCtx.Neg(&d.Decimal, &x.Decimal)

	return err
}

// NewFromString creates a new decimal from s.
// The returned Decimal has its exponents restricted by the context and its value rounded if it contains more digits than the context's precision.
func (c *context) NewFromString(s string) (*Decimal, error) {
	d, _, err := apdCtx.NewFromString(s)
	if err != nil {
		return nil, err
	}

	return &Decimal{*d}, err
}
