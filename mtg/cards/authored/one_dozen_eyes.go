package authored

import (
	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const oneDozenEyesOracleID = "b1fbf818-6699-4f05-9a91-19aa296526bf"

var (
	oneDozenEyesBeastProgram = game.ProgramReference{
		ID: oneDozenEyesOracleID + "/spell/0", Version: "1",
	}
	oneDozenEyesInsectProgram = game.ProgramReference{
		ID: oneDozenEyesOracleID + "/spell/1", Version: "1",
	}
)

var OneDozenEyes = cardset.Entry{
	OracleID: oneDozenEyesOracleID,
	Name:     "One Dozen Eyes",
	New:      newOneDozenEyes,
	Programs: []cardprogram.Program{
		{Reference: oneDozenEyesBeastProgram, Resolve: resolveOneDozenEyesBeast},
		{Reference: oneDozenEyesInsectProgram, Resolve: resolveOneDozenEyesInsects},
	},
	Verification: "mtg/rules/one_dozen_eyes_test.go: TestOneDozenEyesModesAndEntwine",
}

func newOneDozenEyes() *game.CardDef {
	const text = "Choose one —\n• Create a 5/5 green Beast creature token.\n• Create five 1/1 green Insect creature tokens.\nEntwine {G}{G}{G} (Choose both if you pay the entwine cost.)"
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:       "One Dozen Eyes",
			ManaCost:   opt.Val(cost.Mana{cost.O(5), cost.G}),
			Types:      []types.Card{types.Sorcery},
			Colors:     []color.Color{color.Green},
			OracleText: text,
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					{Text: "Create a 5/5 green Beast creature token.", Program: oneDozenEyesBeastProgram},
					{Text: "Create five 1/1 green Insect creature tokens.", Program: oneDozenEyesInsectProgram},
				},
				MinModes: 1,
				MaxModes: 2,
				// With only two modes, the second mode is exactly one additional entwine cost.
				EscalateCost: opt.Val(cost.Mana{cost.G, cost.G, cost.G}),
			}),
		},
	}
}

func resolveOneDozenEyesBeast(ctx cardprogram.Context) error {
	_, err := ctx.CreateTokens(
		oneDozenEyesToken("Beast Token", types.Beast, 5),
		cardprogram.TokenRecipient{Player: ctx.View().Source().Controller, Count: 1},
	)
	return err
}

func resolveOneDozenEyesInsects(ctx cardprogram.Context) error {
	_, err := ctx.CreateTokens(
		oneDozenEyesToken("Insect Token", types.Insect, 1),
		cardprogram.TokenRecipient{Player: ctx.View().Source().Controller, Count: 5},
	)
	return err
}

func oneDozenEyesToken(name string, subtype types.Sub, power int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name: name, Types: []types.Card{types.Creature}, Subtypes: []types.Sub{subtype},
		Colors:    []color.Color{color.Green},
		Power:     opt.Val(game.PT{Value: power}),
		Toughness: opt.Val(game.PT{Value: power}),
	}}
}
