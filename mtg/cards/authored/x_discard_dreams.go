package authored

import (
	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

var (
	restlessDreamsProgram  = game.ProgramReference{ID: "65f1dd09-0640-47f8-8bfd-7aceaa5646a6/spell/0", Version: "1"}
	turbulentDreamsProgram = game.ProgramReference{ID: "83134570-debe-4bc0-9656-2a16801a3592/spell/0", Version: "1"}
	nostalgicDreamsProgram = game.ProgramReference{ID: "fabbc474-682c-4565-861c-d41de05abb8f/spell/0", Version: "1"}
)

var RestlessDreams = cardset.Entry{
	OracleID: "65f1dd09-0640-47f8-8bfd-7aceaa5646a6", Name: "Restless Dreams", New: newRestlessDreams,
	Programs:     []cardprogram.Program{{Reference: restlessDreamsProgram, Resolve: resolveRestlessDreams}},
	Verification: "mtg/rules/x_discard_dreams_test.go: TestRestlessDreamsDiscardsXAndReturnsCreatureCards",
}

var TurbulentDreams = cardset.Entry{
	OracleID: "83134570-debe-4bc0-9656-2a16801a3592", Name: "Turbulent Dreams", New: newTurbulentDreams,
	Programs:     []cardprogram.Program{{Reference: turbulentDreamsProgram, Resolve: resolveTurbulentDreams}},
	Verification: "mtg/rules/x_discard_dreams_test.go: TestTurbulentDreamsDiscardsXAndReturnsNonlandPermanents",
}

var NostalgicDreams = cardset.Entry{
	OracleID: "fabbc474-682c-4565-861c-d41de05abb8f", Name: "Nostalgic Dreams", New: newNostalgicDreams,
	Programs:     []cardprogram.Program{{Reference: nostalgicDreamsProgram, Resolve: resolveNostalgicDreams}},
	Verification: "mtg/rules/x_discard_dreams_test.go: TestNostalgicDreamsDiscardsXReturnsCardsAndExiles",
}

func newRestlessDreams() *game.CardDef {
	const text = "As an additional cost to cast this spell, discard X cards.\nReturn X target creature cards from your graveyard to your hand."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Restless Dreams", OracleText: text,
			ManaCost: opt.Val(cost.Mana{cost.B}),
			Types:    []types.Card{types.Sorcery}, Colors: []color.Color{color.Black},
			AdditionalCosts: []cost.Additional{{
				Kind: cost.AdditionalDiscard, Text: "discard X cards", AmountFromX: true, Source: zone.Hand,
			}},
			SpellAbility: opt.Val(game.Mode{
				Text: text, Program: restlessDreamsProgram,
				Targets: []game.TargetSpec{dreamsGraveyardTargets(
					"target creature cards from your graveyard",
					game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou},
				)},
			}.Ability()),
		},
	}
}

func newTurbulentDreams() *game.CardDef {
	const text = "As an additional cost to cast this spell, discard X cards.\nReturn X target nonland permanents to their owners' hands."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Turbulent Dreams", OracleText: text,
			ManaCost: opt.Val(cost.Mana{cost.U, cost.U}),
			Types:    []types.Card{types.Sorcery}, Colors: []color.Color{color.Blue},
			AdditionalCosts: []cost.Additional{{
				Kind: cost.AdditionalDiscard, Text: "discard X cards", AmountFromX: true, Source: zone.Hand,
			}},
			SpellAbility: opt.Val(game.Mode{
				Text: text, Program: turbulentDreamsProgram,
				Targets: []game.TargetSpec{{
					MinTargets: 0, MaxTargets: 1, CountEqualsX: true,
					Constraint: "target nonland permanents", Allow: game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Land}}),
				}},
			}.Ability()),
		},
	}
}

func newNostalgicDreams() *game.CardDef {
	const text = "As an additional cost to cast this spell, discard X cards.\nReturn X target cards from your graveyard to your hand. Exile Nostalgic Dreams."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Nostalgic Dreams", OracleText: text,
			ManaCost: opt.Val(cost.Mana{cost.G, cost.G}),
			Types:    []types.Card{types.Sorcery}, Colors: []color.Color{color.Green},
			AdditionalCosts: []cost.Additional{{
				Kind: cost.AdditionalDiscard, Text: "discard X cards", AmountFromX: true, Source: zone.Hand,
			}},
			SpellAbility: opt.Val(game.Mode{
				Text: text, Program: nostalgicDreamsProgram,
				Targets: []game.TargetSpec{dreamsGraveyardTargets(
					"target cards from your graveyard",
					game.Selection{Controller: game.ControllerYou},
				)},
			}.Ability()),
		},
	}
}

func dreamsGraveyardTargets(constraint string, selection game.Selection) game.TargetSpec {
	return game.TargetSpec{
		MinTargets: 0, MaxTargets: 1, CountEqualsX: true,
		Constraint: constraint, Allow: game.TargetAllowCard,
		TargetZone: zone.Graveyard, Selection: opt.Val(selection),
	}
}

func resolveRestlessDreams(ctx cardprogram.Context) error {
	return resolveNumericHandReturn(ctx)
}

func resolveTurbulentDreams(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.MovePermanent{
		Object: game.AllTargetPermanentsReference(0), Destination: zone.Hand,
	})
	return err
}

func resolveNostalgicDreams(ctx cardprogram.Context) error {
	if err := resolveNumericHandReturn(ctx); err != nil {
		return err
	}
	_, err := ctx.Execute(game.MoveResolvingSpell{Destination: zone.Exile})
	return err
}
