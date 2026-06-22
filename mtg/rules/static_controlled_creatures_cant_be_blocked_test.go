package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestControlledCreaturesCantBeBlockedStaticGrantsMassEvasion models the runtime
// behavior of "Creatures you control can't be blocked." (the unconditional
// mass-evasion static): every creature the source's controller controls must be
// unblockable by every legal blocker, while creatures controlled by an opponent
// remain blockable. The effect comes from a battlefield static ability whose rule
// effect is scoped to the controller's creatures, so no per-creature effect is
// placed on the game.
func TestControlledCreaturesCantBeBlockedStaticGrantsMassEvasion(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Mass Evasion Enchantment",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:               game.RuleEffectCantBeBlocked,
				AffectedController: game.ControllerYou,
				PermanentTypes:     []types.Card{types.Creature},
			}},
		}},
	}})
	yourCreature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	otherYourCreature := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	opponentCreature := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	opponentBlocker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)

	for _, attacker := range []*game.Permanent{yourCreature, otherYourCreature} {
		if canBlockAttacker(g, blocker, attacker) {
			t.Fatal("controlled-creatures can't-be-blocked static let a creature you control be blocked")
		}
	}
	if !canBlockAttacker(g, opponentBlocker, opponentCreature) {
		t.Fatal("controlled-creatures can't-be-blocked static prevented blocking an opponent's creature")
	}
}
