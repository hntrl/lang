package nodes

import (
	"bufio"
	"strings"

	"github.com/hntrl/lang/language/parser"
	"github.com/hntrl/lang/language/tokens"

	"github.com/go-test/deep"
	"github.com/pkg/errors"
)

type TestFixture struct {
	lit          string
	parseFn      func(p *parser.Parser) (Node, error)
	expects      Node
	expectsError error
	endingToken  tokens.Token
}

var errHandler = func(pos tokens.Position, msg string) {}

func evaluateTest(test TestFixture) error {
	lexer := parser.NewLexer(bufio.NewReader(strings.NewReader(test.lit)), errHandler)
	parser := parser.NewParser(lexer)

	node, err := test.parseFn(parser)
	if diff := deep.Equal(err, test.expectsError); diff != nil {
		return errors.Errorf("Expected error to be %v, but got %v", test.expectsError, err)
	}
	if test.expects != nil {
		if diff := deep.Equal(node, test.expects); diff != nil {
			return errors.New(strings.Join(diff, "\n"))
		}
	}

	if test.expectsError == nil {
		_, tok, _ := parser.Scan()
		if tok != test.endingToken {
			return errors.Errorf("Expected %v, got %v", test.endingToken, tok)
		}
	}
	return nil
}
