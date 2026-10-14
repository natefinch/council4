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

var generousGiftProgram = game.ProgramReference{ID: "fae37e28-e137-4177-b973-fa8b4dd8f409/spell/0", Version: "1"}
var pongifyProgram = game.ProgramReference{ID: "05849bd6-8f38-4031-be2b-e2aa03beb8cc/spell/0", Version: "1"}
var rapidHybridizationProgram = game.ProgramReference{ID: "06692cd9-ac2f-4a32-8fd1-043ba3c0fe71/spell/0", Version: "1"}
var resculptProgram = game.ProgramReference{ID: "814e321d-16fd-4f7b-a8d8-2b089be76f2c/spell/0", Version: "1"}

// GenerousGift uses Oracle printing revision c3f9b454-97e9-4c77-b37a-e974221ae385.
var GenerousGift = cardset.Entry{
	OracleID: "fae37e28-e137-4177-b973-fa8b4dd8f409", Name: "Generous Gift", New: newGenerousGift,
	Programs:     []cardprogram.Program{{Reference: generousGiftProgram, Resolve: resolveGenerousGift}},
	Verification: "mtg/rules/removal_token_spells_test.go: TestAuthoredGenerousGiftScenario; TestAuthoredRemovalTokensProtectedTargets/Generous_Gift; TestAuthoredRemovalTokensAllIllegalTargets/Generous_Gift",
}

// Pongify uses Oracle printing revision 4131fa63-5afc-4b63-a4b4-f47b4bced87f.
var Pongify = cardset.Entry{
	OracleID: "05849bd6-8f38-4031-be2b-e2aa03beb8cc", Name: "Pongify", New: newPongify,
	Programs:     []cardprogram.Program{{Reference: pongifyProgram, Resolve: resolvePongify}},
	Verification: "mtg/rules/removal_token_spells_test.go: TestAuthoredRemovalTokensCastAndTargetTypes/Pongify; TestAuthoredRemovalTokensProtectedTargets/Pongify; TestAuthoredRemovalTokensAllIllegalTargets/Pongify",
}

// RapidHybridization uses Oracle printing revision ccda5bda-5afd-4548-bd48-12c780bfa68b.
var RapidHybridization = cardset.Entry{
	OracleID: "06692cd9-ac2f-4a32-8fd1-043ba3c0fe71", Name: "Rapid Hybridization", New: newRapidHybridization,
	Programs:     []cardprogram.Program{{Reference: rapidHybridizationProgram, Resolve: resolveRapidHybridization}},
	Verification: "mtg/rules/removal_token_spells_test.go: TestAuthoredRemovalTokensCastAndTargetTypes/Rapid_Hybridization; TestAuthoredRemovalTokensProtectedTargets/Rapid_Hybridization; TestAuthoredRemovalTokensAllIllegalTargets/Rapid_Hybridization",
}

// Resculpt uses Oracle printing revision 5443f46d-03f9-468e-a78d-a059529b1151.
var Resculpt = cardset.Entry{
	OracleID: "814e321d-16fd-4f7b-a8d8-2b089be76f2c", Name: "Resculpt", New: newResculpt,
	Programs:     []cardprogram.Program{{Reference: resculptProgram, Resolve: resolveResculpt}},
	Verification: "mtg/rules/removal_token_spells_test.go: TestAuthoredRemovalTokensCastAndTargetTypes/Resculpt; TestAuthoredRemovalTokensControllerSnapshot/Resculpt; TestAuthoredRemovalTokensZoneAndTokenReplacements/Resculpt; TestAuthoredRemovalTokensAllIllegalTargets/Resculpt",
}

func newGenerousGift() *game.CardDef {
	const text = "Destroy target permanent. Its controller creates a 3/3 green Elephant creature token."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Generous Gift", OracleText: text, ManaCost: opt.Val(cost.Mana{cost.O(2), cost.W}),
			Types: []types.Card{types.Instant}, Colors: []color.Color{color.White},
			SpellAbility: opt.Val(game.Mode{
				Text: text, Program: generousGiftProgram,
				Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent}},
			}.Ability()),
		},
	}
}

func newPongify() *game.CardDef {
	const text = "Destroy target creature. It can't be regenerated. Its controller creates a 3/3 green Ape creature token."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Pongify", OracleText: text, ManaCost: opt.Val(cost.Mana{cost.U}),
			Types: []types.Card{types.Instant}, Colors: []color.Color{color.Blue},
			SpellAbility: opt.Val(game.Mode{
				Text: text, Program: pongifyProgram,
				Targets: []game.TargetSpec{{
					MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
				}},
			}.Ability()),
		},
	}
}

func newRapidHybridization() *game.CardDef {
	const text = "Destroy target creature. It can't be regenerated. That creature's controller creates a 3/3 green Frog Lizard creature token."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Rapid Hybridization", OracleText: text, ManaCost: opt.Val(cost.Mana{cost.U}),
			Types: []types.Card{types.Instant}, Colors: []color.Color{color.Blue},
			SpellAbility: opt.Val(game.Mode{
				Text: text, Program: rapidHybridizationProgram,
				Targets: []game.TargetSpec{{
					MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
				}},
			}.Ability()),
		},
	}
}

func newResculpt() *game.CardDef {
	const text = "Exile target artifact or creature. Its controller creates a 4/4 blue and red Elemental creature token."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Resculpt", OracleText: text, ManaCost: opt.Val(cost.Mana{cost.O(1), cost.U}),
			Types: []types.Card{types.Instant}, Colors: []color.Color{color.Blue},
			SpellAbility: opt.Val(game.Mode{
				Text: text, Program: resculptProgram,
				Targets: []game.TargetSpec{{
					MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature}}),
				}},
			}.Ability()),
		},
	}
}

func resolveGenerousGift(ctx cardprogram.Context) error {
	return destroyAndCreateToken(ctx, false, removalCreatureToken("Elephant Token", 3, []color.Color{color.Green}, types.Elephant))
}

func resolvePongify(ctx cardprogram.Context) error {
	return destroyAndCreateToken(ctx, true, removalCreatureToken("Ape Token", 3, []color.Color{color.Green}, types.Ape))
}

func resolveRapidHybridization(ctx cardprogram.Context) error {
	return destroyAndCreateToken(ctx, true, removalCreatureToken("Frog Lizard Token", 3, []color.Color{color.Green}, types.Frog, types.Lizard))
}

func destroyAndCreateToken(ctx cardprogram.Context, preventRegeneration bool, token *game.CardDef) error {
	outcome, err := ctx.Destroy(game.Destroy{
		Object: game.TargetPermanentReference(0), PreventRegeneration: preventRegeneration,
	})
	if err != nil {
		return err
	}
	// A legal target earns a token even if destruction was prevented or replaced.
	for _, member := range outcome.Members {
		controller := member.Before.Controller
		if !member.Moved {
			current, available, err := ctx.View().Player(game.ObjectControllerReference(game.TargetPermanentReference(0)))
			if err != nil {
				return err
			}
			if available {
				controller = current
			}
		}
		if _, err := ctx.CreateTokens(token, cardprogram.TokenRecipient{Player: controller, Count: 1}); err != nil {
			return err
		}
	}
	return nil
}

func resolveResculpt(ctx cardprogram.Context) error {
	target := game.TargetPermanentReference(0)
	controller, available, err := ctx.View().Player(game.ObjectControllerReference(target))
	if err != nil || !available {
		return err
	}
	if _, err := ctx.Execute(game.MovePermanent{Object: target, Destination: zone.Exile}); err != nil {
		return err
	}
	_, err = ctx.CreateTokens(
		removalCreatureToken("Elemental Token", 4, []color.Color{color.Blue, color.Red}, types.Elemental),
		cardprogram.TokenRecipient{Player: controller, Count: 1},
	)
	return err
}

func removalCreatureToken(name string, size int, colors []color.Color, subtypes ...types.Sub) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name: name, Types: []types.Card{types.Creature}, Subtypes: subtypes, Colors: colors,
		Power: opt.Val(game.PT{Value: size}), Toughness: opt.Val(game.PT{Value: size}),
	}}
}
