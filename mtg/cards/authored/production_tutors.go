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
	eladamrisCallProgram    = game.ProgramReference{ID: "4acb6612-54e8-428d-acb6-c7259a5ad6a8/spell/0", Version: "1"}
	steelshapersGiftProgram = game.ProgramReference{ID: "d9abda7e-6ca2-42ea-ab24-c542e57014f1/spell/0", Version: "1"}
	fabricateProgram        = game.ProgramReference{ID: "422e1869-134f-463d-9fa1-86b66a998b3e/spell/0", Version: "1"}
	idyllicTutorProgram     = game.ProgramReference{ID: "57c9ff89-dd30-467b-bb3e-499eeea8cb94/spell/0", Version: "1"}
	sylvanScryingProgram    = game.ProgramReference{ID: "ee24bf27-484d-4e1c-998e-6a74e3d3f6c4/spell/0", Version: "1"}
)

// EladamrisCall is an instant-speed qualified search with optional fail-to-find.
var EladamrisCall = cardset.Entry{
	OracleID: "4acb6612-54e8-428d-acb6-c7259a5ad6a8", Name: "Eladamri's Call", New: newEladamrisCall,
	Programs:     []cardprogram.Program{{Reference: eladamrisCallProgram, Resolve: resolveEladamrisCall}},
	Verification: "mtg/rules/production_tutors_test.go: TestProductionTutorsSearchChoicesAndVisibility/Eladamri's_Call, TestProductionTutorsCastTimingAndPayment, TestProductionTutorsReplacementAndSearchControl, TestProductionTutorsCopyAndClone; TestProductionBatchLiveAndSimulator; WotC 2004-10-04; CR 2026-09-25 701.23b/e",
}

// SteelshapersGift searches the Equipment subtype, not all artifacts.
var SteelshapersGift = cardset.Entry{
	OracleID: "d9abda7e-6ca2-42ea-ab24-c542e57014f1", Name: "Steelshaper's Gift", New: newSteelshapersGift,
	Programs:     []cardprogram.Program{{Reference: steelshapersGiftProgram, Resolve: resolveSteelshapersGift}},
	Verification: "mtg/rules/production_tutors_test.go: TestProductionTutorsSearchChoicesAndVisibility/Steelshaper's_Gift, TestProductionTutorsCastTimingAndPayment, TestProductionTutorsReplacementAndSearchControl, TestProductionTutorsCopyAndClone; TestProductionBatchLiveAndSimulator; CR 2026-09-25 701.23b/e",
}

// Fabricate reveals its qualified find before replacement-aware hand placement.
var Fabricate = cardset.Entry{
	OracleID: "422e1869-134f-463d-9fa1-86b66a998b3e", Name: "Fabricate", New: newFabricate,
	Programs:     []cardprogram.Program{{Reference: fabricateProgram, Resolve: resolveFabricate}},
	Verification: "mtg/rules/production_tutors_test.go: TestProductionTutorsSearchChoicesAndVisibility/Fabricate, TestProductionTutorsCastTimingAndPayment, TestProductionTutorsReplacementAndSearchControl, TestProductionTutorsCopyAndClone; TestProductionBatchLiveAndSimulator; CR 2026-09-25 701.23b/e",
}

// IdyllicTutor shares the qualified revealed-search placement operation.
var IdyllicTutor = cardset.Entry{
	OracleID: "57c9ff89-dd30-467b-bb3e-499eeea8cb94", Name: "Idyllic Tutor", New: newIdyllicTutor,
	Programs:     []cardprogram.Program{{Reference: idyllicTutorProgram, Resolve: resolveIdyllicTutor}},
	Verification: "mtg/rules/production_tutors_test.go: TestProductionTutorsSearchChoicesAndVisibility/Idyllic_Tutor, TestProductionTutorsCastTimingAndPayment, TestProductionTutorsReplacementAndSearchControl, TestProductionTutorsCopyAndClone; TestProductionBatchLiveAndSimulator; CR 2026-09-25 701.23b/e",
}

// SylvanScrying accepts any land card, not just basic lands.
var SylvanScrying = cardset.Entry{
	OracleID: "ee24bf27-484d-4e1c-998e-6a74e3d3f6c4", Name: "Sylvan Scrying", New: newSylvanScrying,
	Programs:     []cardprogram.Program{{Reference: sylvanScryingProgram, Resolve: resolveSylvanScrying}},
	Verification: "mtg/rules/production_tutors_test.go: TestProductionTutorsSearchChoicesAndVisibility/Sylvan_Scrying, TestProductionTutorsCastTimingAndPayment, TestProductionTutorsReplacementAndSearchControl, TestProductionTutorsCopyAndClone; TestProductionBatchLiveAndSimulator; CR 2026-09-25 701.23b/e",
}

func newEladamrisCall() *game.CardDef {
	const text = "Search your library for a creature card, reveal that card, put it into your hand, then shuffle."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green, color.White),
		CardFace: game.CardFace{
			Name: "Eladamri's Call", OracleText: text, ManaCost: opt.Val(cost.Mana{cost.G, cost.W}),
			Types: []types.Card{types.Instant}, Colors: []color.Color{color.Green, color.White},
			SpellAbility: opt.Val(game.Mode{Text: text, Program: eladamrisCallProgram}.Ability()),
		},
	}
}

func newSteelshapersGift() *game.CardDef {
	const text = "Search your library for an Equipment card, reveal that card, put it into your hand, then shuffle."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Steelshaper's Gift", OracleText: text, ManaCost: opt.Val(cost.Mana{cost.W}),
			Types: []types.Card{types.Sorcery}, Colors: []color.Color{color.White},
			SpellAbility: opt.Val(game.Mode{Text: text, Program: steelshapersGiftProgram}.Ability()),
		},
	}
}

func newFabricate() *game.CardDef {
	const text = "Search your library for an artifact card, reveal it, put it into your hand, then shuffle."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Fabricate", OracleText: text, ManaCost: opt.Val(cost.Mana{cost.O(2), cost.U}),
			Types: []types.Card{types.Sorcery}, Colors: []color.Color{color.Blue},
			SpellAbility: opt.Val(game.Mode{Text: text, Program: fabricateProgram}.Ability()),
		},
	}
}

func newIdyllicTutor() *game.CardDef {
	const text = "Search your library for an enchantment card, reveal it, put it into your hand, then shuffle."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Idyllic Tutor", OracleText: text, ManaCost: opt.Val(cost.Mana{cost.O(2), cost.W}),
			Types: []types.Card{types.Sorcery}, Colors: []color.Color{color.White},
			SpellAbility: opt.Val(game.Mode{Text: text, Program: idyllicTutorProgram}.Ability()),
		},
	}
}

func newSylvanScrying() *game.CardDef {
	const text = "Search your library for a land card, reveal it, put it into your hand, then shuffle."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Sylvan Scrying", OracleText: text, ManaCost: opt.Val(cost.Mana{cost.O(1), cost.G}),
			Types: []types.Card{types.Sorcery}, Colors: []color.Color{color.Green},
			SpellAbility: opt.Val(game.Mode{Text: text, Program: sylvanScryingProgram}.Ability()),
		},
	}
}

func resolveEladamrisCall(ctx cardprogram.Context) error {
	return resolveRevealedHandTutor(ctx, game.Selection{RequiredTypes: []types.Card{types.Creature}})
}

func resolveSteelshapersGift(ctx cardprogram.Context) error {
	return resolveRevealedHandTutor(ctx, game.Selection{SubtypesAny: []types.Sub{types.Equipment}})
}

func resolveFabricate(ctx cardprogram.Context) error {
	return resolveRevealedHandTutor(ctx, game.Selection{RequiredTypes: []types.Card{types.Artifact}})
}

func resolveIdyllicTutor(ctx cardprogram.Context) error {
	return resolveRevealedHandTutor(ctx, game.Selection{RequiredTypes: []types.Card{types.Enchantment}})
}

func resolveSylvanScrying(ctx cardprogram.Context) error {
	return resolveRevealedHandTutor(ctx, game.Selection{RequiredTypes: []types.Card{types.Land}})
}

func resolveRevealedHandTutor(ctx cardprogram.Context, filter game.Selection) error {
	found, err := ctx.Search(game.Search{
		Player: game.ControllerReference(), Amount: game.Fixed(1),
		Spec: game.SearchSpec{
			SourceZone: zone.Library, RevealOnly: true, Reveal: true, FailToFindPolicy: game.SearchMayFailToFind,
			Filter: filter,
		},
	})
	if err != nil {
		return err
	}
	if _, err := ctx.MoveCards(cardprogram.CardPlacement{Cards: found.Cards, Destination: zone.Hand}); err != nil {
		return err
	}
	_, err = ctx.Execute(game.ShuffleLibrary{Player: game.ControllerReference()})
	return err
}
