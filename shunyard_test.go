package main

import (
	"math/rand"
	"testing"

	l "github.com/Jose-Prince/UWULexer/lib"
)

func generateExpectedPostfix(r *rand.Rand) []l.RX_Token {
	expressionCount := r.Intn(2) + 1 // Minimum of 1 expressions
	postfixExpr := []l.RX_Token{}
	possibleChars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789,.-;:_¿?¡!'{}+*|\"#$%&/()=[]<>°¬")
	getRandomRune := func() rune {
		return possibleChars[r.Intn(len(possibleChars))]
	}
	getRandomTwoOp := func() l.Operator {
		switch r.Intn(2) {
		case 0:
			return l.OR
		default:
			return l.AND
		}
	}

	for i := range expressionCount {
		switch r.Intn(5) {
		default: // Simple two value expression
			a := l.CreateValueToken(getRandomRune())
			b := l.CreateValueToken(getRandomRune())
			op := l.CreateOperatorToken(getRandomTwoOp())

			postfixExpr = append(postfixExpr, a)
			postfixExpr = append(postfixExpr, b)
			postfixExpr = append(postfixExpr, op)
		}

		addOneOp := r.Intn(2) == 0
		if addOneOp {
			postfixExpr = append(postfixExpr, l.CreateOperatorToken(l.ZERO_OR_MANY))
		}

		if i > 0 {
			postfixExpr = append(postfixExpr, l.CreateOperatorToken(getRandomTwoOp()))
		}
	}

	return postfixExpr
}

func fromPostfixToInfix(postfix []l.RX_Token) []l.RX_Token {
	stack := l.Stack[[]l.RX_Token]{}

	for _, elem := range postfix {
		if elem.IsOperator() {
			op := elem.GetOperator()
			switch op {
			case l.OR, l.AND:
				b := stack.Pop()
				a := stack.Pop()

				combined := []l.RX_Token{l.CreateOperatorToken(l.LEFT_PAREN)}
				combined = append(combined, a.GetValue()...)
				combined = append(combined, elem)
				combined = append(combined, b.GetValue()...)
				combined = append(combined, l.CreateOperatorToken(l.RIGHT_PAREN))

				stack.Push(combined)

			case l.ZERO_OR_MANY, l.ONE_OR_MANY, l.OPTIONAL:
				a := stack.Pop()

				combined := []l.RX_Token{l.CreateOperatorToken(l.LEFT_PAREN)}
				combined = append(combined, a.GetValue()...)
				combined = append(combined, l.CreateOperatorToken(l.RIGHT_PAREN))

				combined = append(combined, elem)
				stack.Push(combined)
			default:
				panic("No brackets/parenthesis or set negation are allowed when the expression is postfix!")
			}

		} else {
			stack.Push([]l.RX_Token{elem})
		}
	}

	return stack.Pop().GetValue()
}

func FuzzInfixToPostfix(f *testing.F) {
	f.Add(int64(69420))
	f.Fuzz(func(t *testing.T, seed int64) {
		source := rand.NewSource(seed)
		random := rand.New(source)

		expected := generateExpectedPostfix(random)
		infixExpr := fromPostfixToInfix(expected)
		infixStr := fromTokenStreamToInfixString(infixExpr)

		result := DEFAULT_ALPHABET.ToPostfix(&infixExpr)
		compareTokensStreams(t, infixStr, expected, result)
	})
}

func TestDummyTokens(t *testing.T) {
	dummyCode := "Hello"
	expected := []l.RX_Token{
		l.CreateValueToken('a'),
		l.CreateValueToken('b'),
		l.CreateOperatorToken(l.OR),
		l.CreateDummyToken(l.DummyInfo{Code: dummyCode}),
		l.CreateOperatorToken(l.AND),
	}
	infix := []l.RX_Token{
		l.CreateValueToken('a'),
		l.CreateOperatorToken(l.OR),
		l.CreateValueToken('b'),
		l.CreateOperatorToken(l.AND),
		l.CreateDummyToken(l.DummyInfo{Code: dummyCode}),
	}
	result := DEFAULT_ALPHABET.ToPostfix(&infix)
	compareTokensStreams(t, "a|b (Dummy token)", expected, result)
}

func TestZeroOrManyOperator(t *testing.T) {
	expected := []l.RX_Token{
		l.CreateValueToken('a'),
		l.CreateValueToken('b'),
		l.CreateOperatorToken(l.OR),
		l.CreateValueToken('a'),
		l.CreateValueToken('b'),
		l.CreateOperatorToken(l.OR),
		l.CreateOperatorToken(l.ZERO_OR_MANY),
		l.CreateOperatorToken(l.AND),
	}
	infix := []l.RX_Token{
		l.CreateOperatorToken(l.LEFT_PAREN),
		l.CreateValueToken('a'),
		l.CreateOperatorToken(l.OR),
		l.CreateValueToken('b'),
		l.CreateOperatorToken(l.RIGHT_PAREN),
		l.CreateOperatorToken(l.ONE_OR_MANY),
	}
	result := DEFAULT_ALPHABET.ToPostfix(&infix)
	compareTokensStreams(t, "(a|b)+", expected, result)
}

func TestOptionalOperator(t *testing.T) {
	expected := []l.RX_Token{
		l.CreateValueToken('a'),
		l.CreateValueToken('b'),
		l.CreateOperatorToken(l.AND),
		l.CreateEpsilonToken(),
		l.CreateOperatorToken(l.OR),
	}
	infix := []l.RX_Token{
		l.CreateOperatorToken(l.LEFT_PAREN),
		l.CreateValueToken('a'),
		l.CreateOperatorToken(l.AND),
		l.CreateValueToken('b'),
		l.CreateOperatorToken(l.RIGHT_PAREN),
		l.CreateOperatorToken(l.OPTIONAL),
	}
	result := DEFAULT_ALPHABET.ToPostfix(&infix)
	compareTokensStreams(t, "(ab)?", expected, result)
}

// func TestFuzzFail(t *testing.T) {
// 	source := rand.NewSource(int64(69326))
// 	random := rand.New(source)
//
// 	expected := generateExpectedPostfix(random)
// 	infixExpr := fromPostfixToInfix(expected)
// 	infixStr := fromTokenStreamToInfixString(infixExpr)
//
// 	result := DEFAULT_ALPHABET.ToPostfix(&infixExpr)
// 	compareTokensStreams(t, infixStr, expected, result)
// }
