package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4:
// Restless Dreams Oracle ID 65f1dd09-0640-47f8-8bfd-7aceaa5646a6, printing
// 63795680-5310-4c99-92c1-70e453391de9 ("As an additional cost to cast this
// spell, discard X cards. Return X target creature cards from your graveyard to
// your hand."); Turbulent Dreams Oracle ID 83134570-debe-4bc0-9656-2a16801a3592,
// printing d879226b-8ee3-4a48-b03a-d183594dc586 ("As an additional cost to cast
// this spell, discard X cards. Return X target nonland permanents to their
// owners' hands."); Nostalgic Dreams Oracle ID fabbc474-682c-4565-861c-d41de05abb8f,
// printing 299065ad-ee2e-4973-9b41-fd04c36d19a4 ("As an additional cost to cast
// this spell, discard X cards. Return X target cards from your graveyard to your
// hand. Exile Nostalgic Dreams."). Current Scryfall card records returned the
// same text and exact Oracle IDs on 2026-11-11. The rulings URLs
// https://api.scryfall.com/cards/63795680-5310-4c99-92c1-70e453391de9/rulings,
// https://api.scryfall.com/cards/d879226b-8ee3-4a48-b03a-d183594dc586/rulings,
// and https://api.scryfall.com/cards/299065ad-ee2e-4973-9b41-fd04c36d19a4/rulings
// each returned {"data":[],"has_more":false}.
//
// Official WotC rules page https://magic.wizards.com/en/rules linked the
// Comprehensive Rules TXT
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt,
// effective 2026-11-06 (the latest edition listed there; the repository's
// existing support report records the prior 2026-09-25 rules pin). Read 107.3a/i
// and 601.2b/c/f-h: announce one X for the spell, select exactly X targets,
// then pay the mandatory X-card discard alongside its mana cost. Rule 118.8
// establishes that a mandatory additional cost is part of casting. Rules
// 115.1a and 608.2b require targets to be chosen on casting and revalidated on
// resolution; 608.2c/f require following instructions and processing group
// movement simultaneously. Rule 108.3 makes a card's owner, not its controller,
// define ownership and 400.3 sends a returned card to that owner's hand. Rule
// 400.7 makes zone changes create new objects. Rules 608.2m/n require a resolving
// spell to finish resolving after it leaves the stack and give it a normal
// graveyard destination; Nostalgic Dreams' explicit exile instruction replaces
// that destination through MoveResolvingSpell.

func xDiscardDreamsRegisteredCard(t *testing.T, entry cardset.Entry) (*game.CardDef, *Engine) {
	t.Helper()
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(entry.OracleID)
	if err != nil {
		t.Fatalf("%s exact default-registry lookup failed: %v", entry.Name, err)
	}
	if def.Name != entry.Name {
		t.Fatalf("%s exact default-registry lookup returned %q", entry.Name, def.Name)
	}
	return def, NewEngineWithPrograms(nil, registry.Programs())
}

func TestRestlessDreamsDiscardsXAndReturnsCreatureCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	card, engine := xDiscardDreamsRegisteredCard(t, authored.RestlessDreams)
	targets := []game.Target{
		currentCardTarget(t, g, addCardToGraveyard(g, game.Player1, vanillaCreature("Dream creature one", 2, 2))),
		currentCardTarget(t, g, addCardToGraveyard(g, game.Player1, vanillaCreature("Dream creature two", 3, 3))),
	}
	artifact := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Dream artifact card", Types: []types.Card{types.Artifact},
	}})
	discardOne := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Dream discard one"}})
	discardTwo := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Dream discard two"}})
	spell := addCardToHand(g, game.Player1, card)
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spell, targets, 2, nil)) {
		t.Fatal("Restless Dreams could not pay its X=2 discard cost and target two creature cards")
	}
	if !g.Players[game.Player1].Graveyard.Contains(discardOne) ||
		!g.Players[game.Player1].Graveyard.Contains(discardTwo) {
		t.Fatal("Restless Dreams did not discard exactly two cards as its additional cost")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	for _, target := range targets {
		if !g.Players[game.Player1].Hand.Contains(target.CardID) {
			t.Fatal("Restless Dreams did not return a targeted creature card")
		}
	}
	if !g.Players[game.Player1].Graveyard.Contains(artifact) || g.Players[game.Player1].Hand.Contains(artifact) {
		t.Fatal("Restless Dreams returned a noncreature card")
	}
	if !g.Players[game.Player1].Graveyard.Contains(spell) {
		t.Fatal("Restless Dreams did not finish in its owner's graveyard")
	}
}

func TestTurbulentDreamsDiscardsXAndReturnsNonlandPermanents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	card, engine := xDiscardDreamsRegisteredCard(t, authored.TurbulentDreams)
	first := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Dream creature permanent", Types: []types.Card{types.Creature},
	}})
	first.Controller = game.Player2
	second := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Dream artifact permanent", Types: []types.Card{types.Artifact},
	}})
	second.Controller = game.Player1
	land := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Dream land", Types: []types.Card{types.Land},
	}})
	discardOne := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Dream discard one"}})
	discardTwo := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Dream discard two"}})
	spell := addCardToHand(g, game.Player1, card)
	g.Players[game.Player1].ManaPool.Add(mana.U, 2)

	targets := []game.Target{game.PermanentTarget(first.ObjectID), game.PermanentTarget(second.ObjectID)}
	if !engine.applyAction(g, game.Player1, action.CastSpell(spell, targets, 2, nil)) {
		t.Fatal("Turbulent Dreams could not pay its X=2 discard cost and target two nonland permanents")
	}
	if !g.Players[game.Player1].Graveyard.Contains(discardOne) ||
		!g.Players[game.Player1].Graveyard.Contains(discardTwo) {
		t.Fatal("Turbulent Dreams did not discard exactly two cards as its additional cost")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Hand.Contains(first.CardInstanceID) ||
		!g.Players[game.Player2].Hand.Contains(second.CardInstanceID) ||
		g.Players[game.Player2].Hand.Contains(first.CardInstanceID) ||
		g.Players[game.Player1].Hand.Contains(second.CardInstanceID) {
		t.Fatal("Turbulent Dreams did not return each nonland permanent to its owner's hand")
	}
	if permanentByCardID(g, land.CardInstanceID) == nil {
		t.Fatal("Turbulent Dreams returned a land")
	}
	if !g.Players[game.Player1].Graveyard.Contains(spell) {
		t.Fatal("Turbulent Dreams did not finish in its owner's graveyard")
	}
}

func TestNostalgicDreamsDiscardsXReturnsCardsAndExiles(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	card, engine := xDiscardDreamsRegisteredCard(t, authored.NostalgicDreams)
	first := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Dream artifact card", Types: []types.Card{types.Artifact},
	}})
	second := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Dream land card", Types: []types.Card{types.Land},
	}})
	discardOne := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Dream discard one"}})
	discardTwo := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Dream discard two"}})
	spell := addCardToHand(g, game.Player1, card)
	g.Players[game.Player1].ManaPool.Add(mana.G, 2)

	targets := []game.Target{currentCardTarget(t, g, first), currentCardTarget(t, g, second)}
	if !engine.applyAction(g, game.Player1, action.CastSpell(spell, targets, 2, nil)) {
		t.Fatal("Nostalgic Dreams could not pay its X=2 discard cost and target two cards")
	}
	if !g.Players[game.Player1].Graveyard.Contains(discardOne) ||
		!g.Players[game.Player1].Graveyard.Contains(discardTwo) {
		t.Fatal("Nostalgic Dreams did not discard exactly two cards as its additional cost")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Hand.Contains(first) || !g.Players[game.Player1].Hand.Contains(second) {
		t.Fatal("Nostalgic Dreams did not return both targeted cards to hand")
	}
	if !g.Players[game.Player1].Exile.Contains(spell) || g.Players[game.Player1].Graveyard.Contains(spell) {
		t.Fatal("Nostalgic Dreams did not exile its resolving spell")
	}
}
