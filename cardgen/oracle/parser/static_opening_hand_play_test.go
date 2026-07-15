package parser

import "testing"

func TestParseStaticOpeningHandPlayDeclarationMeaning(t *testing.T) {
	t.Parallel()
	declarations := parseStaticDeclarationSyntax(t,
		"If this card is in your opening hand, you may begin the game with it on the battlefield.",
		Context{CardName: "Leyline of Sanctity"})
	if len(declarations) != 1 {
		t.Fatalf("declarations = %#v, want one", declarations)
	}
	if declarations[0].Kind != StaticDeclarationOpeningHandPlay {
		t.Fatalf("kind = %v, want opening-hand play", declarations[0].Kind)
	}
}

func TestParseStaticOpeningHandPlayDeclarationFailsClosed(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		// "a card" is a different subject than the printed "this card".
		"a card instead of this card": "If a card is in your opening hand, you may begin the game with it on the battlefield.",
		// "opening hands" (plural) is not the singular opening-hand wording.
		"plural opening hands": "If this card is in your opening hands, you may begin the game with it on the battlefield.",
		// "start" is not the printed verb "begin".
		"start instead of begin": "If this card is in your opening hand, you may start the game with it on the battlefield.",
		// "onto" is not the printed preposition "on".
		"onto instead of on": "If this card is in your opening hand, you may begin the game with it onto the battlefield.",
		// Dropping the comma changes the token shape and is rejected.
		"missing comma": "If this card is in your opening hand you may begin the game with it on the battlefield.",
		// CR 103.6b reveal-only variants are an explicit non-goal and must not match.
		"reveal only variant": "If this card is in your opening hand, you may reveal it. If you do, put it onto the battlefield.",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(source, Context{CardName: "Tester"})
			for _, ability := range document.Abilities {
				for _, declaration := range ability.StaticDeclarations {
					if declaration.Kind == StaticDeclarationOpeningHandPlay {
						t.Fatalf("source %q matched the opening-hand play declaration", source)
					}
				}
			}
		})
	}
}
