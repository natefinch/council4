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

const (
	labRatsOracleID           = "c53d6d3e-64e3-48b9-afab-3bde3164888f"
	mysticSpeculationOracleID = "a58adf82-d9e2-4b47-933e-68180f280ba2"
	sproutSwarmOracleID       = "8b33d890-7d67-44a8-a253-e2b171d7ca9d"

	labRatsOracleText = "Buyback {4} (You may pay an additional {4} as you cast this spell. If you do, put this card into your hand as it resolves.)\n" +
		"Create a 1/1 black Rat creature token."
	mysticSpeculationOracleText = "Buyback {2} (You may pay an additional {2} as you cast this spell. If you do, put this card into your hand as it resolves.)\n" +
		"Scry 3. (Look at the top three cards of your library, then put any number of them on the bottom and the rest on top in any order.)"
	sproutSwarmOracleText = "Convoke (Your creatures can help cast this spell. Each creature you tap while casting this spell pays for {1} or one mana of that creature's color.)\n" +
		"Buyback {3} (You may pay an additional {3} as you cast this spell. If you do, put this card into your hand as it resolves.)\n" +
		"Create a 1/1 green Saproling creature token."
)

var (
	labRatsProgram           = game.ProgramReference{ID: labRatsOracleID + "/spell/0", Version: "1"}
	mysticSpeculationProgram = game.ProgramReference{ID: mysticSpeculationOracleID + "/spell/0", Version: "1"}
	sproutSwarmProgram       = game.ProgramReference{ID: sproutSwarmOracleID + "/spell/0", Version: "1"}

	labRatsRatToken = &game.CardDef{CardFace: game.CardFace{
		Name: "Rat Token", Types: []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Rat}, Colors: []color.Color{color.Black},
		Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1}),
	}}
	sproutSwarmSaprolingToken = &game.CardDef{CardFace: game.CardFace{
		Name: "Saproling Token", Types: []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Saproling}, Colors: []color.Color{color.Green},
		Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1}),
	}}

	LabRats = cardset.Entry{
		OracleID: labRatsOracleID,
		Name:     "Lab Rats",
		New:      newLabRats,
		Programs: []cardprogram.Program{{Reference: labRatsProgram, Resolve: resolveLabRats}},
		Verification: "mtg/rules/buyback_misc_test.go: " +
			"TestLabRatsScenario",
	}
	MysticSpeculation = cardset.Entry{
		OracleID: mysticSpeculationOracleID,
		Name:     "Mystic Speculation",
		New:      newMysticSpeculation,
		Programs: []cardprogram.Program{{Reference: mysticSpeculationProgram, Resolve: resolveMysticSpeculation}},
		Verification: "mtg/rules/buyback_misc_test.go: " +
			"TestMysticSpeculationScenario",
	}
	SproutSwarm = cardset.Entry{
		OracleID: sproutSwarmOracleID,
		Name:     "Sprout Swarm",
		New:      newSproutSwarm,
		Programs: []cardprogram.Program{{Reference: sproutSwarmProgram, Resolve: resolveSproutSwarm}},
		Verification: "mtg/rules/buyback_misc_test.go: " +
			"TestSproutSwarmScenario",
	}
)

func newLabRats() *game.CardDef {
	const spellText = "Create a 1/1 black Rat creature token."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Lab Rats", OracleText: labRatsOracleText,
			ManaCost: opt.Val(cost.Mana{cost.B}),
			Types:    []types.Card{types.Sorcery},
			Colors:   []color.Color{color.Black},
			StaticAbilities: []game.StaticAbility{
				buybackStaticWithCost(4),
			},
			SpellAbility: opt.Val(game.Mode{
				Text: spellText, Program: labRatsProgram,
			}.Ability()),
		},
	}
}

func newMysticSpeculation() *game.CardDef {
	const spellText = "Scry 3."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Mystic Speculation", OracleText: mysticSpeculationOracleText,
			ManaCost: opt.Val(cost.Mana{cost.U}),
			Types:    []types.Card{types.Sorcery},
			Colors:   []color.Color{color.Blue},
			StaticAbilities: []game.StaticAbility{
				buybackStaticWithCost(2),
			},
			SpellAbility: opt.Val(game.Mode{
				Text: spellText, Program: mysticSpeculationProgram,
			}.Ability()),
		},
	}
}

func newSproutSwarm() *game.CardDef {
	const spellText = "Create a 1/1 green Saproling creature token."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Sprout Swarm", OracleText: sproutSwarmOracleText,
			ManaCost: opt.Val(cost.Mana{cost.O(1), cost.G}),
			Types:    []types.Card{types.Instant},
			Colors:   []color.Color{color.Green},
			StaticAbilities: []game.StaticAbility{
				game.ConvokeStaticBody,
				buybackStaticWithCost(3),
			},
			SpellAbility: opt.Val(game.Mode{
				Text: spellText, Program: sproutSwarmProgram,
			}.Ability()),
		},
	}
}

func resolveLabRats(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.CreateToken{
		Amount:    game.Fixed(1),
		Source:    game.TokenDef(labRatsRatToken),
		Recipient: opt.Val(game.ControllerReference()),
	})
	return err
}

func resolveMysticSpeculation(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.Scry{
		Player: game.ControllerReference(),
		Amount: game.Fixed(3),
	})
	return err
}

func resolveSproutSwarm(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.CreateToken{
		Amount:    game.Fixed(1),
		Source:    game.TokenDef(sproutSwarmSaprolingToken),
		Recipient: opt.Val(game.ControllerReference()),
	})
	return err
}
