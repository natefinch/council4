package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func addCanopyCoverAura(g *game.Game, controller game.PlayerID, enchanted *game.Permanent) *game.Permanent {
	aura := addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Canopy Cover",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{
			{
				KeywordAbilities: []game.KeywordAbility{game.EnchantKeyword{
					Target: game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				}},
			},
			{RuleEffects: []game.RuleEffect{{
				Kind:               game.RuleEffectCantBeBlockedExceptBy,
				AffectedAttached:   true,
				BlockerRestriction: game.BlockerRestriction{Kind: game.BlockerRestrictionFlyingOrReach},
			}}},
			{RuleEffects: []game.RuleEffect{{
				Kind:             game.RuleEffectCantBeTargetedByControllerOpponents,
				AffectedAttached: true,
			}}},
		},
	}})
	if !attachPermanent(g, aura, enchanted) {
		panic("failed to attach Canopy Cover")
	}
	return aura
}

func TestCanopyCoverAllowsOnlyFlyingOrReachBlockers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	addCanopyCoverAura(g, game.Player1, attacker)
	ground := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	flying := addCombatCreaturePermanentWithPower(g, game.Player2, 2, game.Flying)
	reach := addCombatCreaturePermanentWithPower(g, game.Player2, 2, game.Reach)

	if canBlockAttacker(g, ground, attacker) {
		t.Fatal("ground creature could block Canopy Cover attacker")
	}
	if !canBlockAttacker(g, flying, attacker) || !canBlockAttacker(g, reach, attacker) {
		t.Fatal("flying and reach creatures must be able to block Canopy Cover attacker")
	}
}

func TestCanopyCoverTargetRestrictionUsesAuraController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	enchanted := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	addCanopyCoverAura(g, game.Player1, enchanted)
	target := game.PermanentTarget(enchanted.ObjectID)

	if targetProtectedFromSource(g, game.Player1, nil, 0, target) {
		t.Fatal("Aura controller could not target the enchanted creature")
	}
	if !targetProtectedFromSource(g, game.Player2, nil, 0, target) {
		t.Fatal("enchanted creature's controller could target it despite opposing Canopy Cover")
	}
	if !targetProtectedFromSource(g, game.Player3, nil, 0, target) {
		t.Fatal("another opponent of the Aura controller could target the enchanted creature")
	}
}
