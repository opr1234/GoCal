package calculator

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrInvalidCharacter  = errors.New("invalid character in expression")
	ErrDivisionByZero    = errors.New("division by zero")
	ErrTimeout           = errors.New("calculation timeout")
	ErrInvalidExpression = errors.New("invalid expression structure")
)

type Evaluator struct {
	operators map[string]int
}

func NewEvaluator() *Evaluator {
	return &Evaluator{
		operators: map[string]int{
			"+": 1,
			"-": 1,
			"*": 2,
			"/": 2,
			"^": 3,
		},
	}
}

func (e *Evaluator) Validate(expr string) error {
	allowed := "0123456789+-*/^() ."
	for _, c := range expr {
		if !strings.ContainsRune(allowed, c) {
			return fmt.Errorf("%w: '%c'", ErrInvalidCharacter, c)
		}
	}
	return nil
}

func (e *Evaluator) Evaluate(ctx context.Context, expr string) (float64, error) {
	select {
	case <-ctx.Done():
		return 0, ErrTimeout
	default:
	}

	if err := e.Validate(expr); err != nil {
		return 0, err
	}

	tokens, err := e.tokenize(expr)
	if err != nil {
		return 0, err
	}

	postfix, err := e.infixToPostfix(tokens)
	if err != nil {
		return 0, err
	}

	return e.evaluatePostfix(postfix)
}

func (e *Evaluator) tokenize(expr string) ([]string, error) {
	var tokens []string
	var numberBuffer strings.Builder
	expr = strings.ReplaceAll(expr, " ", "")

	for i, char := range expr {
		switch {
		case char >= '0' && char <= '9' || char == '.':
			numberBuffer.WriteRune(char)
		case char == '-' && (i == 0 || expr[i-1] == '('):
			numberBuffer.WriteRune(char)
		default:
			if numberBuffer.Len() > 0 {
				tokens = append(tokens, numberBuffer.String())
				numberBuffer.Reset()
			}
			tokens = append(tokens, string(char))
		}
	}

	if numberBuffer.Len() > 0 {
		tokens = append(tokens, numberBuffer.String())
	}

	return tokens, nil
}

func (e *Evaluator) infixToPostfix(tokens []string) ([]string, error) {
	var output []string
	var stack []string

	for _, token := range tokens {
		if _, err := strconv.ParseFloat(token, 64); err == nil {
			output = append(output, token)
		} else if token == "(" {
			stack = append(stack, token)
		} else if token == ")" {
			for len(stack) > 0 && stack[len(stack)-1] != "(" {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 {
				return nil, ErrInvalidExpression
			}
			stack = stack[:len(stack)-1]
		} else {
			for len(stack) > 0 && e.operators[stack[len(stack)-1]] >= e.operators[token] && stack[len(stack)-1] != "(" {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, token)
		}
	}

	for len(stack) > 0 {
		if stack[len(stack)-1] == "(" {
			return nil, ErrInvalidExpression
		}
		output = append(output, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}

	return output, nil
}

func (e *Evaluator) evaluatePostfix(postfix []string) (float64, error) {
	stack := []float64{}

	for _, token := range postfix {
		if num, err := strconv.ParseFloat(token, 64); err == nil {
			stack = append(stack, num)
		} else {
			if len(stack) < 2 {
				return 0, ErrInvalidExpression
			}

			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			var res float64
			switch token {
			case "+":
				res = a + b
			case "-":
				res = a - b
			case "*":
				res = a * b
			case "/":
				if b == 0 {
					return 0, ErrDivisionByZero
				}
				res = a / b
			case "^":
				res = 1
				for i := 0; i < int(b); i++ {
					res *= a
				}
			default:
				return 0, fmt.Errorf("unknown operator: %s", token)
			}

			stack = append(stack, res)
		}
	}

	if len(stack) != 1 {
		return 0, ErrInvalidExpression
	}

	return stack[0], nil
}
