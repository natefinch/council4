package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

// FaceDocument pairs a face's name with the parser Document produced by parsing
// its Oracle text. It is the parser-only counterpart of loweredFaceAbilities: it
// stops after parser.Parse and never compiles or lowers, so callers can measure
// parser-stage coverage independently of the compiler and lowering layers.
type FaceDocument struct {
	Name       string
	TypeLine   string
	OracleText string
	Document   parser.Document
}

// ParseCardFaces parses every executable face of a card with the same parser
// Context construction lowerFaceAbilities uses, but stops after parser.Parse: it
// runs no compiler or lowering. Faces with no supported card type or empty Oracle
// text yield a zero Document so the returned slice stays aligned with
// executableFaces.
func ParseCardFaces(card *ScryfallCard) []FaceDocument {
	faces := executableFaces(card)
	documents := make([]FaceDocument, 0, len(faces))
	for _, face := range faces {
		documents = append(documents, FaceDocument{
			Name:       face.Name,
			TypeLine:   face.TypeLine,
			OracleText: face.OracleText,
			Document:   parseFaceDocument(face),
		})
	}
	return documents
}

func parseFaceDocument(face scryfallFaceFields) parser.Document {
	parsedType := ParseTypeLine(face.TypeLine)
	if len(parsedType.Types) == 0 || face.OracleText == "" {
		return parser.Document{}
	}
	document, _ := parser.Parse(face.OracleText, parser.Context{
		InstantOrSorcery: slices.Contains(parsedType.Types, "Instant") || slices.Contains(parsedType.Types, "Sorcery"),
		Planeswalker:     slices.Contains(parsedType.Types, "Planeswalker"),
		Saga:             slices.Contains(parsedType.Subtypes, "Saga"),
		Leveler:          face.Layout == "leveler",
		CardName:         face.Name,
	})
	return document
}
