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

var twincastProgram = game.ProgramReference{ID: "8f878efc-850f-43d2-a6fe-5ea8d1dd5afb/spell/0", Version: "1"}
var reverberateProgram = game.ProgramReference{ID: "a1f55890-31c5-4ed4-a2cd-7a4a9f05f8ca/spell/0", Version: "1"}
var dualcasterProgram = game.ProgramReference{ID: "8eb7c0a5-6190-40de-b473-2d1daa3bbe28/enter/0", Version: "1"}
var cacklingCounterpartProgram = game.ProgramReference{ID: "9e2adca5-f39c-4a09-bcce-8238ebac2c4a/spell/0", Version: "1"}
var fatedInfatuationProgram = game.ProgramReference{ID: "ae771e78-1d29-4b8c-8511-9da196a78dcf/spell/0", Version: "1"}

// Twincast uses Oracle printing revision f38eec33-a40e-4739-ad2c-2f57008cef4e.
var Twincast = cardset.Entry{
	OracleID: "8f878efc-850f-43d2-a6fe-5ea8d1dd5afb", Name: "Twincast", New: newTwincast,
	Programs:     []cardprogram.Program{{Reference: twincastProgram, Resolve: resolveTargetSpellCopy}},
	Verification: "mtg/rules/authored_copy_spells_test.go: TestAuthoredSpellCopiesChoicesAndCasting/Twincast; TestAuthoredSpellCopiesIllegalTarget/Twincast; TestAuthoredSpellCopiesUntargeted/Twincast",
}

// Reverberate uses Oracle printing revision 5996feb4-02ac-45e8-a7f2-966cf74391dc.
var Reverberate = cardset.Entry{
	OracleID: "a1f55890-31c5-4ed4-a2cd-7a4a9f05f8ca", Name: "Reverberate", New: newReverberate,
	Programs:     []cardprogram.Program{{Reference: reverberateProgram, Resolve: resolveTargetSpellCopy}},
	Verification: "mtg/rules/authored_copy_spells_test.go: TestAuthoredSpellCopiesChoicesAndCasting/Reverberate; TestAuthoredSpellCopiesIllegalTarget/Reverberate; TestAuthoredSpellCopiesUntargeted/Reverberate",
}

// DualcasterMage uses Oracle printing revision defcc4a3-40e0-4f5d-b23c-6cd6a614abc1.
var DualcasterMage = cardset.Entry{
	OracleID: "8eb7c0a5-6190-40de-b473-2d1daa3bbe28", Name: "Dualcaster Mage", New: newDualcasterMage,
	Programs:     []cardprogram.Program{{Reference: dualcasterProgram, Resolve: resolveTargetSpellCopy}},
	Verification: "mtg/rules/authored_copy_spells_test.go: TestAuthoredDualcasterMageFlashAndSourceDeparture; TestAuthoredSpellCopiesChoicesAndCasting/Dualcaster_Mage; TestAuthoredSpellCopiesUntargeted/Dualcaster_Mage",
}

// CacklingCounterpart uses Oracle printing revision a865fdc5-955a-4c4a-90c0-342c15fd7bbc.
var CacklingCounterpart = cardset.Entry{
	OracleID: "9e2adca5-f39c-4a09-bcce-8238ebac2c4a", Name: "Cackling Counterpart", New: newCacklingCounterpart,
	Programs:     []cardprogram.Program{{Reference: cacklingCounterpartProgram, Resolve: resolveCreatureCopyToken}},
	Verification: "mtg/rules/authored_copy_spells_test.go: TestAuthoredCreatureCopyTokens/Cackling_Counterpart; TestAuthoredCreatureCopyTargetLegality/Cackling_Counterpart; TestAuthoredCacklingCounterpartFlashback; TestAuthoredCreatureCopyRetargetedSpellAndClone/Cackling_Counterpart",
}

// FatedInfatuation uses Oracle printing revision 9e5ba99d-d95d-4a3b-b6b6-a5e2589d538e.
var FatedInfatuation = cardset.Entry{
	OracleID: "ae771e78-1d29-4b8c-8511-9da196a78dcf", Name: "Fated Infatuation", New: newFatedInfatuation,
	Programs:     []cardprogram.Program{{Reference: fatedInfatuationProgram, Resolve: resolveFatedInfatuation}},
	Verification: "mtg/rules/authored_copy_spells_test.go: TestAuthoredCreatureCopyTokens/Fated_Infatuation; TestAuthoredCreatureCopyTargetLegality/Fated_Infatuation; TestAuthoredFatedInfatuationScry; TestAuthoredCreatureCopyRetargetedSpellAndClone/Fated_Infatuation",
}

const targetSpellCopyText = "Copy target instant or sorcery spell. You may choose new targets for the copy."
const creatureCopyText = "Create a token that's a copy of target creature you control."

func instantSorcerySpellTarget() game.TargetSpec {
	return game.TargetSpec{
		MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowStackObject,
		Predicate: game.TargetPredicate{
			StackObjectKinds:  []game.StackObjectKind{game.StackSpell},
			SpellCardTypesAny: []types.Card{types.Instant, types.Sorcery},
		},
	}
}

func controlledCreatureTarget() game.TargetSpec {
	return game.TargetSpec{
		MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent,
		Selection: opt.Val(game.Selection{
			RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou,
		}),
	}
}

func newTwincast() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Twincast", ManaCost: opt.Val(cost.Mana{cost.U, cost.U}),
			Types: []types.Card{types.Instant}, Colors: []color.Color{color.Blue}, OracleText: targetSpellCopyText,
			SpellAbility: opt.Val(game.Mode{
				Text: targetSpellCopyText, Targets: []game.TargetSpec{instantSorcerySpellTarget()}, Program: twincastProgram,
			}.Ability()),
		},
	}
}

func newReverberate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Reverberate", ManaCost: opt.Val(cost.Mana{cost.R, cost.R}),
			Types: []types.Card{types.Instant}, Colors: []color.Color{color.Red}, OracleText: targetSpellCopyText,
			SpellAbility: opt.Val(game.Mode{
				Text: targetSpellCopyText, Targets: []game.TargetSpec{instantSorcerySpellTarget()}, Program: reverberateProgram,
			}.Ability()),
		},
	}
}

func newDualcasterMage() *game.CardDef {
	const text = "Flash\nWhen this creature enters, copy target instant or sorcery spell. You may choose new targets for the copy."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Dualcaster Mage", ManaCost: opt.Val(cost.Mana{cost.O(1), cost.R, cost.R}),
			Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Human, types.Wizard},
			Colors: []color.Color{color.Red}, Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 2}),
			OracleText: text, StaticAbilities: []game.StaticAbility{game.FlashStaticBody},
			TriggeredAbilities: []game.TriggeredAbility{{
				Trigger: game.TriggerCondition{Pattern: game.TriggerPattern{
					Event: game.EventPermanentEnteredBattlefield, Source: game.TriggerSourceSelf,
				}},
				Content: game.Mode{
					Text: targetSpellCopyText, Targets: []game.TargetSpec{instantSorcerySpellTarget()}, Program: dualcasterProgram,
				}.Ability(),
			}},
		},
	}
}

func newCacklingCounterpart() *game.CardDef {
	const text = creatureCopyText + "\nFlashback {5}{U}{U} (You may cast this card from your graveyard for its flashback cost. Then exile it.)"
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Cackling Counterpart", ManaCost: opt.Val(cost.Mana{cost.O(1), cost.U, cost.U}),
			Types: []types.Card{types.Instant}, Colors: []color.Color{color.Blue}, OracleText: text,
			StaticAbilities: []game.StaticAbility{{
				Text:             "Flashback {5}{U}{U}",
				KeywordAbilities: []game.KeywordAbility{game.FlashbackKeyword{Cost: cost.Mana{cost.O(5), cost.U, cost.U}}},
			}},
			SpellAbility: opt.Val(game.Mode{
				Text: creatureCopyText, Targets: []game.TargetSpec{controlledCreatureTarget()}, Program: cacklingCounterpartProgram,
			}.Ability()),
		},
	}
}

func newFatedInfatuation() *game.CardDef {
	const text = creatureCopyText + " If it's your turn, scry 2."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Fated Infatuation", ManaCost: opt.Val(cost.Mana{cost.U, cost.U, cost.U}),
			Types: []types.Card{types.Instant}, Colors: []color.Color{color.Blue}, OracleText: text,
			SpellAbility: opt.Val(game.Mode{
				Text: text, Targets: []game.TargetSpec{controlledCreatureTarget()}, Program: fatedInfatuationProgram,
			}.Ability()),
		},
	}
}

func resolveTargetSpellCopy(ctx cardprogram.Context) error {
	_, err := ctx.Copy(game.CopyStackObject{
		Object: game.TargetStackObjectReference(0), MayChooseNewTargets: true,
	}, ctx.View().Source().Controller)
	return err
}

func resolveCreatureCopyToken(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.CreateToken{
		Amount: game.Fixed(1),
		Source: game.TokenCopyOf(game.TokenCopySpec{
			Source: game.TokenCopySourceObject, Object: game.TargetPermanentReference(0),
		}),
	})
	return err
}

func resolveFatedInfatuation(ctx cardprogram.Context) error {
	if err := resolveCreatureCopyToken(ctx); err != nil {
		return err
	}
	activePlayer, err := ctx.View().ActivePlayer()
	if err != nil {
		return err
	}
	if activePlayer == ctx.View().Source().Controller {
		_, err = ctx.Execute(game.Scry{Amount: game.Fixed(2), Player: game.ControllerReference()})
	}
	return err
}
