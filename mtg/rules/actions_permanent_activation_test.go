package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// arrestSelfPermanent gives controller a battlefield permanent whose static
// ability pins itself: it can't attack or block and its own activated abilities
// can't be activated. When exemptMana is set the prohibition spares its mana
// abilities, mirroring Faith's Fetters. The rule effect self-scopes through
// AffectedSource so the prohibition matches only this permanent.
func arrestSelfPermanent(g *game.Game, controller game.PlayerID, exemptMana bool) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Arrested Permanent",
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{
				{
					Kind:                game.RuleEffectCantActivateAbilitiesOfPermanent,
					AffectedSource:      true,
					ExemptManaAbilities: exemptMana,
				},
			},
		}},
	}})
}

func TestAbilityActivationProhibitedByPermanentScope(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	pinned := arrestSelfPermanent(g, game.Player1, false)

	other := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Free Permanent", Types: []types.Card{types.Creature}}})

	if !abilityActivationProhibited(g, game.Player1, pinned, false) {
		t.Fatal("the pinned permanent's non-mana abilities should be prohibited")
	}
	if !abilityActivationProhibited(g, game.Player1, pinned, true) {
		t.Fatal("without a mana exemption even the pinned permanent's mana abilities are prohibited")
	}
	if abilityActivationProhibited(g, game.Player1, other, false) {
		t.Fatal("the permanent-scoped prohibition must match only the affected permanent")
	}
}

func TestPermanentActivationProhibitionSparesManaAbilities(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	pinned := arrestSelfPermanent(g, game.Player1, true)

	if !abilityActivationProhibited(g, game.Player1, pinned, false) {
		t.Fatal("the mana-exempt prohibition still forbids non-mana abilities")
	}
	if abilityActivationProhibited(g, game.Player1, pinned, true) {
		t.Fatal("the mana-exempt prohibition must spare mana abilities")
	}
}
