package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RecruitmentDrive is the card definition for Recruitment Drive.
//
// Type: Sorcery
// Cost: {2}{W}
//
// Oracle text:
//
//	Roll a d20.
//	1—9 | Create two 1/1 white Soldier creature tokens.
//	10—19 | Create two 2/2 white Knight creature tokens.
//	20 | Create three 2/2 white Knight creature tokens.
var RecruitmentDrive = newRecruitmentDrive

func newRecruitmentDrive() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Recruitment Drive",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.RollDie{
							Sides: 20,
						},
						PublishResult: game.ResultKey("die-roll-result"),
					},
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(2),
							Source: game.TokenDef(recruitmentDriveToken),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:         "die-roll-result",
							AmountRange: opt.Val(game.IntRange{Min: 1, Max: 9}),
						}),
					},
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(2),
							Source: game.TokenDef(recruitmentDriveToken2),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:         "die-roll-result",
							AmountRange: opt.Val(game.IntRange{Min: 10, Max: 19}),
						}),
					},
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(3),
							Source: game.TokenDef(recruitmentDriveToken2),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:         "die-roll-result",
							AmountRange: opt.Val(game.IntRange{Min: 20, Max: 20}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Roll a d20.
			1—9 | Create two 1/1 white Soldier creature tokens.
			10—19 | Create two 2/2 white Knight creature tokens.
			20 | Create three 2/2 white Knight creature tokens.
		`,
		},
	}
}

var recruitmentDriveToken = newRecruitmentDriveToken()

func newRecruitmentDriveToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Soldier",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}

var recruitmentDriveToken2 = newRecruitmentDriveToken2()

func newRecruitmentDriveToken2() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Knight",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
