package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func confluenceDefinition(t *testing.T, oracleID, name string) *game.CardDef {
	t.Helper()
	def, err := cards.NewDefaultRegistry().LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("default registry lookup for %q failed: %v", name, err)
	}
	if def.Name != name {
		t.Fatalf("Oracle lookup %s returned %q, want %q", oracleID, def.Name, name)
	}
	return def
}

func castConfluence(t *testing.T, engine *Engine, g *game.Game, def *game.CardDef, targets []game.Target, modes []int) *game.StackObject {
	t.Helper()
	cardID := addCardToHand(g, game.Player1, def)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, targets, 0, modes)) {
		t.Fatalf("casting %s with modes %v and targets %v failed", def.Name, modes, targets)
	}
	spell, ok := g.Stack.Peek()
	if !ok || spell.SourceID != cardID {
		t.Fatalf("stack top = %+v, want %s", spell, def.Name)
	}
	if !slices.Equal(spell.ChosenModes, modes) {
		t.Fatalf("%s modes = %v, want %v", def.Name, spell.ChosenModes, modes)
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("%s left %d mana unspent after casting", def.Name, g.Players[game.Player1].ManaPool.Total())
	}
	return spell
}

// Source review: Oracle input revision 85d431c2-e4fa-426d-83cb-ecb782e7e87d;
// pinned Oracle text:
// "Choose three. You may choose the same mode more than once.
// • Fiery Confluence deals 1 damage to each creature.
// • Fiery Confluence deals 2 damage to each opponent.
// • Destroy target artifact."
// https://api.scryfall.com/cards/85d431c2-e4fa-426d-83cb-ecb782e7e87d/rulings
// returned seven rulings, including that repeat damage modes are separate damage
// events and that chosen modes resolve in printed order. Official Comprehensive
// Rules effective 2026-11-06, https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (linked from https://magic.wizards.com/en/rules), sections 601.2b-c, 608.2c,
// 700.2a,d, 120.1, 120.4 and 701.8: modes/targets are chosen while casting,
// repeated damage instructions resolve in order, and the artifact is destroyed.
func TestFieryConfluenceRepeatsOpponentDamageAndDestroysArtifact(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	artifact := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Target Artifact", Types: []types.Card{types.Artifact},
	}})
	startingLife := [game.NumPlayers]int{}
	for player := range game.NumPlayers {
		startingLife[player] = g.Players[player].Life
	}
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.R, 2)

	castConfluence(t, engine, g, confluenceDefinition(t,
		"3c22e031-4804-4c31-bd3c-c3f29d456b34", "Fiery Confluence"),
		[]game.Target{game.PermanentTarget(artifact.ObjectID)}, []int{1, 1, 2})
	engine.resolveTopOfStack(g, &TurnLog{})

	for player := game.Player2; player < game.NumPlayers; player++ {
		if got, want := g.Players[player].Life, startingLife[player]-4; got != want {
			t.Errorf("player %d life = %d, want %d after the repeated opponent-damage mode", player, got, want)
		}
	}
	if _, ok := permanentByObjectID(g, artifact.ObjectID); ok {
		t.Fatal("target artifact remained on the battlefield")
	}
	if !g.Players[game.Player2].Graveyard.Contains(artifact.CardInstanceID) {
		t.Fatal("target artifact was not destroyed into its owner's graveyard")
	}
}

type mysticConfluencePayPolicy struct {
	simPassPolicy
	requests []game.ChoiceRequest
}

func (p *mysticConfluencePayPolicy) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	p.requests = append(p.requests, request)
	if request.Kind == game.ChoiceMay {
		return []int{1}
	}
	return request.DefaultSelection
}

// Source review: Oracle input revision 62e5d409-c0b6-4123-802d-eb32f223bd1a;
// pinned Oracle text:
// "Choose three. You may choose the same mode more than once.
// • Counter target spell unless its controller pays {3}.
// • Return target creature to its owner's hand.
// • Draw a card."
// https://api.scryfall.com/cards/62e5d409-c0b6-4123-802d-eb32f223bd1a/rulings
// returned five rulings, including the controller's pay-or-counter choice,
// mode-dependent targets, and resolution in printed order. Official Comprehensive
// Rules effective 2026-11-06, https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (linked from https://magic.wizards.com/en/rules), sections 601.2b-c, 608.2c,
// 700.2a,d, 701.6a-b, 118.12a and 121.1: modes and targets are chosen while
// casting, the target's controller may pay on resolution, and repeated draw
// modes resolve in order after that choice.
func TestMysticConfluencePaysForCounterAndDrawsTwice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	targetSpell := addStackSpell(g, game.Player2, "Target Spell", []types.Card{types.Instant})
	drawnOne := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Draw One"}})
	drawnTwo := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Draw Two"}})
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	g.Players[game.Player1].ManaPool.Add(mana.U, 2)
	g.Players[game.Player2].ManaPool.Add(mana.C, 3)

	castConfluence(t, engine, g, confluenceDefinition(t,
		"cd11c27b-9368-4622-bc5a-2e0b993cc42b", "Mystic Confluence"),
		[]game.Target{game.StackObjectTarget(targetSpell.ID)}, []int{0, 2, 2})
	payPolicy := &mysticConfluencePayPolicy{}
	policies := simPassPolicies()
	policies[game.Player2] = payPolicy
	engine.resolveTopOfStackWithChoices(g, policies, &TurnLog{})

	if len(payPolicy.requests) != 1 || payPolicy.requests[0].Kind != game.ChoiceMay ||
		payPolicy.requests[0].Player != game.Player2 {
		t.Fatalf("payment requests = %+v, want one pay choice for the targeted spell's controller", payPolicy.requests)
	}
	if g.Players[game.Player2].ManaPool.Total() != 0 {
		t.Fatalf("target spell's controller has %d mana after paying {3}", g.Players[game.Player2].ManaPool.Total())
	}
	if top, ok := g.Stack.Peek(); !ok || top.ID != targetSpell.ID {
		t.Fatalf("target spell stack top = %+v, want the spell to remain after its controller paid", top)
	}
	if !g.Players[game.Player1].Hand.Contains(drawnOne) || !g.Players[game.Player1].Hand.Contains(drawnTwo) {
		t.Fatal("repeated draw mode did not draw both cards for Mystic Confluence's controller")
	}
}

// Source review: Oracle input revision 89c2faf6-e65b-4962-82d4-0294912c2ddb;
// pinned Oracle text:
// "Choose three. You may choose the same mode more than once.
// • Create a 2/2 white Knight creature token with vigilance.
// • Exile target enchantment.
// • You gain 5 life."
// https://api.scryfall.com/cards/89c2faf6-e65b-4962-82d4-0294912c2ddb/rulings
// returned seven rulings, including that repeated modes resolve in chosen order
// and repeated life modes are separate events. Official Comprehensive Rules
// effective 2026-11-06, https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (linked from https://magic.wizards.com/en/rules), sections 601.2b-c, 608.2c,
// 700.2a,d, 111.1-4, 119.3, 701.13a and 702.20a-b: repeated modes resolve in
// order, tokens are 2/2 white Knights with vigilance, the enchantment is exiled,
// and the life mode increases your life total.
func TestRighteousConfluenceCreatesTwoVigilantKnightsAndExilesEnchantment(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngineWithPrograms(nil, cards.NewDefaultRegistry().Programs())
	enchantment := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Target Enchantment", Types: []types.Card{types.Enchantment},
	}})
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	g.Players[game.Player1].ManaPool.Add(mana.W, 2)

	def := confluenceDefinition(t, "ee28937b-7ead-45ad-8448-35c51a09fb32", "Righteous Confluence")
	castConfluence(t, engine, g, def,
		[]game.Target{game.PermanentTarget(enchantment.ObjectID)}, []int{0, 0, 1})
	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player2].Exile.Contains(enchantment.CardInstanceID) {
		t.Fatal("target enchantment was not exiled")
	}
	knights := 0
	for _, permanent := range g.Battlefield {
		if !permanent.Token || permanent.TokenDef == nil ||
			!slices.Contains(permanent.TokenDef.Types, types.Creature) ||
			!slices.Contains(permanent.TokenDef.Subtypes, types.Knight) {
			continue
		}
		knights++
		face := permanent.TokenDef.CardFace
		if permanent.Controller != game.Player1 ||
			permanent.Owner != game.Player1 ||
			!slices.Contains(face.Colors, color.White) ||
			!slices.Contains(face.Types, types.Creature) ||
			!slices.Contains(face.Subtypes, types.Knight) ||
			!face.Power.Exists || face.Power.Val.Value != 2 ||
			!face.Toughness.Exists || face.Toughness.Val.Value != 2 ||
			!permanent.TokenDef.HasKeyword(game.Vigilance) {
			t.Errorf("created token is not a 2/2 white vigilant Knight: permanent=%+v face=%+v", permanent, face)
		}
	}
	if knights != 2 {
		t.Fatalf("Knight tokens = %d, want 2 from choosing the token mode twice", knights)
	}
}

// Source review: Oracle input revision 20203edb-1363-4ed3-8c5b-b1d96ff2b7c0;
// pinned Oracle text:
// "Choose three. You may choose the same mode more than once.
// • Target player draws a card and loses 1 life.
// • Target creature gets -2/-2 until end of turn.
// • Return target creature card from your graveyard to your hand."
// https://api.scryfall.com/cards/20203edb-1363-4ed3-8c5b-b1d96ff2b7c0/rulings
// returned six rulings, including that modes resolve in chosen order and that
// each repeated targeted mode may use its own target. Official Comprehensive
// Rules effective 2026-11-06, https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (linked from https://magic.wizards.com/en/rules), sections 601.2b-c, 608.2c,
// 700.2a,d, 119.3, 121.1, 400.3, 400.7 and 611.2a: repeated occurrences have
// separate target choices and instructions, each player draws and loses life,
// and the creature card moves to its owner's hand.
func TestWretchedConfluenceRepeatsPlayerModeAndReturnsCreatureCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	playerTwoDraw := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Player Two Draw"}})
	playerThreeDraw := addCardToLibrary(g, game.Player3, &game.CardDef{CardFace: game.CardFace{Name: "Player Three Draw"}})
	creatureCard := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Creature Card", Types: []types.Card{types.Creature},
	}})
	playerTwoLife := g.Players[game.Player2].Life
	playerThreeLife := g.Players[game.Player3].Life
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	g.Players[game.Player1].ManaPool.Add(mana.B, 2)

	castConfluence(t, engine, g, confluenceDefinition(t,
		"fb430640-81b6-4720-84d6-b8ec3af6a2c3", "Wretched Confluence"),
		[]game.Target{
			game.PlayerTarget(game.Player2),
			game.PlayerTarget(game.Player3),
			game.CardTarget(creatureCard),
		},
		[]int{0, 0, 2})
	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player2].Life != playerTwoLife-1 ||
		g.Players[game.Player3].Life != playerThreeLife-1 {
		t.Fatalf("life totals after repeated player mode = (%d, %d), want (%d, %d)",
			g.Players[game.Player2].Life, g.Players[game.Player3].Life, playerTwoLife-1, playerThreeLife-1)
	}
	if !g.Players[game.Player2].Hand.Contains(playerTwoDraw) ||
		!g.Players[game.Player3].Hand.Contains(playerThreeDraw) {
		t.Fatal("each target player did not draw from their repeated Wretched Confluence mode")
	}
	if !g.Players[game.Player1].Hand.Contains(creatureCard) ||
		g.Players[game.Player1].Graveyard.Contains(creatureCard) {
		t.Fatal("target creature card was not returned from its owner's graveyard to hand")
	}
}
