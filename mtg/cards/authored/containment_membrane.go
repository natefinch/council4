package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const containmentMembraneOracleID = "e3bf1a95-4363-4915-8f43-b423714861ec"

var ContainmentMembrane = cardset.Entry{
	OracleID: containmentMembraneOracleID,
	Name:     "Containment Membrane",
	New:      newContainmentMembrane,
	Verification: "mtg/rules/containment_membrane_test.go: " +
		"TestContainmentMembraneSurgeAttachesAndStopsUntap",
}

func newContainmentMembrane() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Containment Membrane",
			ManaCost:   opt.Val(cost.Mana{cost.O(2), cost.U}),
			Colors:     []color.Color{color.Blue},
			Types:      []types.Card{types.Enchantment},
			Subtypes:   []types.Sub{types.Aura},
			OracleText: "Surge {U} (You may cast this spell for its surge cost if you or a teammate has cast another spell this turn.)\nEnchant creature\nEnchanted creature doesn't untap during its controller's untap step.",
			AlternativeCosts: []cost.Alternative{{
				Label:     "Surge {U}",
				ManaCost:  opt.Val(cost.Mana{cost.U}),
				Condition: cost.AlternativeConditionYourTeamCastSpellThisTurn,
			}},
			StaticAbilities: []game.StaticAbility{
				{
					Text: "Enchant creature",
					KeywordAbilities: []game.KeywordAbility{game.EnchantKeyword{
						Target: game.TargetSpec{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "target creature",
							Allow:      game.TargetAllowPermanent,
							Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
						},
					}},
				},
				{
					Text: "Enchanted creature doesn't untap during its controller's untap step.",
					RuleEffects: []game.RuleEffect{{
						Kind:             game.RuleEffectDoesntUntap,
						AffectedAttached: true,
					}},
				},
			},
		},
	}
}
