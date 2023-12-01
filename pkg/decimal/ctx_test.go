package decimal

import (
	"fmt"
	"testing"
)

type TestDecimal func(s string) (*Decimal, error)

func mustNewFromString(d *Decimal, err error) *Decimal {
	if err != nil {
		fmt.Println("mustNewFromString err: ", err)

		return d
	}

	return d
}

func TestSub(t *testing.T) {
	tests := []struct {
		inputA *Decimal
		inputB *Decimal
		want   string
	}{
		{
			inputA: mustNewFromString(Ctx.NewFromString("2000000000.00000000")),
			inputB: mustNewFromString(Ctx.NewFromString("1000000000.00000000")),
			want:   "1000000000.00000000",
		},
		{
			inputA: mustNewFromString(Ctx.NewFromString("1000000000000.00000002")),
			inputB: mustNewFromString(Ctx.NewFromString("1000000000000.00000001")),
			want:   "0.00000001",
		},
		{
			inputA: mustNewFromString(Ctx.NewFromString("9999999999998.00000001")),
			inputB: mustNewFromString(Ctx.NewFromString("0000000000001.00000001")),
			want:   "9999999999997.00000000",
		},
		{
			inputA: mustNewFromString(Ctx.NewFromString("1.00000001")),
			inputB: mustNewFromString(Ctx.NewFromString("1.00000001")),
			want:   "0.00000000",
		},
		{
			inputA: mustNewFromString(Ctx.NewFromString("2.002")),
			inputB: mustNewFromString(Ctx.NewFromString("1.001")),
			want:   "1.001",
		},
		{
			inputA: mustNewFromString(Ctx.NewFromString("0.00000002")),
			inputB: mustNewFromString(Ctx.NewFromString("0.00000001")),
			want:   "0.00000001",
		},
	}

	for _, tt := range tests {
		t.Run("TestSub", func(t *testing.T) {
			got := mustNewFromString(Ctx.NewFromString("0"))
			err := Ctx.Sub(got, tt.inputA, tt.inputB)

			if err != nil || got.Text('f') != tt.want {
				t.Errorf("Decimal.Sub() = %v, want %v", got.Text('f'), tt.want)
			}
		})
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		inputA *Decimal
		inputB *Decimal
		want   string
	}{
		{
			inputA: mustNewFromString(Ctx.NewFromString("1000000000.00000000")),
			inputB: mustNewFromString(Ctx.NewFromString("1000000000.00000000")),
			want:   "2000000000.00000000",
		},
		{
			inputA: mustNewFromString(Ctx.NewFromString("1000000000000.00000001")),
			inputB: mustNewFromString(Ctx.NewFromString("1000000000000.00000001")),
			want:   "2000000000000.00000002",
		},
		{
			inputA: mustNewFromString(Ctx.NewFromString("9999999999998.00000001")),
			inputB: mustNewFromString(Ctx.NewFromString("0000000000001.00000001")),
			want:   "9999999999999.00000002",
		},
		{
			inputA: mustNewFromString(Ctx.NewFromString("1.00000001")),
			inputB: mustNewFromString(Ctx.NewFromString("1.00000001")),
			want:   "2.00000002",
		},
		{
			inputA: mustNewFromString(Ctx.NewFromString("1.001")),
			inputB: mustNewFromString(Ctx.NewFromString("1.001")),
			want:   "2.002",
		},
		{
			inputA: mustNewFromString(Ctx.NewFromString("0.00000001")),
			inputB: mustNewFromString(Ctx.NewFromString("0.00000001")),
			want:   "0.00000002",
		},
	}

	for _, tt := range tests {
		t.Run("TestAdd", func(t *testing.T) {
			got := mustNewFromString(Ctx.NewFromString("0"))
			err := Ctx.Add(got, tt.inputA, tt.inputB)

			if err != nil || got.Text('f') != tt.want {
				t.Errorf("Decimal.Add() = %v, want %v", got.Text('f'), tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input *Decimal
		want  string
	}{
		{
			input: mustNewFromString(Ctx.NewFromString("100.54")),
			want:  "100.54",
		},
		{
			input: mustNewFromString(Ctx.NewFromString("100.5")),
			want:  "100.5",
		},
		{
			input: mustNewFromString(Ctx.NewFromString("10999.39")),
			want:  "10999.39",
		},
		{
			input: mustNewFromString(Ctx.NewFromString("10999.39989")),
			want:  "10999.39989",
		},
		{
			input: mustNewFromString(Ctx.NewFromString("100000009.50000000")),
			want:  "100000009.5",
		},
		{
			input: mustNewFromString(Ctx.NewFromString("100.499")),
			want:  "100.499",
		},
		{
			input: mustNewFromString(Ctx.NewFromString("100.5000000099")),
			want:  "100.5000000099",
		},
		{
			input: mustNewFromString(Ctx.NewFromString("99999999.999999999")),
			want:  "99999999.999999999",
		},
	}

	for _, tt := range tests {
		t.Run("TestTruncate", func(t *testing.T) {
			truncated := tt.input.Truncate(MAX_PRECISION, MIN_EXPONENT, DEFAULT_ROUNDING)

			if truncated.String() != tt.want {
				t.Errorf("Decimal.Truncate() = %v, want %v", truncated.String(), tt.want)
			}
		})
	}
}
