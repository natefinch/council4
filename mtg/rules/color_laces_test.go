package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

func castColorLaceOnPermanent(t *testing.T, entry cardset.Entry, payment mana.Color, wantColors ...color.Color) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCreaturePermanent(g, game.Player2)
	targetDef := g.CardInstances[target.CardInstanceID].Def
	targetDef.Colors = []color.Color{color.Blue, color.Red}
	targetDef.ManaCost = opt.Val(cost.Mana{cost.U, cost.R})

	lace := addCardToHand(g, game.Player1, entry.New())
	g.Players[game.Player1].ManaPool.Add(payment, 1)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(lace,
		[]game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)) {
		t.Fatalf("casting %s for its printed mana cost targeting a permanent failed", entry.Name)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := permanentEffectiveColors(g, target); !slices.Equal(got, wantColors) {
		t.Fatalf("%s target colors = %v, want %v", entry.Name, got, wantColors)
	}
	if got := targetDef.ManaCost.Val; !slices.Equal(got, cost.Mana{cost.U, cost.R}) {
		t.Fatalf("%s changed the target's mana symbols: got %v, want {U}{R}", entry.Name, got)
	}
}

// Source review: Purelace Oracle ID 3773001a-8868-49ec-a406-298cf72359c2,
// Fourth Edition printing revision 5736fcc7-fa95-4ef4-b821-392ec00e03bf;
// mana cost {W}; Oracle text: "Target spell or permanent becomes white. (Mana symbols on that permanent remain unchanged.)"
// Scryfall rulings URI https://api.scryfall.com/cards/5736fcc7-fa95-4ef4-b821-392ec00e03bf/rulings returned no rulings.
// Official CR effective November 6, 2026, linked from https://magic.wizards.com/en/rules:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt. Read 105.3 (a new color
// replaces previous colors), 115.1a and 601.2c/f/h (choose targets, determine and pay costs),
// 202.2 (color and mana symbols are separate characteristics), 400.7a (color changes on a
// permanent spell continue to its resulting permanent), 608.2b (recheck target legality),
// 611.2a/c (no-duration resolving effect lasts through the game and fixes its affected set),
// and 613.1e (color changes apply in layer 5). Thus the chosen permanent becomes white without
// changing its mana symbols.
func TestPurelaceTargetPermanentBecomesWhite(t *testing.T) {
	castColorLaceOnPermanent(t, authored.Purelace, mana.W, color.White)
}

// Source review: Moonlace Oracle ID 52d4dc8a-2799-4b4e-9c20-f14bb75c2279,
// Time Spiral printing revision c781cc80-5f57-4677-a9da-3523191cc7c6;
// mana cost {U}; Oracle text: "Target spell or permanent becomes colorless."
// Scryfall rulings URI https://api.scryfall.com/cards/c781cc80-5f57-4677-a9da-3523191cc7c6/rulings
// returned two rulings: the effect doesn't wear off, and a targeted permanent spell enters and
// remains colorless until it leaves or another effect changes its color.
// Official CR effective November 6, 2026, linked from https://magic.wizards.com/en/rules:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt. Read 105.3 (an effect
// can make a colored object colorless), 115.1a and 601.2c/f/h (choose targets, determine and pay
// costs), 202.2 (color and mana symbols are separate characteristics), 400.7a (color changes on
// a permanent spell continue to its resulting permanent), 608.2b (recheck target legality),
// 611.2a/c (no-duration resolving effect lasts through the game and fixes its affected set),
// and 613.1e (color changes apply in layer 5). Thus the chosen permanent becomes colorless while
// retaining its mana symbols.
func TestMoonlaceTargetPermanentBecomesColorless(t *testing.T) {
	castColorLaceOnPermanent(t, authored.Moonlace, mana.U)
}

// Source review: Thoughtlace Oracle ID 6452b6a6-6235-46a3-a712-a26592450438,
// Fourth Edition printing revision 185674a4-db97-444d-b0e9-7dcfc245ce4b;
// mana cost {U}; Oracle text: "Target spell or permanent becomes blue. (Mana symbols on that permanent remain unchanged.)"
// Scryfall rulings URI https://api.scryfall.com/cards/185674a4-db97-444d-b0e9-7dcfc245ce4b/rulings returned no rulings.
// Official CR effective November 6, 2026, linked from https://magic.wizards.com/en/rules:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt. Read 105.3 (a new color
// replaces previous colors), 115.1a and 601.2c/f/h (choose targets, determine and pay costs),
// 202.2 (color and mana symbols are separate characteristics), 400.7a (color changes on a
// permanent spell continue to its resulting permanent), 608.2b (recheck target legality),
// 611.2a/c (no-duration resolving effect lasts through the game and fixes its affected set),
// and 613.1e (color changes apply in layer 5). Thus the chosen permanent becomes blue without
// changing its mana symbols.
func TestThoughtlaceTargetPermanentBecomesBlue(t *testing.T) {
	castColorLaceOnPermanent(t, authored.Thoughtlace, mana.U, color.Blue)
}

// Source review: Lifelace Oracle ID eec1de80-4b3d-481d-a235-c299e0381830,
// Fourth Edition printing revision bf312acb-2fa6-440a-964b-424ad8abc331;
// mana cost {G}; Oracle text: "Target spell or permanent becomes green. (Mana symbols on that permanent remain unchanged.)"
// Scryfall rulings URI https://api.scryfall.com/cards/bf312acb-2fa6-440a-964b-424ad8abc331/rulings returned no rulings.
// Official CR effective November 6, 2026, linked from https://magic.wizards.com/en/rules:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt. Read 105.3 (a new color
// replaces previous colors), 115.1a and 601.2c/f/h (choose targets, determine and pay costs),
// 202.2 (color and mana symbols are separate characteristics), 400.7a (color changes on a
// permanent spell continue to its resulting permanent), 608.2b (recheck target legality),
// 611.2a/c (no-duration resolving effect lasts through the game and fixes its affected set),
// and 613.1e (color changes apply in layer 5). Thus the chosen permanent becomes green without
// changing its mana symbols.
func TestLifelaceTargetPermanentBecomesGreen(t *testing.T) {
	castColorLaceOnPermanent(t, authored.Lifelace, mana.G, color.Green)
}

// Source review: Deathlace Oracle ID fb80aaba-352a-4b58-8db2-1e02d542819c,
// Fourth Edition printing revision 237e37fb-383d-432c-8ac3-1332096567db;
// mana cost {B}; Oracle text: "Target spell or permanent becomes black. (Mana symbols on that permanent remain unchanged.)"
// Scryfall rulings URI https://api.scryfall.com/cards/237e37fb-383d-432c-8ac3-1332096567db/rulings returned no rulings.
// Official CR effective November 6, 2026, linked from https://magic.wizards.com/en/rules:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt. Read 105.3 (a new color
// replaces previous colors), 115.1a and 601.2c/f/h (choose targets, determine and pay costs),
// 202.2 (color and mana symbols are separate characteristics), 400.7a (color changes on a
// permanent spell continue to its resulting permanent), 608.2b (recheck target legality),
// 611.2a/c (no-duration resolving effect lasts through the game and fixes its affected set),
// and 613.1e (color changes apply in layer 5). Thus the chosen permanent becomes black without
// changing its mana symbols.
func TestDeathlaceTargetPermanentBecomesBlack(t *testing.T) {
	castColorLaceOnPermanent(t, authored.Deathlace, mana.B, color.Black)
}
