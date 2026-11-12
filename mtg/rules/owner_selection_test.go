package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Ownership selection (CR 108.3, 108.4): "you own" compares a permanent's or
// card's owner, never its controller. A card outside the battlefield has no
// controller (CR 108.4a); the selection reads its owner directly.

func ownedTargetArtifact(selection game.Selection, constraint string, destination zone.Type) *game.CardDef {
	var primitive game.Primitive = game.Destroy{Object: game.TargetPermanentReference(0)}
	if destination == zone.Exile {
		primitive = game.MovePermanent{Object: game.TargetPermanentReference(0), Destination: zone.Exile}
	}
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Ownership Rod",
		Types: []types.Card{types.Artifact},
		ActivatedAbilities: []game.ActivatedAbility{{
			Text:            "{T}: " + constraint,
			AdditionalCosts: cost.Tap,
			ZoneOfFunction:  zone.Battlefield,
			Content: game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: constraint,
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(selection),
				}},
				Sequence: []game.Instruction{{Primitive: primitive}},
			}.Ability(),
		}},
	}}
}

func addOwnedByPermanent(g *game.Game, owner, controller game.PlayerID) *game.Permanent {
	permanent := addCombatPermanent(g, owner, bearDef())
	permanent.Controller = controller
	return permanent
}

func activatedPermanentTargets(g *game.Game, e *Engine, player game.PlayerID, sourceID id.ID) []id.ID {
	var targets []id.ID
	for _, act := range e.legalActions(g, player) {
		payload, ok := act.ActivateAbilityPayload()
		if !ok || payload.SourceID != sourceID || len(payload.Targets) != 1 {
			continue
		}
		targets = append(targets, payload.Targets[0].PermanentID)
	}
	return targets
}

func TestTargetPermanentYouOwnMatchesOwnerNotController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	g.Turn.Phase, g.Turn.Step = game.PhasePrecombatMain, game.StepNone
	rod := addCombatPermanent(g, game.Player1, ownedTargetArtifact(
		game.Selection{Owner: game.OwnerYou}, "Destroy target permanent you own.", zone.Graveyard))
	stolenFromMe := addOwnedByPermanent(g, game.Player1, game.Player2)
	stolenByMe := addOwnedByPermanent(g, game.Player2, game.Player1)
	mine := addOwnedByPermanent(g, game.Player1, game.Player1)
	theirs := addOwnedByPermanent(g, game.Player2, game.Player2)

	targets := activatedPermanentTargets(g, e, game.Player1, rod.ObjectID)
	for _, want := range []id.ID{stolenFromMe.ObjectID, mine.ObjectID, rod.ObjectID} {
		if !slices.Contains(targets, want) {
			t.Fatalf("targets = %v, want owned permanent %v", targets, want)
		}
	}
	for _, unwanted := range []id.ID{stolenByMe.ObjectID, theirs.ObjectID} {
		if slices.Contains(targets, unwanted) {
			t.Fatalf("targets = %v, must exclude permanent %v owned by an opponent", targets, unwanted)
		}
	}
	if e.applyAction(g, game.Player1, action.ActivateAbility(rod.ObjectID, 0, []game.Target{game.PermanentTarget(stolenByMe.ObjectID)}, 0)) {
		t.Fatal("activation targeting a permanent you control but don't own was accepted")
	}

	stolenCard := stolenFromMe.CardInstanceID
	if !e.applyAction(g, game.Player1, action.ActivateAbility(rod.ObjectID, 0, []game.Target{game.PermanentTarget(stolenFromMe.ObjectID)}, 0)) {
		t.Fatal("activation targeting a permanent you own but don't control was rejected")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	if permanentForCard(g, stolenCard) != nil {
		t.Fatal("owned permanent controlled by an opponent was not destroyed")
	}
	if !g.Players[game.Player1].Graveyard.Contains(stolenCard) {
		t.Fatal("destroyed permanent did not go to its owner's graveyard")
	}
}

func TestTargetPermanentYouOwnOrControlUsesSelectionAlternatives(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	g.Turn.Phase, g.Turn.Step = game.PhasePrecombatMain, game.StepNone
	rod := addCombatPermanent(g, game.Player1, ownedTargetArtifact(game.Selection{AnyOf: []game.Selection{
		{Owner: game.OwnerYou},
		{Controller: game.ControllerYou},
	}}, "Exile target permanent you own or control.", zone.Exile))
	stolenFromMe := addOwnedByPermanent(g, game.Player1, game.Player2)
	stolenByMe := addOwnedByPermanent(g, game.Player2, game.Player1)
	theirs := addOwnedByPermanent(g, game.Player2, game.Player2)

	targets := activatedPermanentTargets(g, e, game.Player1, rod.ObjectID)
	for _, want := range []id.ID{stolenFromMe.ObjectID, stolenByMe.ObjectID} {
		if !slices.Contains(targets, want) {
			t.Fatalf("targets = %v, want permanent %v you own or control", targets, want)
		}
	}
	if slices.Contains(targets, theirs.ObjectID) {
		t.Fatalf("targets = %v, must exclude a permanent an opponent owns and controls", targets)
	}

	stolenCard := stolenByMe.CardInstanceID
	if !e.applyAction(g, game.Player1, action.ActivateAbility(rod.ObjectID, 0, []game.Target{game.PermanentTarget(stolenByMe.ObjectID)}, 0)) {
		t.Fatal("activation targeting a permanent you control but don't own was rejected")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player2].Exile.Contains(stolenCard) {
		t.Fatal("exiled permanent did not go to its owner's exile")
	}
}

func TestDestroyAllCreaturesYouOwnOrControlGroupReadsAlternatives(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	g.Turn.Phase, g.Turn.Step = game.PhasePrecombatMain, game.StepNone
	sweeper := &game.CardDef{CardFace: game.CardFace{
		Name:     "Ownership Sweep",
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(cost.Mana{cost.O(1)}),
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{Primitive: game.Destroy{
			Group: game.BattlefieldGroup(game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				AnyOf:         []game.Selection{{Owner: game.OwnerYou}, {Controller: game.ControllerYou}},
			}),
		}}}}.Ability()),
	}}
	cardID := addCardToHand(g, game.Player1, sweeper)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	stolenFromMe := addOwnedByPermanent(g, game.Player1, game.Player2)
	stolenByMe := addOwnedByPermanent(g, game.Player2, game.Player1)
	theirs := addOwnedByPermanent(g, game.Player2, game.Player2)

	if !e.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("cast was rejected")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	if permanentForCard(g, stolenFromMe.CardInstanceID) != nil || permanentForCard(g, stolenByMe.CardInstanceID) != nil {
		t.Fatal("creatures you own or control survived")
	}
	if permanentForCard(g, theirs.CardInstanceID) == nil {
		t.Fatal("creature an opponent owns and controls was destroyed")
	}
}

func ownedExileReturnSpell(owner game.OwnerRelation) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Ownership Recall",
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(cost.Mana{cost.O(1)}),
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Constraint: "target exiled card with flashback",
				Allow:      game.TargetAllowCard,
				TargetZone: zone.Exile,
				Selection:  opt.Val(game.Selection{Owner: owner, Keyword: game.Flashback}),
			}},
			Sequence: []game.Instruction{{Primitive: game.MoveCard{
				Card:        game.CardReference{Kind: game.CardReferenceTarget},
				FromZone:    zone.Exile,
				Destination: zone.Hand,
			}}},
		}.Ability()),
	}}
}

func castCardTargets(g *game.Game, e *Engine, player game.PlayerID, cardID id.ID) []game.Target {
	var targets []game.Target
	for _, act := range e.legalActions(g, player) {
		payload, ok := act.CastSpellPayload()
		if ok && payload.CardID == cardID && len(payload.Targets) == 1 {
			targets = append(targets, payload.Targets[0])
		}
	}
	return targets
}

func TestTargetExiledCardYouOwnMatchesCardOwner(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	g.Turn.Phase, g.Turn.Step = game.PhasePrecombatMain, game.StepNone
	spellID := addCardToHand(g, game.Player1, ownedExileReturnSpell(game.OwnerYou))
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	mine := addCardToExile(g, game.Player1, flashbackSpell())
	theirs := addCardToExile(g, game.Player2, flashbackSpell())
	addCardToExile(g, game.Player1, bearDef())

	targets := castCardTargets(g, e, game.Player1, spellID)
	if len(targets) != 1 || targets[0].CardID != mine {
		t.Fatalf("targets = %v, want only your exiled flashback card %v", targets, mine)
	}
	if !e.applyAction(g, game.Player1, action.CastSpell(spellID, targets, 0, nil)) {
		t.Fatal("cast was rejected")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Hand.Contains(mine) {
		t.Fatal("exiled card you own did not return to your hand")
	}
	if !g.Players[game.Player2].Exile.Contains(theirs) {
		t.Fatal("opponent's exiled card moved")
	}
}

func TestTargetExiledCardOpponentOwnsMatchesCardOwner(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	g.Turn.Phase, g.Turn.Step = game.PhasePrecombatMain, game.StepNone
	spellID := addCardToHand(g, game.Player1, ownedExileReturnSpell(game.OwnerOpponent))
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	addCardToExile(g, game.Player1, flashbackSpell())
	theirs := addCardToExile(g, game.Player2, flashbackSpell())

	targets := castCardTargets(g, e, game.Player1, spellID)
	if len(targets) != 1 || targets[0].CardID != theirs {
		t.Fatalf("targets = %v, want only the opponent-owned exiled card %v", targets, theirs)
	}
}
