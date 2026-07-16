package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerValleyFloodcaller proves the whole card composes end-to-end from
// existing mechanisms: the Flash keyword, a noncreature "cast as though they had
// flash" permission, and a noncreature cast trigger whose ordered body pumps the
// four controlled subtype groups and untaps that same group.
//
//	Flash
//	You may cast noncreature spells as though they had flash.
//	Whenever you cast a noncreature spell, Birds, Frogs, Otters, and Rats you
//	control get +1/+1 until end of turn. Untap them.
func TestLowerValleyFloodcaller(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:      "Valley Floodcaller",
		Layout:    "normal",
		ManaCost:  "{1}{G}{U}",
		TypeLine:  "Legendary Creature — Otter Wizard",
		Power:     new("2"),
		Toughness: new("2"),
		OracleText: "Flash\n" +
			"You may cast noncreature spells as though they had flash.\n" +
			"Whenever you cast a noncreature spell, Birds, Frogs, Otters, and Rats you control get +1/+1 until end of turn. Untap them.",
	})

	// Two static abilities: the Flash keyword body and the flash-grant permission.
	if len(face.StaticAbilities) != 2 {
		t.Fatalf("static abilities = %d, want 2 (flash keyword + flash grant)", len(face.StaticAbilities))
	}
	var sawFlashKeyword bool
	var permission *game.RuleEffect
	for i := range face.StaticAbilities {
		body := face.StaticAbilities[i].Body
		if reflect.DeepEqual(body, game.FlashStaticBody) {
			sawFlashKeyword = true
			continue
		}
		for j := range body.RuleEffects {
			if body.RuleEffects[j].Kind == game.RuleEffectCastSpellsAsThoughFlash {
				permission = &body.RuleEffects[j]
			}
		}
	}
	if !sawFlashKeyword {
		t.Fatalf("no Flash keyword static ability among %#v", face.StaticAbilities)
	}
	if permission == nil {
		t.Fatalf("no cast-as-though-flash permission among %#v", face.StaticAbilities)
	}
	if permission.AffectedPlayer != game.PlayerYou {
		t.Fatalf("permission affected player = %v, want you", permission.AffectedPlayer)
	}
	if len(permission.SpellTypes) != 0 || len(permission.SpellSubtypes) != 0 {
		t.Fatalf("permission carries a positive filter %#v, want only an exclusion", permission)
	}
	if len(permission.ExcludedSpellTypes) != 1 || permission.ExcludedSpellTypes[0] != types.Creature {
		t.Fatalf("permission excluded types = %#v, want [Creature]", permission.ExcludedSpellTypes)
	}

	// One triggered ability: whenever you cast a noncreature spell.
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	if trigger.Trigger.Type != game.TriggerWhenever ||
		trigger.Trigger.Pattern.Event != game.EventSpellCast ||
		trigger.Trigger.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("trigger = %#v, want whenever you cast a spell", trigger.Trigger)
	}
	sel := trigger.Trigger.Pattern.CardSelection
	if len(sel.ExcludedTypes) != 1 || sel.ExcludedTypes[0] != types.Creature {
		t.Fatalf("trigger card selection excluded types = %#v, want [Creature]", sel.ExcludedTypes)
	}

	// The trigger body is a pump-then-untap of the same four-subtype group.
	sequence := trigger.Content.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("trigger sequence = %#v, want pump then untap", sequence)
	}
	apply, ok := sequence[0].Primitive.(game.ApplyContinuous)
	if !ok || apply.Object.Exists || apply.Duration != game.DurationUntilEndOfTurn || len(apply.ContinuousEffects) != 1 {
		t.Fatalf("sequence[0] = %#v, want unanchored group pump until end of turn", sequence[0].Primitive)
	}
	pump := apply.ContinuousEffects[0]
	if pump.Layer != game.LayerPowerToughnessModify || pump.PowerDelta != 1 || pump.ToughnessDelta != 1 {
		t.Fatalf("pump = %+v, want +1/+1", pump)
	}
	untap, ok := sequence[1].Primitive.(game.Untap)
	if !ok || untap.ChooseUpTo || untap.ChooseOne || untap.Object != (game.ObjectReference{}) {
		t.Fatalf("sequence[1] = %#v, want a plain mass group untap", sequence[1].Primitive)
	}
	wantSubs := []types.Sub{types.Bird, types.Frog, types.Otter, types.Rat}
	pumpSel := pump.Group.Selection()
	untapSel := untap.Group.Selection()
	if !subtypesAnyEqual(pumpSel.SubtypesAny, wantSubs) || pumpSel.Controller != game.ControllerYou {
		t.Fatalf("pump group = %+v, want controlled Bird/Frog/Otter/Rat group", pumpSel)
	}
	if !subtypesAnyEqual(untapSel.SubtypesAny, wantSubs) ||
		untapSel.Controller != game.ControllerYou ||
		untap.Group.Domain() != pump.Group.Domain() {
		t.Fatalf("untap group = %+v, want the same group as the pump", untapSel)
	}
}

func subtypesAnyEqual(got, want []types.Sub) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
