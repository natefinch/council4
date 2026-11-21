package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: Odyssey printing e588cc41-4b86-4e93-94e8-dd765b27ab63;
// Oracle ID 1c4b6543-777e-4c3b-a9fb-5b7210d458d5; pinned revision
// e588cc41-4b86-4e93-94e8-dd765b27ab63. The current card response and exact
// Oracle text are retained at .cardwork/skycloud-egg/card.json. The printing's
// https://api.scryfall.com/cards/e588cc41-4b86-4e93-94e8-dd765b27ab63/rulings
// response is retained at .cardwork/skycloud-egg/rulings.json; it contains one
// WotC ruling dated 2008-08-01, which calls this a mana ability and says it can
// be activated while casting a spell or activating an ability.
//
// Read the current official Comprehensive Rules edition 2026-11-06 from the
// Wizards rules page (https://magic.wizards.com/en/rules), whose current links
// identify https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.docx
// and the corresponding .txt. Verified local copies are retained at
// .cardwork/skycloud-egg/comprehensive-rules-2026-11-06.docx and .txt. CR 121.1
// defines drawing as moving the top library card to hand, which fails CR 605.1a's
// no-library-movement criterion for a mana ability. The 2008 ruling conflicts
// with this current rule. Accordingly, the card is represented as an ordinary
// activated ability: CR 113.3b and 602.2a-b put it on the stack, 602.5e enforces
// its instant-speed instruction, and 608.2c resolves its Add {W}, Add {U}, Draw
// instructions in order. It is not a mana ability under this rules edition.
const skycloudEggTestOracleID = "1c4b6543-777e-4c3b-a9fb-5b7210d458d5"

func skycloudEggTestDefinition(t *testing.T) *game.CardDef {
	t.Helper()
	def, err := cards.NewDefaultRegistry().LookupOracleID(skycloudEggTestOracleID)
	if err != nil {
		t.Fatalf("exact Skycloud Egg default-registry lookup failed: %v", err)
	}
	if def.Name != "Skycloud Egg" {
		t.Fatalf("exact Skycloud Egg lookup returned %q", def.Name)
	}
	return def
}

func TestSkycloudEggPaysCostsThenAddsManaAndDraws(t *testing.T) {
	def := skycloudEggTestDefinition(t)
	g := mainPhaseGame(game.Player1)
	setMainPhasePriority(g, game.Player1)
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, def)
	drawnCard := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Skycloud Egg draw"}})
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)

	activation := action.ActivateAbility(source.ObjectID, def.ActivatedAbilityIndex(0), nil, 0)
	if !actionsContain(engine.legalActions(g, game.Player1), activation) ||
		!engine.applyAction(g, game.Player1, activation) {
		t.Fatal("{2}, tap, and sacrifice activation was not legally completed")
	}
	if !source.Tapped || !g.Players[game.Player1].Graveyard.Contains(source.CardInstanceID) ||
		g.Players[game.Player1].ManaPool.Total() != 0 || g.Players[game.Player1].Hand.Size() != 0 ||
		!g.Players[game.Player1].Library.Contains(drawnCard) || g.Stack.Size() != 1 {
		t.Fatal("activation did not pay {2}, tap and sacrifice Skycloud Egg, and put the ability on the stack before its effects")
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].ManaPool.Amount(mana.W) != 1 ||
		g.Players[game.Player1].ManaPool.Amount(mana.U) != 1 ||
		!g.Players[game.Player1].Hand.Contains(drawnCard) ||
		g.Players[game.Player1].Library.Size() != 0 || g.Stack.Size() != 0 {
		t.Fatal("resolving Skycloud Egg did not add {W}{U}, draw the top card, and remove the ability from the stack")
	}
}
