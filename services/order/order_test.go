package order

import "testing"

func TestTotal(t *testing.T) {
	for _, tc := range []struct {
		name string
		req  Request
		want int64
		bad  bool
	}{
		{"tax", Request{1, []Item{{1000, 2}}}, 2200, false},
		{"discount and tax", Request{1, []Item{{10000, 1}}}, 10450, false},
		{"negative quantity", Request{1, []Item{{1000, -1}}}, 0, true},
		{"bounded amount", Request{1, []Item{{1 << 62, 1000}}}, 0, true},
		{"missing items", Request{CustomerID: 1}, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Total(tc.req)
			if (err != nil) != tc.bad || got != tc.want {
				t.Fatalf("got (%d, %v), want (%d, bad=%v)", got, err, tc.want, tc.bad)
			}
		})
	}
}
