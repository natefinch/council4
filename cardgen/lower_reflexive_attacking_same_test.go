package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// curseFaceForRider lowers a single-face Aura Curse whose oracle text is the
// enchant line followed by the given ability text, returning the lowered face for
// the reflexive attacking-opponent rider tests.
func curseFaceForRider(t *testing.T, name, ability string) loweredFaceAbilities {
	t.Helper()
	return lowerSingleFace(t, &ScryfallCard{
		Name:       name,
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura Curse",
		OracleText: "Enchant player\n" + ability,
	})
}

// reflexiveRiderSequence returns the instruction sequence of the lowered curse's
// single enchanted-player-attacked triggered ability, asserting the trigger shape
// and that the ability folded to exactly two instructions: the controller effect
// and the reflexive opponents-attacking group effect.
func reflexiveRiderSequence(t *testing.T, face loweredFaceAbilities) []game.Instruction {
	t.Helper()
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	pattern := face.TriggeredAbilities[0].Trigger.Pattern
	if pattern.Event != game.EventAttackerDeclared || !pattern.AttackedPlayerIsSourceEnchantedPlayer || !pattern.OneOrMore {
		t.Fatalf("trigger pattern = %#v, want attacker-declared enchanted-player once-per-combat", pattern)
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %d instructions, want 2 (controller + reflexive group)", len(mode.Sequence))
	}
	return mode.Sequence
}

// TestLowerReflexiveAttackingGainLifeRider proves the anaphoric "Each opponent
// attacking that player does the same." rider on a lone controller gain-life
// effect (Curse of Vitality) lowers to the controller gain plus a second gain
// widened to the opponents-attacking-trigger-player group, mirroring the exact
// generated card.
func TestLowerReflexiveAttackingGainLifeRider(t *testing.T) {
	t.Parallel()
	seq := reflexiveRiderSequence(t, curseFaceForRider(t, "Rider Vitality",
		"Whenever enchanted player is attacked, you gain 2 life. Each opponent attacking that player does the same."))

	wantController := game.GainLife{Amount: game.Fixed(2), Player: game.ControllerReference()}
	if !reflect.DeepEqual(seq[0].Primitive, wantController) {
		t.Fatalf("controller instruction = %#v, want %#v", seq[0].Primitive, wantController)
	}
	wantGroup := game.GainLife{Amount: game.Fixed(2), PlayerGroup: game.OpponentsAttackingTriggerPlayerReference()}
	if !reflect.DeepEqual(seq[1].Primitive, wantGroup) {
		t.Fatalf("group instruction = %#v, want %#v", seq[1].Primitive, wantGroup)
	}
	if seq[1].ForEachPlayerGroup.Exists {
		t.Fatalf("group gain-life carries ForEachPlayerGroup = %#v, want none (group is on the primitive)", seq[1].ForEachPlayerGroup)
	}
}

// TestLowerReflexiveAttackingDrawRider proves the anaphoric rider on a lone
// controller draw effect (Curse of Verbosity) lowers to the controller draw plus a
// second draw widened to the opponents-attacking-trigger-player group.
func TestLowerReflexiveAttackingDrawRider(t *testing.T) {
	t.Parallel()
	seq := reflexiveRiderSequence(t, curseFaceForRider(t, "Rider Verbosity",
		"Whenever enchanted player is attacked, you draw a card. Each opponent attacking that player does the same."))

	wantController := game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}
	if !reflect.DeepEqual(seq[0].Primitive, wantController) {
		t.Fatalf("controller instruction = %#v, want %#v", seq[0].Primitive, wantController)
	}
	wantGroup := game.Draw{Amount: game.Fixed(1), PlayerGroup: game.OpponentsAttackingTriggerPlayerReference()}
	if !reflect.DeepEqual(seq[1].Primitive, wantGroup) {
		t.Fatalf("group instruction = %#v, want %#v", seq[1].Primitive, wantGroup)
	}
}

// TestLowerReflexiveAttackingUntapRider proves the explicit "Each opponent
// attacking that player untaps all nonland permanents they control." rider (Curse
// of Bounty) lowers to the controller's own nonland untap plus a per-attacker
// untap iterated over the opponents-attacking group. The controller untap targets
// the controller's nonland permanents, while the group untap targets each member's
// nonland permanents through the group-offer member reference.
func TestLowerReflexiveAttackingUntapRider(t *testing.T) {
	t.Parallel()
	seq := reflexiveRiderSequence(t, curseFaceForRider(t, "Rider Bounty",
		"Whenever enchanted player is attacked, untap all nonland permanents you control. Each opponent attacking that player untaps all nonland permanents they control."))

	wantController := game.Untap{Group: game.BattlefieldGroup(game.Selection{
		ExcludedTypes: []types.Card{types.Land},
		Controller:    game.ControllerYou,
	})}
	if !reflect.DeepEqual(seq[0].Primitive, wantController) {
		t.Fatalf("controller instruction = %#v, want %#v", seq[0].Primitive, wantController)
	}
	if seq[0].ForEachPlayerGroup.Exists {
		t.Fatalf("controller untap carries ForEachPlayerGroup = %#v, want none", seq[0].ForEachPlayerGroup)
	}

	wantGroup := game.Untap{Group: game.PlayerControlledGroup(game.GroupOfferMemberReference(), game.Selection{
		ExcludedTypes: []types.Card{types.Land},
	})}
	if !reflect.DeepEqual(seq[1].Primitive, wantGroup) {
		t.Fatalf("group instruction = %#v, want %#v", seq[1].Primitive, wantGroup)
	}
	if !seq[1].ForEachPlayerGroup.Exists ||
		seq[1].ForEachPlayerGroup.Val.Kind != game.PlayerGroupReferenceOpponentsAttackingTriggerPlayer {
		t.Fatalf("group untap ForEachPlayerGroup = %#v, want opponents-attacking-trigger-player", seq[1].ForEachPlayerGroup)
	}
}

// TestLowerReflexiveAttackingRiderFailsClosed proves the reflexive rider stays
// closed for shapes it cannot represent, so no partial card is generated. An
// anaphoric "does the same." after an unsupported controller effect (life loss)
// and an explicit untap rider whose group differs from the controller untap both
// report diagnostics rather than silently dropping the rider.
func TestLowerReflexiveAttackingRiderFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		ability string
	}{
		{
			name:    "unsupported controller effect",
			ability: "Whenever enchanted player is attacked, you lose 2 life. Each opponent attacking that player does the same.",
		},
		{
			name:    "mismatched untap group",
			ability: "Whenever enchanted player is attacked, untap all creatures you control. Each opponent attacking that player untaps all nonland permanents they control.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
				Name:       "Unsupported " + test.name,
				Layout:     "normal",
				TypeLine:   "Enchantment — Aura Curse",
				OracleText: "Enchant player\n" + test.ability,
			})
		})
	}
}
