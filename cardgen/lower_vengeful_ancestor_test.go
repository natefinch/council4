package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestLowerVengefulAncestor locks in the two mechanics Vengeful Ancestor needs
// beyond the pre-existing "enters or attacks" union trigger and goad-target
// action: (1) a goaded trigger subject — "Whenever a goaded creature attacks"
// lowers to an attacker-declared trigger whose SubjectSelection.MatchGoaded
// restricts it to creatures that are goaded right now; and (2) an
// event-permanent damage source — "it deals 1 damage to its controller" lowers
// to damage whose source is the triggering permanent (EventPermanentReference)
// and whose recipient is that permanent's controller.
func TestLowerVengefulAncestor(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Vengeful Ancestor",
		Layout:   "normal",
		ManaCost: "{2}{R}{R}",
		TypeLine: "Creature — Spirit Dragon",
		OracleText: "Flying\n" +
			"Whenever this creature enters or attacks, goad target creature. (Until your next turn, that creature attacks each combat if able and attacks a player other than you if able.)\n" +
			"Whenever a goaded creature attacks, it deals 1 damage to its controller.",
		Power:     new("3"),
		Toughness: new("4"),
	})

	if len(face.StaticAbilities) != 1 {
		t.Fatalf("got %d static abilities, want 1 (Flying)", len(face.StaticAbilities))
	}
	if len(face.TriggeredAbilities) != 2 {
		t.Fatalf("got %d triggered abilities, want 2", len(face.TriggeredAbilities))
	}

	// Ability 1: "Whenever this creature enters or attacks, goad target
	// creature." — the enters-or-attacks union trigger firing goad on a target.
	union := face.TriggeredAbilities[0].Trigger.Pattern
	if union.Event != game.EventPermanentEnteredBattlefield {
		t.Fatalf("union trigger event = %v, want EventPermanentEnteredBattlefield", union.Event)
	}
	if union.UnionEvent != game.EventAttackerDeclared {
		t.Fatalf("union trigger UnionEvent = %v, want EventAttackerDeclared", union.UnionEvent)
	}
	if union.Source != game.TriggerSourceSelf {
		t.Fatalf("union trigger source = %v, want TriggerSourceSelf", union.Source)
	}
	unionModes := face.TriggeredAbilities[0].Content.Modes
	if len(unionModes) != 1 || len(unionModes[0].Sequence) != 1 {
		t.Fatalf("union trigger content = %#v, want one mode with one instruction", unionModes)
	}
	goad, ok := unionModes[0].Sequence[0].Primitive.(game.Goad)
	if !ok {
		t.Fatalf("union trigger primitive = %#v, want Goad", unionModes[0].Sequence[0].Primitive)
	}
	if goad.Object != game.TargetPermanentReference(0) {
		t.Fatalf("goad object = %#v, want target permanent 0", goad.Object)
	}

	// Ability 2: "Whenever a goaded creature attacks, it deals 1 damage to its
	// controller." — a goaded-subject attack trigger dealing damage sourced from
	// the attacking creature to that creature's controller.
	goadedPattern := face.TriggeredAbilities[1].Trigger.Pattern
	if goadedPattern.Event != game.EventAttackerDeclared {
		t.Fatalf("goaded trigger event = %v, want EventAttackerDeclared", goadedPattern.Event)
	}
	if !goadedPattern.SubjectSelection.MatchGoaded {
		t.Fatal("goaded trigger SubjectSelection.MatchGoaded = false, want true")
	}
	if !reflect.DeepEqual(goadedPattern.SubjectSelection.RequiredTypes, []types.Card{types.Creature}) {
		t.Fatalf("goaded trigger SubjectSelection.RequiredTypes = %#v, want [Creature]", goadedPattern.SubjectSelection.RequiredTypes)
	}

	damageModes := face.TriggeredAbilities[1].Content.Modes
	if len(damageModes) != 1 || len(damageModes[0].Sequence) != 1 {
		t.Fatalf("goaded trigger content = %#v, want one mode with one instruction", damageModes)
	}
	damage, ok := damageModes[0].Sequence[0].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("goaded trigger primitive = %#v, want Damage", damageModes[0].Sequence[0].Primitive)
	}
	if damage.Amount.Value() != 1 {
		t.Fatalf("damage amount = %d, want 1", damage.Amount.Value())
	}
	wantRecipient := game.PlayerDamageRecipient(game.ObjectControllerReference(game.EventPermanentReference()))
	if !reflect.DeepEqual(damage.Recipient, wantRecipient) {
		t.Fatalf("damage recipient = %#v, want its-controller of the event permanent", damage.Recipient)
	}
	wantSource := opt.Val(game.EventPermanentReference())
	if !reflect.DeepEqual(damage.DamageSource, wantSource) {
		t.Fatalf("damage source = %#v, want opt.Val(EventPermanentReference())", damage.DamageSource)
	}
}
