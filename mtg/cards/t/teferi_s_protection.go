package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TeferiSProtection is the card definition for Teferi's Protection.
//
// Type: Instant
// Cost: {2}{W}
//
// Oracle text:
//
//	Until your next turn, your life total can't change and you gain protection from everything. All permanents you control phase out. (While they're phased out, they're treated as though they don't exist. They phase in before you untap during your untap step.)
//	Exile Teferi's Protection.
var TeferiSProtection = func() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Teferi's Protection",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
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
							Duration: game.DurationUntilYourNextTurn,
						},
					},
					{
						Primitive: game.ApplyRule{
							RuleEffects: []game.RuleEffect{
								game.RuleEffect{
									Kind:           game.RuleEffectPlayerProtection,
									AffectedPlayer: game.PlayerYou,
									Protection:     game.ProtectionKeyword{Everything: true},
								},
							},
							Duration: game.DurationUntilYourNextTurn,
						},
					},
					{
						Primitive: game.PhaseOut{
							Group: game.BattlefieldGroup(game.Selection{Controller: game.ControllerYou}),
						},
					},
					{
						Primitive: game.Exile{
							SourceSpell: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Until your next turn, your life total can't change and you gain protection from everything. All permanents you control phase out. (While they're phased out, they're treated as though they don't exist. They phase in before you untap during your untap step.)
			Exile Teferi's Protection.
		`,
		},
	}
}
