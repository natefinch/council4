package authored

import (
	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

var (
	deathDeniedProgram         = game.ProgramReference{ID: "94f23a93-cd98-4ee7-9e62-9b7b49ace6fd/spell/0", Version: "1"}
	wildestDreamsProgram       = game.ProgramReference{ID: "0a25a72e-5603-4b50-ac20-36b286cc1a2b/spell/0", Version: "1"}
	ratsFeastProgram           = game.ProgramReference{ID: "327c5ca2-4eeb-49cb-84fb-f6a4b06b24bc/spell/0", Version: "1"}
	shatteredCryptProgram      = game.ProgramReference{ID: "77bef290-7229-420a-a5c6-8c1b6fea63df/spell/0", Version: "1"}
	returnToTheRanksProgram    = game.ProgramReference{ID: "0d4ee553-f206-4e70-a7b2-a3670440aee7/spell/0", Version: "1"}
	championSoulsReturnProgram = game.ProgramReference{ID: "4edbaac4-dcb3-40b4-abd7-ff2683bd366d/activated/0", Version: "1"}
	championSoulsTopProgram    = game.ProgramReference{ID: "4edbaac4-dcb3-40b4-abd7-ff2683bd366d/activated/1", Version: "1"}
)

// DeathDenied returns every exact legal target, not a fixed-width template.
var DeathDenied = cardset.Entry{
	OracleID: "94f23a93-cd98-4ee7-9e62-9b7b49ace6fd", Name: "Death Denied", New: newDeathDenied,
	Programs:     []cardprogram.Program{{Reference: deathDeniedProgram, Resolve: resolveNumericHandReturn}},
	Verification: "mtg/rules/numeric_target_cards_test.go; TestExactNumericProductionDeathDeniedAboveLegacyWidth; CR 2026-09-25 107.3, 601.2, 608.2; #3284",
}

// WildestDreams preserves its resolving-spell exile after the targeted return.
var WildestDreams = cardset.Entry{
	OracleID: "0a25a72e-5603-4b50-ac20-36b286cc1a2b", Name: "Wildest Dreams", New: newWildestDreams,
	Programs:     []cardprogram.Program{{Reference: wildestDreamsProgram, Resolve: resolveWildestDreams}},
	Verification: "mtg/rules/numeric_target_cards_test.go; WotC 2016-09-20; CR 2026-09-25 107.3, 601.2, 608.2; #3284",
}

// RatsFeast exiles its chosen cards from one graveyard.
var RatsFeast = cardset.Entry{
	OracleID: "327c5ca2-4eeb-49cb-84fb-f6a4b06b24bc", Name: "Rats' Feast", New: newRatsFeast,
	Programs:     []cardprogram.Program{{Reference: ratsFeastProgram, Resolve: resolveRatsFeast}},
	Verification: "mtg/rules/numeric_target_cards_test.go; CR 2026-09-25 107.3, 601.2, 608.2; #3284",
}

// ShatteredCrypt loses the announced X life after returning legal targets.
var ShatteredCrypt = cardset.Entry{
	OracleID: "77bef290-7229-420a-a5c6-8c1b6fea63df", Name: "Shattered Crypt", New: newShatteredCrypt,
	Programs:     []cardprogram.Program{{Reference: shatteredCryptProgram, Resolve: resolveShatteredCrypt}},
	Verification: "mtg/rules/numeric_target_cards_test.go; CR 2026-09-25 107.3, 601.2, 608.2; #3284",
}

// ReturnToTheRanks retains convoke and simultaneous creature entry.
var ReturnToTheRanks = cardset.Entry{
	OracleID: "0d4ee553-f206-4e70-a7b2-a3670440aee7", Name: "Return to the Ranks", New: newReturnToTheRanks,
	Programs:     []cardprogram.Program{{Reference: returnToTheRanksProgram, Resolve: resolveNumericCreatureReturn}},
	Verification: "mtg/rules/numeric_target_cards_test.go; WotC 2024-01-12; CR 2026-09-25 107.3, 601.2, 702.51; #3284",
}

// ChampionOfStraySouls keeps both battlefield and graveyard activated abilities.
var ChampionOfStraySouls = cardset.Entry{
	OracleID: "4edbaac4-dcb3-40b4-abd7-ff2683bd366d", Name: "Champion of Stray Souls", New: newChampionOfStraySouls,
	Programs: []cardprogram.Program{
		{Reference: championSoulsReturnProgram, Resolve: resolveNumericCreatureReturn},
		{Reference: championSoulsTopProgram, Resolve: resolveChampionSoulsTop},
	},
	Verification: "mtg/rules/numeric_target_cards_test.go; WotC 2014-02-01; CR 2026-09-25 107.3, 602.2, 608.2; #3284",
}

func numericCardTargets(selection game.Selection) game.TargetSpec {
	return game.TargetSpec{
		MinTargets: 0, MaxTargets: 1, Allow: game.TargetAllowCard,
		TargetZone: zone.Graveyard, Selection: opt.Val(selection), CountEqualsX: true,
	}
}

func newDeathDenied() *game.CardDef {
	const text = "Return X target creature cards from your graveyard to your hand."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Death Denied", OracleText: text, ManaCost: opt.Val(cost.Mana{cost.X, cost.B, cost.B}),
			Types: []types.Card{types.Instant}, Subtypes: []types.Sub{types.Arcane}, Colors: []color.Color{color.Black},
			SpellAbility: opt.Val(game.Mode{Text: text, Program: deathDeniedProgram,
				Targets: []game.TargetSpec{numericCardTargets(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou})},
			}.Ability()),
		},
	}
}

func newWildestDreams() *game.CardDef {
	const text = "Return X target cards from your graveyard to your hand. Exile Wildest Dreams."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Wildest Dreams", OracleText: text, ManaCost: opt.Val(cost.Mana{cost.X, cost.X, cost.G}),
			Types: []types.Card{types.Sorcery}, Colors: []color.Color{color.Green},
			SpellAbility: opt.Val(game.Mode{Text: text, Program: wildestDreamsProgram,
				Targets: []game.TargetSpec{numericCardTargets(game.Selection{Controller: game.ControllerYou})},
			}.Ability()),
		},
	}
}

func newRatsFeast() *game.CardDef {
	const text = "Exile X target cards from a single graveyard."
	target := numericCardTargets(game.Selection{})
	target.SameGraveyard = true
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Rats' Feast", OracleText: text, ManaCost: opt.Val(cost.Mana{cost.X, cost.B}),
			Types: []types.Card{types.Sorcery}, Colors: []color.Color{color.Black},
			SpellAbility: opt.Val(game.Mode{Text: text, Program: ratsFeastProgram, Targets: []game.TargetSpec{target}}.Ability()),
		},
	}
}

func newShatteredCrypt() *game.CardDef {
	const text = "Return X target creature cards from your graveyard to your hand. You lose X life."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Shattered Crypt", OracleText: text, ManaCost: opt.Val(cost.Mana{cost.X, cost.B, cost.B}),
			Types: []types.Card{types.Sorcery}, Colors: []color.Color{color.Black},
			SpellAbility: opt.Val(game.Mode{Text: text, Program: shatteredCryptProgram,
				Targets: []game.TargetSpec{numericCardTargets(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou})},
			}.Ability()),
		},
	}
}

func newReturnToTheRanks() *game.CardDef {
	const text = "Convoke (Your creatures can help cast this spell. Each creature you tap while casting this spell pays for {1} or one mana of that creature's color.)\nReturn X target creature cards with mana value 2 or less from your graveyard to the battlefield."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Return to the Ranks", OracleText: text, ManaCost: opt.Val(cost.Mana{cost.X, cost.W, cost.W}),
			Types: []types.Card{types.Sorcery}, Colors: []color.Color{color.White},
			StaticAbilities: []game.StaticAbility{{KeywordAbilities: game.SimpleKeywords(game.Convoke)}},
			SpellAbility: opt.Val(game.Mode{Text: text, Program: returnToTheRanksProgram,
				Targets: []game.TargetSpec{numericCardTargets(game.Selection{
					RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou,
					ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 2}),
				})},
			}.Ability()),
		},
	}
}

func newChampionOfStraySouls() *game.CardDef {
	const first = "{3}{B}{B}, {T}, Sacrifice X other creatures: Return X target creature cards from your graveyard to the battlefield."
	const second = "{5}{B}{B}: Put this card from your graveyard on top of your library."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Champion of Stray Souls", OracleText: first + "\n" + second, ManaCost: opt.Val(cost.Mana{cost.O(4), cost.B, cost.B}),
			Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Skeleton, types.Warrior},
			Colors: []color.Color{color.Black}, Power: opt.Val(game.PT{Value: 4}), Toughness: opt.Val(game.PT{Value: 4}),
			ActivatedAbilities: []game.ActivatedAbility{
				{Text: first, ManaCost: opt.Val(cost.Mana{cost.O(3), cost.B, cost.B}),
					AdditionalCosts: []cost.Additional{
						{Kind: cost.AdditionalTap},
						{Kind: cost.AdditionalSacrifice, AmountFromX: true, ExcludeSource: true,
							MatchPermanentType: true, PermanentType: types.Creature},
					},
					Content: game.Mode{Text: first, Program: championSoulsReturnProgram,
						Targets: []game.TargetSpec{numericCardTargets(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou})},
					}.Ability(),
				},
				{Text: second, ManaCost: opt.Val(cost.Mana{cost.O(5), cost.B, cost.B}), ZoneOfFunction: zone.Graveyard,
					Content: game.Mode{Text: second, Program: championSoulsTopProgram}.Ability(),
				},
			},
		},
	}
}

func resolveNumericHandReturn(ctx cardprogram.Context) error {
	cards, err := ctx.View().TargetCards()
	if err != nil {
		return err
	}
	_, err = ctx.MoveCards(cardprogram.CardPlacement{Cards: cards, Destination: zone.Hand})
	return err
}

func resolveWildestDreams(ctx cardprogram.Context) error {
	if err := resolveNumericHandReturn(ctx); err != nil {
		return err
	}
	_, err := ctx.Execute(game.MoveResolvingSpell{Destination: zone.Exile})
	return err
}

func resolveShatteredCrypt(ctx cardprogram.Context) error {
	if err := resolveNumericHandReturn(ctx); err != nil {
		return err
	}
	_, err := ctx.Execute(game.LoseLife{Amount: game.Fixed(ctx.View().Source().X), Player: game.ControllerReference()})
	return err
}

func resolveRatsFeast(ctx cardprogram.Context) error {
	cards, err := ctx.View().TargetCards()
	if err != nil {
		return err
	}
	_, err = ctx.MoveCards(cardprogram.CardPlacement{Cards: cards, Destination: zone.Exile})
	return err
}

func resolveNumericCreatureReturn(ctx cardprogram.Context) error {
	var sources []game.BattlefieldSource
	for index, target := range ctx.View().Targets() {
		if target.Kind == game.TargetCard {
			sources = append(sources, game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: index}))
		}
	}
	if len(sources) == 0 {
		return nil
	}
	_, err := ctx.Execute(game.PutOnBattlefield{Sources: sources})
	return err
}

func resolveChampionSoulsTop(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.MoveCard{
		Card: game.CardReference{Kind: game.CardReferenceSource}, FromZone: zone.Graveyard, Destination: zone.Library,
	})
	return err
}
