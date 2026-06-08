package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
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
	if !ok || permanent.Face != game.FaceBack || !permanent.Tapped || !permanentHasType(g, permanent, types.Land) {
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
	if len(g.Battlefield) != 1 || g.Battlefield[0].Face != game.FaceBack || !permanentHasType(g, g.Battlefield[0], types.Artifact) {
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

	resolveInstruction(engine, g, obj, game.Transform{TargetIndex: 0}, nil)

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

	resolveInstruction(engine, g, obj, game.Transform{TargetIndex: 0}, nil)

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
	g.Events = append(g.Events, game.Event{
		Kind:       game.EventSpellCast,
		Controller: game.Player1,
		CardTypes:  []types.Card{types.Instant},
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
	return &game.CardDef{CardFace: game.CardFace{Name: "Front Spell",

		Types: []types.Card{types.Sorcery}}, Layout: game.LayoutModalDFC,

		Back: opt.Val(game.CardFace{
			Name:     "Back Land",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Forest},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
		}),
	}
}

func modalDFCArtifactBack() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Creature Front",

		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1})}, Layout: game.LayoutModalDFC,

		Back: opt.Val(game.CardFace{Name: "Artifact Back", Types: []types.Card{types.Artifact}}),
	}
}

func transformCreature() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Small Front",

		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1})}, Layout: game.LayoutTransform,

		Back: opt.Val(game.CardFace{Name: "Large Back", Types: []types.Card{types.Creature}, Power: opt.Val(game.PT{Value: 4}), Toughness: opt.Val(game.PT{Value: 4})}),
	}
}

func transformCreatureWithBackTrigger() *game.CardDef {
	card := transformCreature()
	back := card.Back.Val
	back.TriggeredAbilities = []game.TriggeredAbility{
		{
			Trigger: game.TriggerCondition{
				Type:    game.TriggerWhenever,
				Pattern: game.TriggerPattern{Event: game.EventSpellCast, Controller: game.TriggerControllerYou},
			},
			Content: game.Mode{
				Sequence: []game.Instruction{{Primitive: game.GainLife{TargetIndex: game.TargetIndexController, Amount: game.Fixed(1)}}},
			}.Ability(),
		},
	}
	card.Back = opt.Val(back)
	return card
}
