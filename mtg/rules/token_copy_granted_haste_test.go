package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: Oracle text from the pinned corpus for Feldon of the Third Path
// (0360f1ff-15c9-48e3-89eb-fbc4bf140c55): "{2}{R}, {T}: Create a token that's a
// copy of target creature card in your graveyard, except it's an artifact in
// addition to its other types. It gains haste. Sacrifice it at the beginning of
// the next end step."; Kiki-Jiki, Mirror Breaker (a2ff0ee3-9600-4c7d-acec-
// 6ec90595384e): "{T}: Create a token that's a copy of target nonlegendary
// creature you control, except it has haste. ..."; Orthion, Hero of Lavabrink
// (71dadbb3-7b8a-4656-973f-65a3284afe07): "{1}{R}, {T}: Create a token that's a
// copy of another target creature you control. It gains haste. ..." and
// "{6}{R}{R}{R}, {T}: Create five tokens ... They gain haste. ..."; Cogwork
// Assembler (e4bfde3f-f7d3-4902-b3cd-23f3fa53eff4): "{7}: Create a token that's
// a copy of target artifact. That token gains haste. Exile it at the beginning
// of the next end step."; Molten Echoes (461d9f1c-53a4-48a6-bafb-0e70099f276c):
// "Whenever a nontoken creature you control of the chosen type enters, create a
// token that's a copy of that creature. That token gains haste. Exile it at the
// beginning of the next end step."; Helm of the Host (70ffc71f-328d-421d-926b-
// 6f2e45ffb812): "At the beginning of combat on your turn, create a token that's
// a copy of equipped creature, except the token isn't legendary. That token
// gains haste."; Chrome Dome (994a01eb-2689-49e7-be23-0713da9da9a8) and
// Saheeli, the Sun's Brilliance (0ba99b60-c7d0-4041-a065-f2c510745223) use the
// same "That token gains haste"/"It gains haste" sentence; Cackling Counterpart
// (a865fdc5-955a-4c4a-90c0-342c15fd7bbc): "Create a token that's a copy of
// target creature you control." Feldon's 2021-03-19 rulings
// (https://api.scryfall.com/cards/0360f1ff-15c9-48e3-89eb-fbc4bf140c55/rulings):
// "If another creature becomes or enters the battlefield as a copy of the token,
// it will be an artifact in addition to its other types, but it won't gain
// haste and it won't be sacrificed." and "If a replacement effect causes the
// ability to create multiple tokens, each of them gains haste and will be
// sacrificed." CR 2026-09-25 (official TXT): 707.2 (a copy acquires only the
// copiable values), 707.9a (an "except it has" exception becomes a copiable
// value), 611.2c (a resolving effect's characteristic change locks in its
// affected objects) and 613.1f (ability-adding effects apply in layer 6). So a
// separate "gains haste" sentence grants haste to every created token through a
// layer-6 effect, never a copiable value, while Kiki-Jiki's "except it has
// haste" stays copiable.

func grantedHasteGame() (*game.Game, *Engine) {
	return game.NewGame([game.NumPlayers]game.PlayerConfig{}), NewEngineWithPrograms(nil, cards.NewDefaultRegistry().Programs())
}

func activateTokenCopyAbility(t *testing.T, g *game.Game, engine *Engine, source *game.Permanent, ability int, target game.Target, tokenName string) []id.ID {
	t.Helper()
	agents := [game.NumPlayers]PlayerAgent{}
	before := playerTokensNamed(g, game.Player1, tokenName)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyActionWithChoices(g, game.Player1, action.ActivateAbility(source.ObjectID, ability, []game.Target{target}, 0), agents, &TurnLog{}) {
		t.Fatalf("could not activate %s", permanentName(g, source))
	}
	resolveStackWithTriggers(engine, g, agents)
	return newTokens(before, playerTokensNamed(g, game.Player1, tokenName))
}

// copyWithCacklingCounterpart casts a real Cackling Counterpart on token and
// returns the single copy it creates.
func copyWithCacklingCounterpart(t *testing.T, g *game.Game, engine *Engine, token *game.Permanent) *game.Permanent {
	t.Helper()
	name := permanentName(g, token)
	before := playerTokensNamed(g, game.Player1, name)
	counterpart := addCardToHand(g, game.Player1, registryCard(t, "Cackling Counterpart"))
	g.Players[game.Player1].ManaPool.Add(mana.U, 3)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(counterpart, []game.Target{game.PermanentTarget(token.ObjectID)}, 0, nil)) {
		t.Fatal("could not cast Cackling Counterpart")
	}
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
	copies := newTokens(before, playerTokensNamed(g, game.Player1, name))
	if len(copies) != 1 {
		t.Fatalf("Cackling Counterpart copies = %d, want 1", len(copies))
	}
	copyOfToken, _ := g.PermanentByID(copies[0])
	return copyOfToken
}

func requireSingleHastyToken(t *testing.T, g *game.Game, tokens []id.ID, card string) *game.Permanent {
	t.Helper()
	if len(tokens) != 1 {
		t.Fatalf("%s created %d tokens, want 1", card, len(tokens))
	}
	token, _ := g.PermanentByID(tokens[0])
	if !hasKeyword(g, token, game.Haste) {
		t.Fatalf("%s's token lacks haste", card)
	}
	return token
}

// Feldon's rulings: the copy of its token is an artifact (copiable exception)
// but neither gains haste nor is sacrificed, while Feldon's token has haste
// and is sacrificed at the next end step.
func TestFeldonGrantedHasteIsNotCopiable(t *testing.T) {
	g, engine := grantedHasteGame()
	feldon := addCombatPermanent(g, game.Player1, registryCard(t, "Feldon of the Third Path"))
	blueprint := addCardToGraveyard(g, game.Player1, copyTestBear("Feldon Bear"))
	g.Players[game.Player1].ManaPool.Add(mana.R, 3)
	tokens := activateTokenCopyAbility(t, g, engine, feldon, 0, game.Target{Kind: game.TargetCard, CardID: blueprint}, "Feldon Bear")
	token := requireSingleHastyToken(t, g, tokens, "Feldon of the Third Path")
	if !permanentHasType(g, token, types.Artifact) {
		t.Fatal("Feldon's token is not an artifact")
	}

	copyOfToken := copyWithCacklingCounterpart(t, g, engine, token)
	if !permanentHasType(g, copyOfToken, types.Artifact) {
		t.Fatal("the copy of Feldon's token lost the copiable artifact exception")
	}
	if hasKeyword(g, copyOfToken, game.Haste) {
		t.Fatal("the copy of Feldon's token copied its granted haste")
	}

	fireEndStep(t, g, engine)
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := g.PermanentByID(token.ObjectID); ok {
		t.Fatal("Feldon's token was not sacrificed")
	}
	if _, ok := g.PermanentByID(copyOfToken.ObjectID); !ok {
		t.Fatal("the copy of Feldon's token was sacrificed")
	}
}

// Kiki-Jiki's "except it has haste" is part of the copy effect (CR 707.9a), so
// a copy of its token also has haste.
func TestKikiJikiExceptHasteRemainsCopiable(t *testing.T) {
	g, engine := grantedHasteGame()
	kiki := addCombatPermanent(g, game.Player1, registryCard(t, "Kiki-Jiki, Mirror Breaker"))
	bear := addCombatPermanent(g, game.Player1, copyTestBear("Kiki Bear"))
	tokens := activateTokenCopyAbility(t, g, engine, kiki, 0, game.PermanentTarget(bear.ObjectID), "Kiki Bear")
	token := requireSingleHastyToken(t, g, tokens, "Kiki-Jiki")
	copyOfToken := copyWithCacklingCounterpart(t, g, engine, token)
	if !hasKeyword(g, copyOfToken, game.Haste) {
		t.Fatal("the copy of Kiki-Jiki's token lost its copiable haste")
	}
}

// Orthion's "It gains haste"/"They gain haste" grants haste to every token it
// creates, and a copy of one of those tokens does not have haste.
func TestOrthionGrantedHasteIsNotCopiable(t *testing.T) {
	g, engine := grantedHasteGame()
	orthion := addCombatPermanent(g, game.Player1, registryCard(t, "Orthion, Hero of Lavabrink"))
	bear := addCombatPermanent(g, game.Player1, copyTestBear("Orthion Bear"))
	g.Players[game.Player1].ManaPool.Add(mana.R, 2)
	tokens := activateTokenCopyAbility(t, g, engine, orthion, 0, game.PermanentTarget(bear.ObjectID), "Orthion Bear")
	token := requireSingleHastyToken(t, g, tokens, "Orthion")
	if hasKeyword(g, bear, game.Haste) {
		t.Fatal("Orthion granted haste to the copied original")
	}
	copyOfToken := copyWithCacklingCounterpart(t, g, engine, token)
	if hasKeyword(g, copyOfToken, game.Haste) {
		t.Fatal("the copy of Orthion's token copied its granted haste")
	}
}

func TestOrthionFiveTokensAllGainHaste(t *testing.T) {
	g, engine := grantedHasteGame()
	orthion := addCombatPermanent(g, game.Player1, registryCard(t, "Orthion, Hero of Lavabrink"))
	bear := addCombatPermanent(g, game.Player1, copyTestBear("Orthion Bear"))
	g.Players[game.Player1].ManaPool.Add(mana.R, 9)
	tokens := activateTokenCopyAbility(t, g, engine, orthion, 1, game.PermanentTarget(bear.ObjectID), "Orthion Bear")
	if len(tokens) != 5 {
		t.Fatalf("Orthion created %d tokens, want 5", len(tokens))
	}
	for _, tokenID := range tokens {
		token, _ := g.PermanentByID(tokenID)
		if !hasKeyword(g, token, game.Haste) {
			t.Fatalf("Orthion token %d lacks haste", tokenID)
		}
	}
	fireEndStep(t, g, engine)
	engine.resolveTopOfStack(g, &TurnLog{})
	if left := onBattlefield(g, tokens); len(left) != 0 {
		t.Fatalf("Orthion tokens %v remained after the next end step", left)
	}
}

// The other "That token gains haste" activated copy abilities grant
// non-copiable haste the same way.
func TestTokenCopyActivatedAbilitiesGrantNonCopiableHaste(t *testing.T) {
	for _, tc := range []struct {
		card string
		pay  map[mana.Color]int
	}{
		{card: "Chrome Dome", pay: map[mana.Color]int{mana.C: 5}},
		{card: "Saheeli, the Sun's Brilliance", pay: map[mana.Color]int{mana.U: 1, mana.R: 1}},
	} {
		t.Run(tc.card, func(t *testing.T) {
			g, engine := grantedHasteGame()
			source := addCombatPermanent(g, game.Player1, registryCard(t, tc.card))
			blueprint := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name:      "Haste Test Construct",
				Types:     []types.Card{types.Artifact, types.Creature},
				Power:     opt.Val(game.PT{Value: 1}),
				Toughness: opt.Val(game.PT{Value: 1}),
			}})
			for color, amount := range tc.pay {
				g.Players[game.Player1].ManaPool.Add(color, amount)
			}
			tokens := activateTokenCopyAbility(t, g, engine, source, 0, game.PermanentTarget(blueprint.ObjectID), "Haste Test Construct")
			token := requireSingleHastyToken(t, g, tokens, tc.card)
			copyOfToken := copyWithCacklingCounterpart(t, g, engine, token)
			if hasKeyword(g, copyOfToken, game.Haste) {
				t.Fatalf("the copy of %s's token copied its granted haste", tc.card)
			}
			fireEndStep(t, g, engine)
			engine.resolveTopOfStack(g, &TurnLog{})
			if _, ok := g.PermanentByID(token.ObjectID); ok {
				t.Fatalf("%s's token was not sacrificed", tc.card)
			}
			if _, ok := g.PermanentByID(copyOfToken.ObjectID); !ok {
				t.Fatalf("the copy of %s's token was sacrificed", tc.card)
			}
		})
	}
}

// Cogwork Assembler copies the target artifact, grants the token haste, and
// exiles that token, not the target, at the next end step.
func TestCogworkAssemblerExilesItsHastyTokenAtEndStep(t *testing.T) {
	g, engine := grantedHasteGame()
	assembler := addCombatPermanent(g, game.Player1, registryCard(t, "Cogwork Assembler"))
	original := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Haste Test Relic",
		Types: []types.Card{types.Artifact},
	}})
	g.Players[game.Player1].ManaPool.Add(mana.C, 7)
	tokens := activateTokenCopyAbility(t, g, engine, assembler, 0, game.PermanentTarget(original.ObjectID), "Haste Test Relic")
	token := requireSingleHastyToken(t, g, tokens, "Cogwork Assembler")
	if _, ok := g.PermanentByID(original.ObjectID); !ok {
		t.Fatal("Cogwork Assembler removed the targeted artifact")
	}
	fireEndStep(t, g, engine)
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := g.PermanentByID(token.ObjectID); ok {
		t.Fatal("Cogwork Assembler's token was not exiled at the next end step")
	}
	if _, ok := g.PermanentByID(original.ObjectID); !ok {
		t.Fatal("Cogwork Assembler's end-step exile removed the original artifact")
	}
}

// Molten Echoes' token of the entering creature gains haste and is exiled at
// the next end step; the entering creature itself is untouched.
func TestMoltenEchoesHastyTokenIsExiledAtEndStep(t *testing.T) {
	g, engine := grantedHasteGame()
	echoes := addCombatPermanent(g, game.Player1, registryCard(t, "Molten Echoes"))
	echoes.EntryChoices = map[game.ChoiceKey]game.ResolutionChoiceResult{
		game.EntryTypeChoiceKey: {Kind: game.ResolutionChoiceSubtype, Subtype: types.Bear},
	}
	bearCard := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Echo Bear",
		ManaCost:  opt.Val(cost.Mana{cost.O(1)}),
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Bear},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(bearCard, nil, 0, nil)) {
		t.Fatal("could not cast Echo Bear")
	}
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
	tokens := playerTokensNamed(g, game.Player1, "Echo Bear")
	token := requireSingleHastyToken(t, g, tokens, "Molten Echoes")
	var original *game.Permanent
	for _, permanent := range g.Battlefield {
		if !permanent.Token && permanentName(g, permanent) == "Echo Bear" {
			original = permanent
		}
	}
	if original == nil || hasKeyword(g, original, game.Haste) {
		t.Fatal("Molten Echoes' entering creature is missing or gained haste")
	}
	fireEndStep(t, g, engine)
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := g.PermanentByID(token.ObjectID); ok {
		t.Fatal("Molten Echoes' token was not exiled at the next end step")
	}
	if _, ok := g.PermanentByID(original.ObjectID); !ok {
		t.Fatal("Molten Echoes' end-step exile removed the entering creature")
	}
}

// Helm of the Host's token is not legendary (copiable) and has haste that a
// copy of it does not get.
func TestHelmOfTheHostGrantedHasteIsNotCopiable(t *testing.T) {
	g, engine := grantedHasteGame()
	helm := addCombatPermanent(g, game.Player1, registryCard(t, "Helm of the Host"))
	legend := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:       "Helm Test Legend",
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Power:      opt.Val(game.PT{Value: 2}),
		Toughness:  opt.Val(game.PT{Value: 2}),
	}})
	attachPlate(legend, helm)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepBeginningOfCombat
	emitBeginningOfStepEvent(g, game.StepBeginningOfCombat)
	agents := [game.NumPlayers]PlayerAgent{}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("Helm of the Host did not trigger at the beginning of combat")
	}
	resolveStackWithTriggers(engine, g, agents)
	tokens := playerTokensNamed(g, game.Player1, "Helm Test Legend")
	token := requireSingleHastyToken(t, g, tokens, "Helm of the Host")
	if hasKeyword(g, legend, game.Haste) {
		t.Fatal("Helm of the Host granted haste to the equipped creature")
	}
	copyOfToken := copyWithCacklingCounterpart(t, g, engine, token)
	if hasKeyword(g, copyOfToken, game.Haste) {
		t.Fatal("the copy of Helm of the Host's token copied its granted haste")
	}
	if _, ok := g.PermanentByID(token.ObjectID); !ok {
		t.Fatal("the nonlegendary Helm token was lost to the legend rule")
	}
}
