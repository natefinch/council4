package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
)

const gylwainOracleText = "Whenever Gylwain or another nontoken creature you control enters, choose one —\n" +
	"• Create a Royal Role token attached to that creature.\n" +
	"• Create a Sorcerer Role token attached to that creature.\n" +
	"• Create a Monster Role token attached to that creature."

func TestLowerGylwainCastingDirector(t *testing.T) {
	t.Parallel()

	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Gylwain, Casting Director",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Human Bard",
		OracleText: gylwainOracleText,
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %#v, want one", face.TriggeredAbilities)
	}
	ability := face.TriggeredAbilities[0]
	pattern := ability.Trigger.Pattern
	if pattern.Event != game.EventPermanentEnteredBattlefield ||
		pattern.Controller != game.TriggerControllerYou ||
		!pattern.SubjectSelectionOrSelf ||
		!slices.Equal(pattern.SubjectSelection.RequiredTypes, []types.Card{types.Creature}) ||
		!pattern.SubjectSelection.NonToken {
		t.Fatalf("trigger pattern = %#v", pattern)
	}
	content := ability.Content
	if content.RandomModes ||
		content.MinModes != 1 ||
		content.MaxModes != 1 ||
		len(content.Modes) != 3 {
		t.Fatalf("modal content = %#v", content)
	}
	for i, wantName := range []string{"Royal Role", "Sorcerer Role", "Monster Role"} {
		mode := content.Modes[i]
		if len(mode.Targets) != 0 || len(mode.Sequence) != 1 {
			t.Fatalf("mode %d = %#v", i, mode)
		}
		create, ok := mode.Sequence[0].Primitive.(game.CreateToken)
		if !ok || !create.EntryAttachedTo.Exists ||
			create.EntryAttachedTo.Val != game.EventPermanentReference() {
			t.Fatalf("mode %d create = %#v", i, mode.Sequence[0].Primitive)
		}
		token, ok := create.Source.TokenDefRef()
		if !ok || token.Name != wantName ||
			!slices.Equal(token.Types, []types.Card{types.Enchantment}) ||
			!slices.Equal(token.Subtypes, []types.Sub{types.Aura, types.Role}) {
			t.Fatalf("mode %d token = %#v", i, token)
		}
		if issues := game.ValidateCardDef(token); len(issues) != 0 {
			t.Fatalf("mode %d token validation = %+v", i, issues)
		}
	}
}

func TestPredefinedRoleTokenDefinitions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		def  *game.CardDef
	}{
		{"Cursed Role", cursedRoleTokenDef()},
		{"Monster Role", monsterRoleTokenDef()},
		{"Royal Role", royalRoleTokenDef()},
		{"Sorcerer Role", sorcererRoleTokenDef()},
		{"Virtuous Role", virtuousRoleTokenDef()},
		{"Wicked Role", wickedRoleTokenDef()},
		{"Young Hero Role", youngHeroRoleTokenDef()},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if test.def.Name != test.name ||
				!slices.Equal(test.def.Types, []types.Card{types.Enchantment}) ||
				!slices.Equal(test.def.Subtypes, []types.Sub{types.Aura, types.Role}) ||
				len(test.def.StaticAbilities) < 2 {
				t.Fatalf("definition = %#v", test.def)
			}
			if issues := game.ValidateCardDef(test.def); len(issues) != 0 {
				t.Fatalf("validation = %+v", issues)
			}
		})
	}

	royalGrant := royalRoleTokenDef().StaticAbilities[1].ContinuousEffects[0]
	ward, ok := royalGrant.AddAbilities[0].(*game.StaticAbility)
	if !ok {
		t.Fatalf("Royal Role grant = %#v", royalGrant.AddAbilities)
	}
	wardCost, ok := game.StaticBodyWardCost(ward)
	if !ok || !slices.Equal(wardCost, cost.Mana{cost.O(1)}) {
		t.Fatalf("Royal Role ward = %v, %v", wardCost, ok)
	}
}
