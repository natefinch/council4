package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

// thatPlayerControlsDamage extracts the single Damage primitive from a lowered
// "that player controls" damage ability sequence and returns its dynamic count
// group so the "that player controls" damage tests can assert the count is
// scoped to the damaged target player.
func thatPlayerControlsDamage(t *testing.T, damage game.Damage, wantMultiplier int) game.GroupReference {
	t.Helper()
	if !damage.Amount.IsDynamic() {
		t.Fatalf("damage amount = %#v, want dynamic count", damage.Amount)
	}
	dynamic := damage.Amount.DynamicAmount().Val
	if dynamic.Kind != game.DynamicAmountCountSelector {
		t.Fatalf("dynamic kind = %v, want DynamicAmountCountSelector", dynamic.Kind)
	}
	if dynamic.Multiplier != wantMultiplier {
		t.Fatalf("multiplier = %d, want %d", dynamic.Multiplier, wantMultiplier)
	}
	group := dynamic.Group
	if group.Domain() != game.GroupDomainPlayerControlled {
		t.Fatalf("group domain = %v, want GroupDomainPlayerControlled", group.Domain())
	}
	anchor, ok := group.PlayerAnchor()
	if !ok {
		t.Fatal("group has no player anchor")
	}
	if anchor.Kind() != game.PlayerReferenceTargetPlayer || anchor.TargetIndex() != 0 {
		t.Fatalf("player anchor = %#v, want target player 0", anchor)
	}
	recipient, ok := damage.Recipient.AnyTargetPlayerReference()
	if !ok {
		t.Fatalf("recipient = %#v, want any-target player", damage.Recipient)
	}
	if recipient.Kind() != game.PlayerReferenceTargetPlayer || recipient.TargetIndex() != 0 {
		t.Fatalf("recipient player = %#v, want target player 0", recipient)
	}
	return group
}

// TestLowerAnathemancerThatPlayerControlsDamage proves Anathemancer's ETB
// "deals damage to target player equal to the number of nonbasic lands that
// player controls" lowers to a Damage whose dynamic count is scoped to the
// target player's nonbasic lands.
func TestLowerAnathemancerThatPlayerControlsDamage(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:          "Anathemancer",
		Layout:        "normal",
		ManaCost:      "{1}{B}{R}",
		TypeLine:      "Creature — Zombie Wizard",
		ColorIdentity: []string{"B", "R"},
		OracleText:    "When this creature enters, it deals damage to target player equal to the number of nonbasic lands that player controls.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	sequence := face.TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(sequence) != 1 {
		t.Fatalf("sequence = %#v, want one Damage instruction", sequence)
	}
	damage, ok := sequence[0].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("primitive = %#v, want Damage", sequence[0].Primitive)
	}
	group := thatPlayerControlsDamage(t, damage, 1)
	selection := group.Selection()
	if len(selection.RequiredTypes) != 1 || selection.RequiredTypes[0] != types.Land {
		t.Fatalf("selection required types = %#v, want [Land]", selection.RequiredTypes)
	}
	if selection.ExcludedSupertype != types.Basic {
		t.Fatalf("selection excluded supertype = %v, want Basic", selection.ExcludedSupertype)
	}
}

// TestLowerJovialEvilThatPlayerControlsDamage proves Jovial Evil's "deals X
// damage to target opponent, where X is twice the number of white creatures
// that player controls" lowers to a Damage whose dynamic count is twice the
// target opponent's white creatures.
func TestLowerJovialEvilThatPlayerControlsDamage(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:          "Jovial Evil",
		Layout:        "normal",
		ManaCost:      "{2}{B}",
		TypeLine:      "Sorcery",
		ColorIdentity: []string{"B"},
		OracleText:    "Jovial Evil deals X damage to target opponent, where X is twice the number of white creatures that player controls.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("Jovial Evil did not lower a spell ability")
	}
	sequence := face.SpellAbility.Val.Modes[0].Sequence
	if len(sequence) != 1 {
		t.Fatalf("sequence = %#v, want one Damage instruction", sequence)
	}
	damage, ok := sequence[0].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("primitive = %#v, want Damage", sequence[0].Primitive)
	}
	group := thatPlayerControlsDamage(t, damage, 2)
	selection := group.Selection()
	if len(selection.RequiredTypes) != 1 || selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("selection required types = %#v, want [Creature]", selection.RequiredTypes)
	}
	if len(selection.ColorsAny) != 1 || selection.ColorsAny[0] != color.White {
		t.Fatalf("selection colors = %#v, want [White]", selection.ColorsAny)
	}
}

// TestLowerGroupThatPlayerControlsDamageFailsClosed proves the group-recipient
// "that player controls" damage cards stay fail-closed: the executable damage
// backend resolves a single group-wide amount for every recipient, so it cannot
// express a per-recipient count ("each player equal to ... that player
// controls"). These must not lower to a spell ability.
func TestLowerGroupThatPlayerControlsDamageFailsClosed(t *testing.T) {
	t.Parallel()
	cards := []struct {
		name          string
		manaCost      string
		typeLine      string
		colorIdentity []string
		oracleText    string
	}{
		{
			name:          "Price of Progress",
			manaCost:      "{1}{R}",
			typeLine:      "Instant",
			colorIdentity: []string{"R"},
			oracleText:    "Price of Progress deals damage to each player equal to twice the number of nonbasic lands that player controls.",
		},
		{
			name:          "Treacherous Terrain",
			manaCost:      "{6}{R}{G}",
			typeLine:      "Sorcery",
			colorIdentity: []string{"R", "G"},
			oracleText:    "Treacherous Terrain deals damage to each opponent equal to the number of lands that player controls.",
		},
		{
			name:          "Typhoon",
			manaCost:      "{2}{G}",
			typeLine:      "Sorcery",
			colorIdentity: []string{"G"},
			oracleText:    "Typhoon deals damage to each opponent equal to the number of Islands that player controls.",
		},
	}
	for _, tc := range cards {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
				Name:          tc.name,
				Layout:        "normal",
				ManaCost:      tc.manaCost,
				TypeLine:      tc.typeLine,
				ColorIdentity: tc.colorIdentity,
				OracleText:    tc.oracleText,
			})
			if face.SpellAbility.Exists {
				t.Fatalf("%s lowered a spell ability; group per-recipient count must fail closed", tc.name)
			}
		})
	}
}
