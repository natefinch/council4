package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

func TestActivatedAbilityCollectEvidenceExilesSelectedGraveyardCardsAsCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: []cost.Additional{{
			Kind:   cost.AdditionalCollectEvidence,
			Text:   "Collect evidence 4",
			Amount: 4,
			Source: zone.Graveyard,
		}},
		Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.GainLife{
				Amount: game.Fixed(1),
				Player: game.ControllerReference(),
			}}},
		}.Ability(),
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)
	first := addCardToGraveyard(g, game.Player1, evidenceCard("Evidence Two", 2))
	second := addCardToGraveyard(g, game.Player1, evidenceCard("Evidence Three", 3))

	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("collect-evidence ability was not legal with enough graveyard mana value")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(collect-evidence ability) = false, want true")
	}
	if g.Players[game.Player1].Graveyard.Contains(first) ||
		g.Players[game.Player1].Graveyard.Contains(second) ||
		!g.Players[game.Player1].Exile.Contains(first) ||
		!g.Players[game.Player1].Exile.Contains(second) {
		t.Fatal("evidence cards did not move from graveyard to exile")
	}
}

func TestActivatedAbilityCollectEvidenceRequiresEnoughManaValue(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: []cost.Additional{{
			Kind:   cost.AdditionalCollectEvidence,
			Text:   "Collect evidence 4",
			Amount: 4,
			Source: zone.Graveyard,
		}},
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.GainLife{
			Amount: game.Fixed(1),
			Player: game.ControllerReference(),
		}}}}.Ability(),
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)
	cardID := addCardToGraveyard(g, game.Player1, evidenceCard("Evidence Two", 2))

	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("collect-evidence ability was legal with insufficient graveyard mana value")
	}
	if engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(insufficient collect evidence) = true, want false")
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) || g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("failed collect-evidence payment mutated zones")
	}
}

func TestCollectEvidenceRejectsStalePreferenceWithoutMutation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Evidence Source"}})
	graveyardCard := addCardToGraveyard(g, game.Player1, evidenceCard("Evidence Four", 4))
	handCard := addCardToHand(g, game.Player1, evidenceCard("Stale Evidence", 4))

	ok := paymentOrch.payAbilityCosts(g, payment.AbilityRequest{
		PlayerID: game.Player1,
		Source:   source,
		AdditionalCosts: []cost.Additional{{
			Kind:   cost.AdditionalCollectEvidence,
			Text:   "Collect evidence 4",
			Amount: 4,
			Source: zone.Graveyard,
		}},
		Prefs: &payment.Preferences{EvidenceChoices: []id.ID{handCard}},
	})
	if ok {
		t.Fatal("stale collect-evidence preference paid successfully")
	}
	if !g.Players[game.Player1].Graveyard.Contains(graveyardCard) ||
		g.Players[game.Player1].Exile.Contains(graveyardCard) ||
		!g.Players[game.Player1].Hand.Contains(handCard) {
		t.Fatal("stale collect-evidence preference mutated zones")
	}
}

func TestCollectEvidenceAndExileCostCannotReuseGraveyardCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Evidence Source"}})
	graveyardCard := addCardToGraveyard(g, game.Player1, evidenceCard("Evidence Four", 4))

	ok := paymentOrch.payAbilityCosts(g, payment.AbilityRequest{
		PlayerID: game.Player1,
		Source:   source,
		AdditionalCosts: []cost.Additional{
			{
				Kind:   cost.AdditionalCollectEvidence,
				Text:   "Collect evidence 4",
				Amount: 4,
				Source: zone.Graveyard,
			},
			{
				Kind:   cost.AdditionalExile,
				Text:   "Exile a card from your graveyard",
				Amount: 1,
				Source: zone.Graveyard,
			},
		},
	})
	if ok {
		t.Fatal("collect-evidence and exile costs reused the same graveyard card")
	}
	if !g.Players[game.Player1].Graveyard.Contains(graveyardCard) ||
		g.Players[game.Player1].Exile.Contains(graveyardCard) {
		t.Fatal("failed combined collect-evidence/exile payment mutated zones")
	}
}

func TestActivatedAbilityCollectEvidenceAndExileCostChoosesDistinctGraveyardCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: []cost.Additional{
			{
				Kind:   cost.AdditionalCollectEvidence,
				Text:   "Collect evidence 4",
				Amount: 4,
				Source: zone.Graveyard,
			},
			{
				Kind:   cost.AdditionalExile,
				Text:   "Exile a card from your graveyard",
				Amount: 1,
				Source: zone.Graveyard,
			},
		},
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.GainLife{
			Amount: game.Fixed(1),
			Player: game.ControllerReference(),
		}}}}.Ability(),
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)
	evidence := addCardToGraveyard(g, game.Player1, evidenceCard("Evidence Four", 4))
	firstFodder := addCardToGraveyard(g, game.Player1, evidenceCard("First Fodder", 1))
	secondFodder := addCardToGraveyard(g, game.Player1, evidenceCard("Second Fodder", 1))

	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("combined collect-evidence/exile ability was not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(combined collect-evidence/exile ability) = false, want true")
	}
	exiledCount := 0
	for _, cardID := range []id.ID{evidence, firstFodder, secondFodder} {
		if g.Players[game.Player1].Exile.Contains(cardID) {
			exiledCount++
		}
	}
	if exiledCount < 2 || !g.Players[game.Player1].Exile.Contains(evidence) {
		t.Fatal("combined collect-evidence/exile payment did not exile distinct graveyard cards")
	}
}

func TestActivatedAbilityCollectEvidencePreservesCardsForLaterEvidenceCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: []cost.Additional{
			{
				Kind:   cost.AdditionalCollectEvidence,
				Text:   "Collect evidence 6",
				Amount: 6,
				Source: zone.Graveyard,
			},
			{
				Kind:   cost.AdditionalCollectEvidence,
				Text:   "Collect evidence 10",
				Amount: 10,
				Source: zone.Graveyard,
			},
		},
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.GainLife{
			Amount: game.Fixed(1),
			Player: game.ControllerReference(),
		}}}}.Ability(),
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)
	six := addCardToGraveyard(g, game.Player1, evidenceCard("Evidence Six", 6))
	ten := addCardToGraveyard(g, game.Player1, evidenceCard("Evidence Ten", 10))

	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("combined collect-evidence ability was not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(combined collect-evidence ability) = false, want true")
	}
	if !g.Players[game.Player1].Exile.Contains(six) ||
		!g.Players[game.Player1].Exile.Contains(ten) {
		t.Fatal("combined collect-evidence payment did not preserve cards for the later threshold")
	}
}

func TestActivatedAbilityCollectEvidencePreservesCreatureForLaterExileCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: []cost.Additional{
			{
				Kind:   cost.AdditionalCollectEvidence,
				Text:   "Collect evidence 4",
				Amount: 4,
				Source: zone.Graveyard,
			},
			{
				Kind:          cost.AdditionalExile,
				Text:          "Exile a creature card from your graveyard",
				Amount:        1,
				Source:        zone.Graveyard,
				MatchCardType: true,
				CardType:      types.Creature,
			},
		},
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.GainLife{
			Amount: game.Fixed(1),
			Player: game.ControllerReference(),
		}}}}.Ability(),
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)
	nonCreature := addCardToGraveyard(g, game.Player1, evidenceCard("Noncreature Evidence", 4))
	creature := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Creature Evidence",
		ManaCost: opt.Val(cost.Mana{cost.O(4)}),
		Types:    []types.Card{types.Creature},
	}})

	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("collect-evidence plus typed-exile ability was not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(collect-evidence plus typed-exile ability) = false, want true")
	}
	if !g.Players[game.Player1].Exile.Contains(nonCreature) ||
		!g.Players[game.Player1].Exile.Contains(creature) {
		t.Fatal("collect-evidence payment did not preserve the creature card for typed exile")
	}
}

func TestCollectEvidenceRejectsUnsupportedVariableManaValue(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: []cost.Additional{{
			Kind:   cost.AdditionalCollectEvidence,
			Text:   "Collect evidence 1",
			Amount: 1,
			Source: zone.Graveyard,
		}},
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.GainLife{
			Amount: game.Fixed(1),
			Player: game.ControllerReference(),
		}}}}.Ability(),
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)
	cardID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Variable Evidence",
		ManaCost: opt.Val(cost.Mana{cost.X}),
		Types:    []types.Card{types.Sorcery},
	}})

	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("collect-evidence ability was legal with only variable mana value evidence")
	}
	if g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("variable evidence card moved to exile")
	}
}

func TestGraveyardActivatedAbilityChecksActivationCondition(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Conditional Escape",
		Types: []types.Card{types.Creature},
		ActivatedAbilities: []game.ActivatedAbility{{
			ZoneOfFunction: zone.Graveyard,
			ActivationCondition: opt.Val(game.Condition{
				ControllerLifeAtLeast: 10,
			}),
			Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.GainLife{
				Amount: game.Fixed(1),
				Player: game.ControllerReference(),
			}}}}.Ability(),
		}},
	}})
	g.Turn.PriorityPlayer = game.Player1
	act := action.ActivateAbility(sourceID, 0, nil, 0)

	g.Players[game.Player1].Life = 9
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("graveyard ability was legal while its activation condition was false")
	}
	if engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(graveyard ability with false activation condition) = true, want false")
	}

	g.Players[game.Player1].Life = 10
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("graveyard ability was not legal while its activation condition was true")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(graveyard ability with true activation condition) = false, want true")
	}
}

func TestGraveyardCollectEvidencePreservesSourceForExileSourceCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Evidence Escape",
		ManaCost: opt.Val(cost.Mana{cost.O(4)}),
		Types:    []types.Card{types.Creature},
		ActivatedAbilities: []game.ActivatedAbility{{
			ZoneOfFunction: zone.Graveyard,
			AdditionalCosts: []cost.Additional{
				{
					Kind:   cost.AdditionalCollectEvidence,
					Text:   "Collect evidence 4",
					Amount: 4,
					Source: zone.Graveyard,
				},
				{
					Kind:   cost.AdditionalExileSource,
					Text:   "Exile this card from your graveyard",
					Amount: 1,
					Source: zone.Graveyard,
				},
			},
			Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.GainLife{
				Amount: game.Fixed(1),
				Player: game.ControllerReference(),
			}}}}.Ability(),
		}},
	}})
	otherEvidence := addCardToGraveyard(g, game.Player1, evidenceCard("Other Evidence", 4))
	g.Turn.PriorityPlayer = game.Player1
	act := action.ActivateAbility(sourceID, 0, nil, 0)

	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("graveyard collect-evidence/exile-source ability was not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(graveyard collect-evidence/exile-source ability) = false, want true")
	}
	if !g.Players[game.Player1].Exile.Contains(sourceID) ||
		!g.Players[game.Player1].Exile.Contains(otherEvidence) {
		t.Fatal("graveyard collect-evidence payment did not preserve source for exile-source cost")
	}
}

func TestGraveyardExileSourcePreservesOtherCardForLaterCollectEvidenceCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Evidence Escape",
		ManaCost: opt.Val(cost.Mana{cost.O(4)}),
		Types:    []types.Card{types.Creature},
		ActivatedAbilities: []game.ActivatedAbility{{
			ZoneOfFunction: zone.Graveyard,
			AdditionalCosts: []cost.Additional{
				{
					Kind:   cost.AdditionalExileSource,
					Text:   "Exile this card from your graveyard",
					Amount: 1,
					Source: zone.Graveyard,
				},
				{
					Kind:   cost.AdditionalCollectEvidence,
					Text:   "Collect evidence 4",
					Amount: 4,
					Source: zone.Graveyard,
				},
			},
			Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.GainLife{
				Amount: game.Fixed(1),
				Player: game.ControllerReference(),
			}}}}.Ability(),
		}},
	}})
	otherEvidence := addCardToGraveyard(g, game.Player1, evidenceCard("Other Evidence", 4))
	g.Turn.PriorityPlayer = game.Player1
	act := action.ActivateAbility(sourceID, 0, nil, 0)

	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("graveyard exile-source/collect-evidence ability was not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(graveyard exile-source/collect-evidence ability) = false, want true")
	}
	if !g.Players[game.Player1].Exile.Contains(sourceID) ||
		!g.Players[game.Player1].Exile.Contains(otherEvidence) {
		t.Fatal("graveyard exile-source payment did not preserve other card for collect-evidence cost")
	}
}
