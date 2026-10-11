package cardlist

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGenerateUsesGoExportedIdentifiers(t *testing.T) {
	for _, tc := range []struct {
		name   string
		source string
		want   []string
	}{
		{
			name: "lazy cards",
			source: `package cards
import "github.com/natefinch/council4/mtg/game"
var AsciiCard = newAsciiCard
var ÉomerCard = newUnicodeCard
var ΩCard = newOtherUnicodeCard
var éomerPrivate = newUnicodeCard
var privateToken = newUnicodeCard()
var _Private = newUnicodeCard
var ÉWrongType = newWrongType
func newAsciiCard() *game.CardDef { return &game.CardDef{CardFace: game.CardFace{Name: "ASCII"}} }
func newUnicodeCard() *game.CardDef { return &game.CardDef{CardFace: game.CardFace{Name: "Éomer"}} }
func newOtherUnicodeCard() *game.CardDef { return &game.CardDef{CardFace: game.CardFace{Name: "Ω"}} }
func newWrongType() int { return 1 }
`,
			want: []string{"AsciiCard", "ÉomerCard", "ΩCard"},
		},
		{
			name: "eager tokens",
			source: `package cards
import "github.com/natefinch/council4/mtg/game"
var AsciiToken = newToken()
var ÉToken = newToken()
var éPrivateToken = newToken()
var privateToken = newToken()
var ÉWrongType = newWrongType()
func newToken() *game.CardDef { return &game.CardDef{CardFace: game.CardFace{Name: "Token"}} }
func newWrongType() int { return 1 }
`,
			want: []string{"AsciiToken", "ÉToken"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "definitions.go"), []byte(tc.source), 0o600); err != nil {
				t.Fatal(err)
			}
			source, err := Generate(dir, "cards", "test")
			if err != nil {
				t.Fatal(err)
			}
			file, err := parser.ParseFile(token.NewFileSet(), "cards.go", source, 0)
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			ast.Inspect(file, func(node ast.Node) bool {
				value, ok := node.(*ast.ValueSpec)
				if !ok || len(value.Names) != 1 || value.Names[0].Name != "Cards" {
					return true
				}
				list, ok := value.Values[0].(*ast.CompositeLit)
				if !ok {
					t.Fatal("Cards is not a registry literal")
				}
				for _, element := range list.Elts {
					if ident, ok := element.(*ast.Ident); ok {
						got = append(got, ident.Name)
						continue
					}
					entry, ok := element.(*ast.CompositeLit)
					if !ok {
						t.Fatal("unexpected registry entry")
					}
					for _, field := range entry.Elts {
						keyed, ok := field.(*ast.KeyValueExpr)
						if !ok {
							t.Fatal("unkeyed registry entry")
						}
						if key, ok := keyed.Key.(*ast.Ident); ok && key.Name == "New" {
							builder, ok := keyed.Value.(*ast.Ident)
							if !ok {
								t.Fatal("constructor is not an identifier")
							}
							got = append(got, builder.Name)
						}
					}
				}
				return false
			})
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("registry constructors=%v want=%v\n%s", got, tc.want, source)
			}
		})
	}
}
