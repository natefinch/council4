package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const (
	boulderSalvoOracleID    = "16df20e2-5228-4543-84b7-74a3170d9d31"
	fallOfTheTitansOracleID = "55e564b4-7c08-4109-bd67-da5f31fe002e"
)

var BoulderSalvo = cardset.Entry{
	OracleID: boulderSalvoOracleID,
	Name:     "Boulder Salvo",
	New:      newBoulderSalvo,
	Verification: "mtg/rules/surge_damage_spells_test.go: " +
		"TestBoulderSalvoSurgeTargetCreatureDamage",
}

func newBoulderSalvo() *game.CardDef {
	const text = "Surge {1}{R} (You may cast this spell for its surge cost if you or a teammate has cast another spell this turn.)\nBoulder Salvo deals 4 damage to target creature."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:       "Boulder Salvo",
			ManaCost:   opt.Val(cost.Mana{cost.O(4), cost.R}),
			Colors:     []color.Color{color.Red},
			Types:      []types.Card{types.Sorcery},
			OracleText: text,
			AlternativeCosts: []cost.Alternative{{
				Label:     "Surge {1}{R}",
				ManaCost:  opt.Val(cost.Mana{cost.O(1), cost.R}),
				Condition: cost.AlternativeConditionYourTeamCastSpellThisTurn,
			}},
			SpellAbility: opt.Val(game.Mode{
				Text: "Boulder Salvo deals 4 damage to target creature.",
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target creature",
					Allow:      game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{
						RequiredTypes: []types.Card{types.Creature},
					}),
				}},
				Sequence: []game.Instruction{{Primitive: game.Damage{
					Amount:    game.Fixed(4),
					Recipient: game.ObjectDamageRecipient(game.TargetPermanentReference(0)),
				}}},
			}.Ability()),
		},
	}
}

var FallOfTheTitans = cardset.Entry{
	OracleID: fallOfTheTitansOracleID,
	Name:     "Fall of the Titans",
	New:      newFallOfTheTitans,
	Verification: "mtg/rules/surge_damage_spells_test.go: " +
		"TestFallOfTheTitansSurgeEachTargetXDamage",
}

func newFallOfTheTitans() *game.CardDef {
	const text = "Surge {X}{R} (You may cast this spell for its surge cost if you or a teammate has cast another spell this turn.)\nFall of the Titans deals X damage to each of up to two targets."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:       "Fall of the Titans",
			ManaCost:   opt.Val(cost.Mana{cost.X, cost.X, cost.R}),
			Colors:     []color.Color{color.Red},
			Types:      []types.Card{types.Instant},
			OracleText: text,
			AlternativeCosts: []cost.Alternative{{
				Label:     "Surge {X}{R}",
				ManaCost:  opt.Val(cost.Mana{cost.X, cost.R}),
				Condition: cost.AlternativeConditionYourTeamCastSpellThisTurn,
			}},
			SpellAbility: opt.Val(game.Mode{
				Text: "Fall of the Titans deals X damage to each of up to two targets.",
				Targets: []game.TargetSpec{{
					MinTargets: 0,
					MaxTargets: 2,
					Constraint: "any target",
					Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
					Selection: opt.Val(game.Selection{
						RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle},
					}),
				}},
				Sequence: []game.Instruction{{Primitive: game.Damage{
					Amount:     game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}),
					Recipient:  game.AnyTargetDamageRecipient(0),
					EachTarget: true,
				}}},
			}.Ability()),
		},
	}
}
