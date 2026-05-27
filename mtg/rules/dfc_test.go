package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/opt"
)

func TestModalDFCLandFaceCanBePlayed(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, modalDFCSpellLand())
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)
	if !actionsContain(legal, action.PlayLandFace(cardID, game.FaceBack)) {
		t.Fatalf("legal actions = %+v, want PlayLandFace(back)", legal)
	}
	if !engine.applyAction(g, game.Player1, action.PlayLandFace(cardID, game.FaceBack)) {
		t.Fatal("applyAction PlayLandFace(back) = false, want true")
	}
	permanent, ok := g.PermanentByID(g.Battlefield[0].ObjectID)
	if !ok || permanent.Face != game.FaceBack || !permanent.Tapped || !permanentHasType(g, permanent, game.TypeLand) {
		t.Fatalf("permanent = %+v, want tapped back-face land", permanent)
	}
}

func TestModalDFCBackPermanentFaceCanBeCastAndResolve(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, modalDFCArtifactBack())
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)
	if !actionsContain(legal, action.CastSpellFace(cardID, game.FaceBack, nil, 0, nil)) {
		t.Fatalf("legal actions = %+v, want CastSpellFace(back)", legal)
	}
	if !engine.applyAction(g, game.Player1, action.CastSpellFace(cardID, game.FaceBack, nil, 0, nil)) {
		t.Fatal("applyAction CastSpellFace(back) = false, want true")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.Face != game.FaceBack {
		t.Fatalf("stack object = %+v, want back face", obj)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if len(g.Battlefield) != 1 || g.Battlefield[0].Face != game.FaceBack || !permanentHasType(g, g.Battlefield[0], game.TypeArtifact) {
		t.Fatalf("battlefield = %+v, want one back-face artifact", g.Battlefield)
	}
}

func TestTransformChangesEffectiveFaceCharacteristics(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardInstance(g, game.Player1, transformCreature())
	permanent := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          game.Player1,
		Controller:     game.Player1,
		Face:           game.FaceFront,
	}
	g.Battlefield = append(g.Battlefield, permanent)
	obj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{game.PermanentTarget(permanent.ObjectID)}}

	engine.resolveEffect(g, obj, game.Effect{Type: game.EffectTransform, TargetIndex: 0}, nil)

	if permanent.Face != game.FaceBack || !permanent.Transformed {
		t.Fatalf("permanent face/transformed = %v/%v, want back/true", permanent.Face, permanent.Transformed)
	}
	if got := effectivePower(g, permanent); got != 4 {
		t.Fatalf("effective power = %d, want 4 from back face", got)
	}
}

func TestTransformDoesNothingToModalDFC(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardInstance(g, game.Player1, modalDFCArtifactBack())
	permanent := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          game.Player1,
		Controller:     game.Player1,
		Face:           game.FaceBack,
	}
	g.Battlefield = append(g.Battlefield, permanent)
	obj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{game.PermanentTarget(permanent.ObjectID)}}

	engine.resolveEffect(g, obj, game.Effect{Type: game.EffectTransform, TargetIndex: 0}, nil)

	if permanent.Face != game.FaceBack || permanent.Transformed {
		t.Fatalf("modal DFC face/transformed = %v/%v, want back/false", permanent.Face, permanent.Transformed)
	}
}

func TestBackFaceTriggeredAbilityResolvesUsingCapturedFace(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardInstance(g, game.Player1, transformCreatureWithBackTrigger())
	permanent := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          game.Player1,
		Controller:     game.Player1,
		Face:           game.FaceBack,
		Transformed:    true,
	}
	g.Battlefield = append(g.Battlefield, permanent)
	g.Events = append(g.Events, game.GameEvent{
		Kind:       game.EventSpellCast,
		Controller: game.Player1,
		CardTypes:  []game.CardType{game.TypeInstant},
	})

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("putTriggeredAbilitiesOnStack() = false, want back-face trigger")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.Face != game.FaceBack {
		t.Fatalf("trigger stack object = %+v, want back face", obj)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Life; got != 41 {
		t.Fatalf("life = %d, want 41 from back-face trigger", got)
	}
}

func addCardInstance(g *game.Game, owner game.PlayerID, def *game.CardDef) game.ObjectID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Def: def, Owner: owner}
	return cardID
}

func modalDFCSpellLand() *game.CardDef {
	return &game.CardDef{
		Name:   "Front Spell",
		Layout: game.LayoutModalDFC,
		Types:  []game.CardType{game.TypeSorcery},
		Faces: []game.CardFace{
			{Name: "Front Spell", Types: []game.CardType{game.TypeSorcery}},
			{Name: "Back Land", Types: []game.CardType{game.TypeLand}, Subtypes: []string{"Forest"}, EntersTapped: true},
		},
	}
}

func modalDFCArtifactBack() *game.CardDef {
	return &game.CardDef{
		Name:   "Creature Front",
		Layout: game.LayoutModalDFC,
		Faces: []game.CardFace{
			{Name: "Creature Front", Types: []game.CardType{game.TypeCreature}, Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1})},
			{Name: "Artifact Back", Types: []game.CardType{game.TypeArtifact}},
		},
	}
}

func transformCreature() *game.CardDef {
	return &game.CardDef{
		Name:   "Small Front",
		Layout: game.LayoutTransform,
		Faces: []game.CardFace{
			{Name: "Small Front", Types: []game.CardType{game.TypeCreature}, Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1})},
			{Name: "Large Back", Types: []game.CardType{game.TypeCreature}, Power: opt.Val(game.PT{Value: 4}), Toughness: opt.Val(game.PT{Value: 4})},
		},
	}
}

func transformCreatureWithBackTrigger() *game.CardDef {
	card := transformCreature()
	card.Faces[1].Abilities = []game.AbilityDef{
		{
			Kind: game.TriggeredAbility,
			Trigger: opt.Val(game.TriggerCondition{
				Type:    game.TriggerWhenever,
				Pattern: game.TriggerPattern{Event: game.EventSpellCast, Controller: game.TriggerControllerYou},
			}),
			Effects: []game.Effect{{Type: game.EffectGainLife, TargetIndex: -1, Amount: 1}},
		},
	}
	return card
}
