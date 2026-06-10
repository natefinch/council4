package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestPreparedPermanentCastsSpellCopyOnce(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, prepareCard())
	for range 4 {
		addBasicLandPermanent(g, game.Player1, types.Plains)
	}
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("casting prepare creature failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	permanent := permanentForCard(g, cardID)
	if permanent == nil || !permanent.Prepared {
		t.Fatalf("resolved permanent = %+v, want prepared", permanent)
	}
	preparedCast := action.CastSpellFaceFromZone(cardID, zone.Battlefield, game.FaceAlternate, nil, 0, nil)
	if !slices.ContainsFunc(engine.legalActions(g, game.Player1), func(got action.Action) bool {
		return actionsEqual(got, preparedCast)
	}) {
		t.Fatal("legal actions do not include casting the prepared spell copy")
	}
	if !engine.applyAction(g, game.Player1, preparedCast) {
		t.Fatal("casting prepared spell copy failed")
	}
	if permanent.Prepared {
		t.Fatal("permanent remained prepared after casting its spell copy")
	}
	if permanentForCard(g, cardID) != permanent {
		t.Fatal("prepared permanent left the battlefield")
	}
	stackObject, ok := g.Stack.Peek()
	if !ok || !stackObject.Copy || stackObject.Face != game.FaceAlternate || stackObject.SourceZone != zone.Battlefield {
		t.Fatalf("stack object = %+v, want alternate-face battlefield copy", stackObject)
	}
	if slices.ContainsFunc(engine.legalActions(g, game.Player1), func(got action.Action) bool {
		return actionsEqual(got, preparedCast)
	}) {
		t.Fatal("unprepared permanent can cast its spell copy again")
	}

	life := g.Players[game.Player1].Life
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Life != life+2 {
		t.Fatalf("life = %d, want %d", g.Players[game.Player1].Life, life+2)
	}
	if permanentForCard(g, cardID) != permanent {
		t.Fatal("prepared permanent left the battlefield when its spell copy resolved")
	}
}

func TestPrepareSpellFaceCannotBeCastFromHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, prepareCard())
	addBasicLandPermanent(g, game.Player1, types.Plains)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if engine.applyAction(g, game.Player1, action.CastSpellFace(cardID, game.FaceAlternate, nil, 0, nil)) {
		t.Fatal("prepare spell face was cast directly from hand")
	}
}

func TestTokenCopyOfPrepareCreatureCastsItsSpellCopy(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	token, ok := createTokenPermanent(g, game.Player1, copyCardDef(prepareCard()))
	if !ok || !token.Prepared {
		t.Fatalf("token = %+v, want prepared token copy", token)
	}
	addBasicLandPermanent(g, game.Player1, types.Plains)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	preparedCast := action.CastSpellFaceFromZone(token.ObjectID, zone.Battlefield, game.FaceAlternate, nil, 0, nil)
	if !engine.applyAction(g, game.Player1, preparedCast) {
		t.Fatal("casting token's prepared spell copy failed")
	}
	if token.Prepared {
		t.Fatal("token remained prepared after casting its spell copy")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceTokenDef != token.TokenDef || obj.SourceID != token.ObjectID {
		t.Fatalf("stack object = %+v, want token-backed prepared spell", obj)
	}
	life := g.Players[game.Player1].Life
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Life != life+2 {
		t.Fatalf("life = %d, want %d", g.Players[game.Player1].Life, life+2)
	}
}

func TestTokenPreparedStormCopiesPreserveSourceDefinition(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	def := copyCardDef(prepareCard())
	alternate := def.Alternate.Val
	alternate.StaticAbilities = []game.StaticAbility{game.StormStaticBody}
	def.Alternate = opt.Val(alternate)
	token, _ := createTokenPermanent(g, game.Player1, def)
	addBasicLandPermanent(g, game.Player1, types.Plains)
	firstID := addCardToHand(g, game.Player1, simpleGainLifeInstant("First Spell"))
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(firstID, nil, 0, nil)) {
		t.Fatal("first spell cast failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	preparedCast := action.CastSpellFaceFromZone(token.ObjectID, zone.Battlefield, game.FaceAlternate, nil, 0, nil)
	if !engine.applyAction(g, game.Player1, preparedCast) {
		t.Fatal("casting token's prepared storm spell failed")
	}
	if g.Stack.Size() != 2 {
		t.Fatalf("stack size = %d, want original plus storm copy", g.Stack.Size())
	}
	life := g.Players[game.Player1].Life
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Life != life+4 {
		t.Fatalf("life = %d, want %d", g.Players[game.Player1].Life, life+4)
	}
}

func TestTokenPreparedSpellHonorsCantBeCounteredRule(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := copyCardDef(prepareCard())
	token, _ := createTokenPermanent(g, game.Player1, def)
	spell := &game.StackObject{
		ID:             g.IDGen.Next(),
		Kind:           game.StackSpell,
		SourceID:       token.ObjectID,
		Face:           game.FaceAlternate,
		SourceTokenDef: token.TokenDef,
		Controller:     game.Player1,
		Copy:           true,
	}
	g.Stack.Push(spell)
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Counter Shield",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:               game.RuleEffectCantBeCountered,
				AffectedController: game.ControllerYou,
				SpellTypes:         []types.Card{types.Sorcery},
			}},
		}},
	}})

	if counterStackObject(g, spell.ID) {
		t.Fatal("token-backed prepared spell was countered despite can't-be-countered rule")
	}
	if _, ok := stackObjectByID(g, spell.ID); !ok {
		t.Fatal("protected token-backed spell was removed from stack")
	}
}

func TestPreparedSpellCopyRespectsTimingAndControl(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, prepareCard())
	card, _ := g.GetCardInstance(cardID)
	g.Players[game.Player1].Hand.Remove(cardID)
	permanent, ok := createCardPermanent(g, card, game.Player1, zone.Hand)
	if !ok {
		t.Fatal("creating prepared permanent failed")
	}
	addBasicLandPermanent(g, game.Player1, types.Plains)
	g.Turn.Phase = game.PhaseBeginning
	g.Turn.Step = game.StepUpkeep

	preparedCast := action.CastSpellFaceFromZone(cardID, zone.Battlefield, game.FaceAlternate, nil, 0, nil)
	if engine.applyAction(g, game.Player1, preparedCast) {
		t.Fatal("sorcery prepared spell copy was cast outside sorcery timing")
	}
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	permanent.Controller = game.Player2
	if engine.applyAction(g, game.Player1, preparedCast) {
		t.Fatal("player cast prepared spell copy from permanent they do not control")
	}
}

func prepareCard() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:           "Prepared Creature",
			ManaCost:       opt.Val(cost.Mana{cost.O(2), cost.W}),
			Colors:         []color.Color{color.White},
			Types:          []types.Card{types.Creature},
			Power:          opt.Val(game.PT{Value: 3}),
			Toughness:      opt.Val(game.PT{Value: 3}),
			EntersPrepared: true,
		},
		Layout: game.LayoutPrepare,
		Alternate: opt.Val(game.CardFace{
			Name:     "Prepared Spell",
			ManaCost: opt.Val(cost.Mana{cost.W}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{{
					Primitive: game.GainLife{Amount: game.Fixed(2), Player: game.ControllerReference()},
				}},
			}.Ability()),
		}),
	}
}
