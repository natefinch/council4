package counter

import "testing"

func TestAgeCounterKind(t *testing.T) {
	t.Parallel()
	if !Age.Valid() {
		t.Fatal("Age.Valid() = false; want true")
	}
	if Age.PlayerOnly() {
		t.Fatal("Age.PlayerOnly() = true; want false")
	}
	if got, want := Age.String(), "age"; got != want {
		t.Fatalf("Age.String() = %q; want %q", got, want)
	}
}

func TestLevelCounterKind(t *testing.T) {
	t.Parallel()
	if !Level.Valid() {
		t.Fatal("Level.Valid() = false; want true")
	}
	if Level.PlayerOnly() {
		t.Fatal("Level.PlayerOnly() = true; want false")
	}
	if got, want := Level.String(), "level"; got != want {
		t.Fatalf("Level.String() = %q; want %q", got, want)
	}
}

func TestAsymmetricPowerToughnessKinds(t *testing.T) {
	t.Parallel()
	cases := []struct {
		kind    Kind
		str     string
		power   int
		toughns int
	}{
		{PlusOnePlusZero, "+1/+0", 1, 0},
		{PlusTwoPlusTwo, "+2/+2", 2, 2},
		{MinusZeroMinusOne, "-0/-1", 0, -1},
		{PlusZeroPlusOne, "+0/+1", 0, 1},
		{MinusZeroMinusTwo, "-0/-2", 0, -2},
		{MinusTwoMinusTwo, "-2/-2", -2, -2},
		{PlusOnePlusTwo, "+1/+2", 1, 2},
		{PlusZeroPlusTwo, "+0/+2", 0, 2},
		{MinusTwoMinusOne, "-2/-1", -2, -1},
		{MinusOneMinusZero, "-1/-0", -1, 0},
	}
	for _, c := range cases {
		if !c.kind.Valid() {
			t.Errorf("%s.Valid() = false; want true", c.str)
		}
		if c.kind.PlayerOnly() {
			t.Errorf("%s.PlayerOnly() = true; want false", c.str)
		}
		if got := c.kind.String(); got != c.str {
			t.Errorf("String() = %q; want %q", got, c.str)
		}
		p, tn, ok := c.kind.powerToughness()
		if !ok {
			t.Errorf("%s.powerToughness() ok = false; want true", c.str)
		}
		if p != c.power || tn != c.toughns {
			t.Errorf("%s.powerToughness() = (%d, %d); want (%d, %d)", c.str, p, tn, c.power, c.toughns)
		}
	}
}

func TestPowerToughnessDelta(t *testing.T) {
	t.Parallel()
	s := NewSet()
	s.Add(PlusOnePlusZero, 2) // +2/+0
	s.Add(PlusZeroPlusOne, 1) // +0/+1
	s.Add(MinusZeroMinusOne, 3)
	s.Add(Charge, 5) // non-P/T counter, ignored
	power, toughness := s.PowerToughnessDelta()
	if power != 2 || toughness != -2 {
		t.Fatalf("PowerToughnessDelta() = (%d, %d); want (2, -2)", power, toughness)
	}
}

func TestPowerToughnessDeltaSymmetric(t *testing.T) {
	t.Parallel()
	s := NewSet()
	s.Add(PlusOnePlusOne, 3)
	s.Add(MinusOneMinusOne, 1)
	power, toughness := s.PowerToughnessDelta()
	if power != 2 || toughness != 2 {
		t.Fatalf("PowerToughnessDelta() = (%d, %d); want (2, 2)", power, toughness)
	}
}
