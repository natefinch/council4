package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestStaticEquipCostModifierGatedByMetalcraft proves the generic "Equipment you
// control have equip {0}" static cost-setting, gated by the Metalcraft
// "as long as you control three or more artifacts" condition: the Equip
// activation cost drops to {0} only while the controller has three or more
// artifacts, and is otherwise unchanged.
func TestStaticEquipCostModifierGatedByMetalcraft(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Puresteel Paladin",
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{{
			Text: "Equipment you control have equip {0} as long as you control three or more artifacts.",
			Condition: opt.Val(game.Condition{
				ControlsMatching: opt.Val(game.SelectionCount{
					Selection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
					MinCount:  3,
				}),
			}),
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectCostModifier,
				AffectedPlayer: game.PlayerYou,
				CostModifier: game.CostModifier{
					Kind:           game.CostModifierAbility,
					AbilityKeyword: game.Equip,
					SetManaCost:    opt.Val(cost.Mana(nil)),
				},
			}},
		}},
	}})

	equipBody := game.EquipActivatedAbility(cost.Mana{cost.O(3)})

	if got := effectiveActivatedAbilityCost(g, game.Player1, nil, &equipBody); manaString(got) != "{3}" {
		t.Fatalf("equip cost without Metalcraft = %q, want {3}", manaString(got))
	}

	for range 3 {
		addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name:  "Relic",
			Types: []types.Card{types.Artifact},
		}})
	}

	if got := effectiveActivatedAbilityCost(g, game.Player1, nil, &equipBody); manaString(got) != "" {
		t.Fatalf("equip cost with three artifacts = %q, want {0} (empty)", manaString(got))
	}
}

func manaString(v opt.V[cost.Mana]) string {
	if !v.Exists {
		return ""
	}
	return v.Val.String()
}
