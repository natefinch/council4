package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
)

// TestLowerMaxSpeedStatic verifies that a "Max speed — <static>" ability
// (CR 702.179) lowers to a static ability gated on the controller having maximum
// speed (game.Condition.ControllerHasMaxSpeed), stripping the ability word so the
// ordinary static path accepts the body.
func TestLowerMaxSpeedStatic(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Swiftwing Assailant",
		Layout:   "normal",
		TypeLine: "Creature — Bird Warrior",
		ManaCost: "{3}{W}",
		OracleText: "Flying\nStart your engines!\n" +
			"Max speed — This creature gets +0/+1 and has vigilance.",
		Power:     new("2"),
		Toughness: new("2"),
	})
	gated := false
	for _, static := range face.StaticAbilities {
		if static.Body.Condition.Exists && static.Body.Condition.Val.ControllerHasMaxSpeed {
			gated = true
		}
	}
	if !gated {
		t.Fatalf("no static ability gated on ControllerHasMaxSpeed: %+v", face.StaticAbilities)
	}
}

// TestLowerMaxSpeedBareKeywordFailsClosed verifies that a "Max speed — <bare
// keyword>" static (whose lowered body would carry KeywordAbilities the runtime
// applies without evaluating the ability's Condition) fails closed rather than
// generating a card that grants the keyword regardless of speed.
func TestLowerMaxSpeedBareKeywordFailsClosed(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Speedster",
		Layout:     "normal",
		TypeLine:   "Creature — Goblin",
		ManaCost:   "{R}",
		OracleText: "Start your engines!\nMax speed — Trample",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("bare-keyword Max speed static unexpectedly compiled; it must fail closed")
	}
}

// TestLowerMaxSpeedGraveyardActivated verifies that a "Max speed — {cost}: ..."
// activated ability that functions from the graveyard lowers with the
// ControllerHasMaxSpeed activation condition (the condition is evaluated against
// the controlling player, so it works without a battlefield source permanent).
func TestLowerMaxSpeedGraveyardActivated(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Goblin Surveyor",
		Layout:   "normal",
		TypeLine: "Creature — Goblin Scout",
		ManaCost: "{R}",
		OracleText: "Start your engines!\n" +
			"Max speed — {3}, Exile this card from your graveyard: Draw a card.",
		Power:     new("1"),
		Toughness: new("1"),
	})
	found := false
	for _, ability := range face.ActivatedAbilities {
		if ability.ZoneOfFunction == zone.Graveyard &&
			ability.ActivationCondition.Exists &&
			ability.ActivationCondition.Val.ControllerHasMaxSpeed {
			found = true
		}
	}
	if !found {
		t.Fatalf("no graveyard activated ability gated on ControllerHasMaxSpeed: %+v", face.ActivatedAbilities)
	}
}
