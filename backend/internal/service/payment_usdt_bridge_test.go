package service

import "testing"

func TestBaseAmountForTargetPayPreservesExactCNYTotal(t *testing.T) {
	for _, test := range []struct {
		target float64
		fee    float64
		want   float64
	}{
		{target: 72.34, fee: 2, want: 70.92},
		{target: 10, fee: 0, want: 10},
		{target: 10.01, fee: 3, want: 9.72},
	} {
		got := baseAmountForTargetPay(test.target, test.fee)
		if got != test.want {
			t.Fatalf("baseAmountForTargetPay(%v, %v) = %v, want %v", test.target, test.fee, got, test.want)
		}
	}
}
