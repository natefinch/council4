package cardgen

import "testing"

func TestLowerDeclareAttackersCastRestriction(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Warrior's Stand",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{1}{W}",
		OracleText: "Cast this spell only during the declare attackers step and only if you've been attacked this step.\nCreatures you control get +2/+2 until end of turn.",
	})
	if len(face.StaticAbilities) != 1 ||
		!face.StaticAbilities[0].Body.CastOnlyAfterAttackedThisStep {
		t.Fatalf("static abilities = %#v, want defensive cast restriction", face.StaticAbilities)
	}
	if !face.SpellAbility.Exists {
		t.Fatal("spell body was not lowered")
	}
}
