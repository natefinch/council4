package main

import (
	"fmt"
	"go/constant"
	"slices"
)

// bitmaskTypes are the reachable enum types whose values are combinations of
// independent flags rather than a single constant. They render as an OR of the
// set flags ("game.DamageRecipientPlayer | game.DamageRecipientPermanent"), so a
// value-to-name table cannot describe them.
//
// The list is explicit rather than inferred because flag shape is not detectable
// from values alone: any enum with constants 0, 1, and 2 looks like a bitmask.
// Generation asserts that each listed type really is flag-shaped, so a type that
// stops being a bitmask fails the build here instead of silently rendering
// nonsense.
var bitmaskTypes = []string{
	"game.AttackRecipientKind",
	"game.DamageRecipientKind",
	"game.TargetAllow",
}

// isBitmask reports whether e renders as an OR of flags.
func isBitmask(e enumType) bool { return slices.Contains(bitmaskTypes, e.Ref()) }

// flags returns the type's non-zero constants in ascending value order, which is
// the order they are ORed together in generated source.
//
// It fails if any non-zero constant is not a distinct power of two, because such
// a constant is either a named combination (which would render twice) or proof
// that the type is not flag-shaped at all.
func flags(e enumType) ([]enumValue, error) {
	var out []enumValue
	seen := map[int64]bool{}
	for _, v := range e.Values {
		n, ok := constant.Int64Val(constant.ToInt(v.Value))
		if !ok {
			return nil, fmt.Errorf("%s: constant %s is not an integer", e.Ref(), v.Name)
		}
		if n == 0 {
			continue
		}
		if n&(n-1) != 0 {
			return nil, fmt.Errorf("%s: constant %s = %d is not a power of two, so %s is not a bitmask", e.Ref(), v.Name, n, e.Ref())
		}
		if seen[n] {
			return nil, fmt.Errorf("%s: constant %s duplicates flag %d", e.Ref(), v.Name, n)
		}
		seen[n] = true
		out = append(out, v)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%s: no non-zero constants to treat as flags", e.Ref())
	}
	slices.SortFunc(out, func(a, b enumValue) int {
		an, _ := constant.Int64Val(constant.ToInt(a.Value))
		bn, _ := constant.Int64Val(constant.ToInt(b.Value))
		return int(an - bn)
	})
	return out, nil
}

// FlagsName is the generated slice variable holding this type's flags in
// ascending value order.
func (e enumType) FlagsName() string {
	return lowerFirst(exportedPrefix(e.Pkg.Name())) + e.Name + "Flags"
}

// checkBitmaskTypes fails if a type named in bitmaskTypes is not reachable, so
// the list cannot silently rot as types are renamed or removed.
func checkBitmaskTypes(enums []enumType) error {
	for _, ref := range bitmaskTypes {
		if !slices.ContainsFunc(enums, func(e enumType) bool { return e.Ref() == ref }) {
			return fmt.Errorf("bitmaskTypes names %s, which is not a reachable enum type", ref)
		}
	}
	return nil
}
