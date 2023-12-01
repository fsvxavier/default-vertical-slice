package decimal

import (
	"testing"
)

func TestDecimal_IsEqual(t *testing.T) {
	tests := []struct {
		x    string
		y    string
		want bool
	}{
		{
			x:    "1.0",
			y:    "1.0",
			want: true,
		},
		{
			x:    "1.0",
			y:    "2.0",
			want: false,
		},
		{
			x:    "1.0",
			y:    "0.000000000000000001",
			want: false,
		},
		{
			x:    "0.000000000000000002",
			y:    "0.000000000000000002",
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run("TestDecimal_IsEqual", func(t *testing.T) {
			x := NewFromString(tt.x)
			y := NewFromString(tt.y)

			if got := x.IsEqual(y); got != tt.want {
				t.Errorf("Decimal.IsEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecimal_IsGreaterThan(t *testing.T) {
	tests := []struct {
		x    string
		y    string
		want bool
	}{
		{
			x:    "1.0",
			y:    "1.0",
			want: false,
		},
		{
			x:    "1.0",
			y:    "2.0",
			want: false,
		},
		{
			x:    "2.0",
			y:    "1.0",
			want: true,
		},
		{
			x:    "1.0",
			y:    "0.000000000000000001",
			want: true,
		},
		{
			x:    "0.000000000000000002",
			y:    "0.000000000000000002",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run("IsGreaterThan", func(t *testing.T) {
			x := NewFromString(tt.x)
			y := NewFromString(tt.y)

			if got := x.IsGreaterThan(y); got != tt.want {
				t.Errorf("Decimal.IsGreaterThan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecimal_IsLessThan(t *testing.T) {
	tests := []struct {
		x    string
		y    string
		want bool
	}{
		{
			x:    "1.0",
			y:    "1.0",
			want: false,
		},
		{
			x:    "1.0",
			y:    "2.0",
			want: true,
		},
		{
			x:    "2.0",
			y:    "1.0",
			want: false,
		},
		{
			x:    "1.0",
			y:    "0.000000000000000001",
			want: false,
		},
		{
			x:    "0.000000000000000001",
			y:    "1.0",
			want: true,
		},
		{
			x:    "0.000000000000000002",
			y:    "0.000000000000000002",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run("IsLessThan", func(t *testing.T) {
			x := NewFromString(tt.x)
			y := NewFromString(tt.y)

			if got := x.IsLessThan(y); got != tt.want {
				t.Errorf("Decimal.IsLessThan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecimal_IsGreaterThanOrEqual(t *testing.T) {
	tests := []struct {
		x    string
		y    string
		want bool
	}{
		{
			x:    "1.0",
			y:    "1.0",
			want: true,
		},
		{
			x:    "1.0",
			y:    "2.0",
			want: false,
		},
		{
			x:    "2.0",
			y:    "1.0",
			want: true,
		},
		{
			x:    "1.0",
			y:    "0.000000000000000001",
			want: true,
		},
		{
			x:    "0.000000000000000001",
			y:    "1.0",
			want: false,
		},
		{
			x:    "0.000000000000000002",
			y:    "0.000000000000000002",
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run("IsGreaterThanOrEqual", func(t *testing.T) {
			x := NewFromString(tt.x)
			y := NewFromString(tt.y)

			if got := x.IsGreaterThanOrEqual(y); got != tt.want {
				t.Errorf("Decimal.IsGreaterThanOrEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecimal_IsLessThanOrEqual(t *testing.T) {
	tests := []struct {
		x    string
		y    string
		want bool
	}{
		{
			x:    "1.0",
			y:    "1.0",
			want: true,
		},
		{
			x:    "1.0",
			y:    "2.0",
			want: true,
		},
		{
			x:    "2.0",
			y:    "1.0",
			want: false,
		},
		{
			x:    "1.0",
			y:    "0.000000000000000001",
			want: false,
		},
		{
			x:    "0.000000000000000001",
			y:    "1.0",
			want: true,
		},
		{
			x:    "0.000000000000000002",
			y:    "0.000000000000000002",
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run("IsLessThanOrEqual", func(t *testing.T) {
			x := NewFromString(tt.x)
			y := NewFromString(tt.y)

			if got := x.IsLessThanOrEqual(y); got != tt.want {
				t.Errorf("Decimal.IsLessThanOrEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}
