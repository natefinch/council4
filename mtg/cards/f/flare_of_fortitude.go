package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FlareOfFortitude is the card definition for Flare of Fortitude.
//
// Type: Instant
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	You may sacrifice a nontoken white creature rather than pay this spell's mana cost.
//	Until end of turn, your life total can't change, and permanents you control gain hexproof and indestructible.
var FlareOfFortitude = newFlareOfFortitude

func newFlareOfFortitude() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Flare of Fortitude",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Types: []types.Card{types.Instant},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label: "Sacrifice a nontoken white creature",
					AdditionalCosts: []cost.Additional{
						{
							Kind:               cost.AdditionalSacrifice,
							Text:               "sacrifice a nontoken white creature",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Creature,
							MatchCardColor:     true,
							CardColor:          color.White,
							RequireNonToken:    true,
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyRule{
							RuleEffects: []game.RuleEffect{
								game.RuleEffect{
									Kind:           game.RuleEffectLifeTotalCantChange,
									AffectedPlayer: game.PlayerYou,
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer: game.LayerAbility,
									Group: game.BattlefieldGroup(game.Selection{Controller: game.ControllerYou}),
									AddKeywords: []game.Keyword{
										game.Hexproof,
										game.Indestructible,
									},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			You may sacrifice a nontoken white creature rather than pay this spell's mana cost.
			Until end of turn, your life total can't change, and permanents you control gain hexproof and indestructible.
		`,
		},
	}
}
