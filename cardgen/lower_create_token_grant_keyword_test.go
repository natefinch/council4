package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerCreateTokenThatTokenGainsHaste verifies the ordered pair "Create a
// ... token. That token gains haste until end of turn." lowers to a token
// creation that publishes its result under a link key, followed by an
// until-end-of-turn keyword grant whose object resolves to that linked token.
func TestLowerCreateTokenThatTokenGainsHaste(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Hasty Goblin Maker",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "When this enchantment enters, create a 1/1 red Goblin creature token. That token gains haste until end of turn.",
	})
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want create then keyword grant", mode.Sequence)
	}
	create, ok := mode.Sequence[0].Primitive.(game.CreateToken)
	if !ok || create.PublishLinked == "" {
		t.Fatalf("create = %#v, want a token creation publishing a link", mode.Sequence[0].Primitive)
	}
	apply, ok := mode.Sequence[1].Primitive.(game.ApplyContinuous)
	if !ok ||
		!apply.Object.Exists ||
		apply.Object.Val.Kind() != game.ObjectReferenceLinkedObject ||
		apply.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("apply = %#v, want linked-object grant until end of turn", mode.Sequence[1].Primitive)
	}
	if apply.Object.Val.LinkID() != string(create.PublishLinked) {
		t.Fatalf("grant link %q != create link %q", apply.Object.Val.LinkID(), create.PublishLinked)
	}
	if len(apply.ContinuousEffects) != 1 ||
		len(apply.ContinuousEffects[0].AddKeywords) != 1 ||
		apply.ContinuousEffects[0].AddKeywords[0] != game.Haste {
		t.Fatalf("continuous effect = %#v, want haste grant", apply.ContinuousEffects)
	}
}

// TestLowerLoyalApprenticeTokenGainsHaste verifies the actual Loyal Apprentice
// wording — a Lieutenant combat trigger gated on controlling your commander —
// lowers its "That token gains haste until end of turn" sub-effect to a linked
// keyword grant on the just-created Thopter token.
func TestLowerLoyalApprenticeTokenGainsHaste(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Loyal Apprentice",
		Layout:     "normal",
		TypeLine:   "Creature — Human Artificer",
		ManaCost:   "{2}",
		OracleText: "Haste\nLieutenant — At the beginning of combat on your turn, if you control your commander, create a 1/1 colorless Thopter artifact creature token with flying. That token gains haste until end of turn.",
	})
	var combatTrigger *game.TriggeredAbility
	for i := range face.TriggeredAbilities {
		if len(face.TriggeredAbilities[i].Content.Modes) == 1 &&
			len(face.TriggeredAbilities[i].Content.Modes[0].Sequence) == 2 {
			combatTrigger = &face.TriggeredAbilities[i]
		}
	}
	if combatTrigger == nil {
		t.Fatalf("no create-then-grant combat trigger found in %#v", face.TriggeredAbilities)
	}
	seq := combatTrigger.Content.Modes[0].Sequence
	create, ok := seq[0].Primitive.(game.CreateToken)
	if !ok || create.PublishLinked == "" {
		t.Fatalf("create = %#v, want a Thopter token publishing a link", seq[0].Primitive)
	}
	apply, ok := seq[1].Primitive.(game.ApplyContinuous)
	if !ok ||
		!apply.Object.Exists ||
		apply.Object.Val.Kind() != game.ObjectReferenceLinkedObject ||
		apply.Object.Val.LinkID() != string(create.PublishLinked) ||
		len(apply.ContinuousEffects) != 1 ||
		len(apply.ContinuousEffects[0].AddKeywords) != 1 ||
		apply.ContinuousEffects[0].AddKeywords[0] != game.Haste {
		t.Fatalf("apply = %#v, want haste granted to the linked token", seq[1].Primitive)
	}
}
