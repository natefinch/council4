package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// optimusScryfallCard is the full Transformers double-faced card as Scryfall
// data so the test exercises the real two-face lowering path, including the
// back face's self-reference to "Optimus Prime" in its delayed convert.
func optimusScryfallCard() *ScryfallCard {
	return &ScryfallCard{
		Name:     "Optimus Prime, Hero // Optimus Prime, Autobot Leader",
		Layout:   "transform",
		TypeLine: "Legendary Artifact Creature — Robot // Legendary Artifact — Vehicle",
		CardFaces: []ScryfallCardFace{
			{
				Name:      "Optimus Prime, Hero",
				ManaCost:  "{2}{U}{R}{W}",
				TypeLine:  "Legendary Artifact Creature — Robot",
				Power:     new("4"),
				Toughness: new("8"),
				OracleText: "More Than Meets the Eye {2}{U}{R}{W} (You may cast this card converted for {2}{U}{R}{W}.)\n" +
					"At the beginning of each end step, bolster 1. (Choose a creature with the least toughness among creatures you control and put a +1/+1 counter on it.)\n" +
					"When Optimus Prime dies, return it to the battlefield converted under its owner's control.",
			},
			{
				Name:      "Optimus Prime, Autobot Leader",
				TypeLine:  "Legendary Artifact — Vehicle",
				Power:     new("6"),
				Toughness: new("8"),
				OracleText: "Living metal (During your turn, this Vehicle is also a creature.)\n" +
					"Trample\n" +
					"Whenever you attack, bolster 2. The chosen creature gains trample until end of turn. When that creature deals combat damage to a player this turn, convert Optimus Prime.",
			},
		},
	}
}

// TestGenerateOptimusPrimeHeroFront proves the front face lowers its
// "At the beginning of each end step, bolster 1." trigger into the reusable
// bolster keyword action (mechanic #1) as a standalone Bolster{Fixed(1)} with no
// linked publication, alongside the already-supported dies-return-converted
// trigger.
func TestGenerateOptimusPrimeHeroFront(t *testing.T) {
	t.Parallel()
	faces, diagnostics := lowerExecutableFaces(optimusScryfallCard())
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if len(faces) < 2 {
		t.Fatalf("lowered faces = %d, want 2", len(faces))
	}
	front := faces[0]
	if len(front.TriggeredAbilities) != 2 {
		t.Fatalf("front triggered abilities = %d, want 2", len(front.TriggeredAbilities))
	}

	bolsterAbility := front.TriggeredAbilities[0]
	if bolsterAbility.Trigger.Type != game.TriggerAt ||
		bolsterAbility.Trigger.Pattern.Event != game.EventBeginningOfStep ||
		bolsterAbility.Trigger.Pattern.Step != game.StepEnd {
		t.Fatalf("bolster trigger = %#v, want at each end step", bolsterAbility.Trigger)
	}
	sequence := bolsterAbility.Content.Modes[0].Sequence
	if len(sequence) != 1 {
		t.Fatalf("bolster sequence len = %d, want 1", len(sequence))
	}
	bolster, ok := sequence[0].Primitive.(game.Bolster)
	if !ok {
		t.Fatalf("bolster primitive = %T, want game.Bolster", sequence[0].Primitive)
	}
	if bolster.Amount != game.Fixed(1) {
		t.Fatalf("bolster amount = %#v, want Fixed(1)", bolster.Amount)
	}
	if bolster.PublishLinked != "" {
		t.Fatalf("front bolster PublishLinked = %q, want empty (no rider references the chosen creature)", bolster.PublishLinked)
	}

	dies := front.TriggeredAbilities[1]
	if dies.Trigger.Pattern.Event != game.EventPermanentDied || dies.Trigger.Pattern.Source != game.TriggerSourceSelf {
		t.Fatalf("dies trigger = %#v, want when this dies", dies.Trigger)
	}
}

// TestGenerateOptimusPrimeAutobotLeaderBack proves the back face lowers its
// "Whenever you attack, bolster 2. The chosen creature gains trample until end
// of turn. When that creature deals combat damage to a player this turn, convert
// Optimus Prime." trigger into the three reusable mechanics: bolster 2 that
// publishes its chosen creature as a linked object (mechanic #1), the linked
// "the chosen creature gains trample until end of turn" grant (mechanic #2), and
// the delayed combat-damage convert bound to that same linked creature
// (mechanic #3).
func TestGenerateOptimusPrimeAutobotLeaderBack(t *testing.T) {
	t.Parallel()
	faces, diagnostics := lowerExecutableFaces(optimusScryfallCard())
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if len(faces) < 2 {
		t.Fatalf("lowered faces = %d, want 2", len(faces))
	}
	back := faces[1]
	if len(back.TriggeredAbilities) != 1 {
		t.Fatalf("back triggered abilities = %d, want 1", len(back.TriggeredAbilities))
	}
	ability := back.TriggeredAbilities[0]
	if ability.Trigger.Type != game.TriggerWhenever ||
		ability.Trigger.Pattern.Event != game.EventAttackerDeclared ||
		ability.Trigger.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("back trigger = %#v, want whenever you attack", ability.Trigger)
	}
	sequence := ability.Content.Modes[0].Sequence
	if len(sequence) != 3 {
		t.Fatalf("back sequence len = %d, want 3", len(sequence))
	}

	// Mechanic #1: bolster 2 that publishes its chosen creature as a linked object.
	bolster, ok := sequence[0].Primitive.(game.Bolster)
	if !ok {
		t.Fatalf("sequence[0] = %T, want game.Bolster", sequence[0].Primitive)
	}
	if bolster.Amount != game.Fixed(2) {
		t.Fatalf("bolster amount = %#v, want Fixed(2)", bolster.Amount)
	}
	linkKey := bolster.PublishLinked
	if linkKey == "" {
		t.Fatal("back bolster does not publish its chosen creature as a linked object")
	}

	// Mechanic #2: the chosen creature gains trample until end of turn, bound to
	// the bolstered creature through the published linked reference.
	grant, ok := sequence[1].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("sequence[1] = %T, want game.ApplyContinuous", sequence[1].Primitive)
	}
	if !grant.Object.Exists || grant.Object.Val != game.LinkedObjectReference(string(linkKey)) {
		t.Fatalf("grant object = %#v, want LinkedObjectReference(%q)", grant.Object, linkKey)
	}
	if grant.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("grant duration = %v, want DurationUntilEndOfTurn", grant.Duration)
	}
	if len(grant.ContinuousEffects) != 1 ||
		len(grant.ContinuousEffects[0].AddKeywords) != 1 ||
		grant.ContinuousEffects[0].AddKeywords[0] != game.Trample {
		t.Fatalf("grant continuous effects = %#v, want a single Trample grant", grant.ContinuousEffects)
	}

	// Mechanic #3: the delayed combat-damage convert bound to the same linked
	// creature, transforming the source this turn.
	delayed, ok := sequence[2].Primitive.(game.CreateDelayedTrigger)
	if !ok {
		t.Fatalf("sequence[2] = %T, want game.CreateDelayedTrigger", sequence[2].Primitive)
	}
	if delayed.Trigger.Window != game.DelayedWindowThisTurn {
		t.Fatalf("delayed window = %v, want DelayedWindowThisTurn", delayed.Trigger.Window)
	}
	if !delayed.Trigger.DamageSourceObject.Exists ||
		delayed.Trigger.DamageSourceObject.Val != game.LinkedObjectReference(string(linkKey)) {
		t.Fatalf("delayed damage source = %#v, want LinkedObjectReference(%q)", delayed.Trigger.DamageSourceObject, linkKey)
	}
	pattern := delayed.Trigger.EventPattern
	if !pattern.Exists ||
		pattern.Val.Event != game.EventDamageDealt ||
		!pattern.Val.RequireCombatDamage ||
		pattern.Val.DamageRecipient != game.DamageRecipientPlayer ||
		!pattern.Val.DamageSourceCaptured {
		t.Fatalf("delayed event pattern = %#v, want captured-source combat damage to a player", pattern)
	}
	inner := delayed.Trigger.Content.Modes[0].Sequence
	if len(inner) != 1 {
		t.Fatalf("delayed inner sequence len = %d, want 1", len(inner))
	}
	transform, ok := inner[0].Primitive.(game.Transform)
	if !ok {
		t.Fatalf("delayed inner primitive = %T, want game.Transform", inner[0].Primitive)
	}
	if transform.Object != game.SourcePermanentReference() {
		t.Fatalf("transform object = %#v, want SourcePermanentReference()", transform.Object)
	}
}
