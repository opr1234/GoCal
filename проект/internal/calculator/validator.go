package calculator

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrInvalidCharacters  = errors.New("invalid characters in expression")
	ErrUnbalancedBrackets = errors.New("unbalanced brackets")
	ErrInvalidOperatorUse = errors.New("invalid operator use")
	ErrEmptyExpression    = errors.New("empty expression")
)

type Validator struct {
	allowedChars    *regexp.Regexp
	operatorPattern *regexp.Regexp
}

func NewValidator() *Validator {
	return &Validator{
		allowedChars: regexp.MustCompile(`^[0-9+\-*/^() .]+$`),
		operatorPattern: regexp.MustCompile(
			`(\d+(?:\.\d+)?|[-+*/^()]|(?:\s+))`,
		),
	}
}

func (v *Validator) Validate(expr string) error {
	if len(strings.TrimSpace(expr)) == 0 {
		return ErrEmptyExpression
	}

	if !v.allowedChars.MatchString(expr) {
		return ErrInvalidCharacters
	}

	if !v.checkBracketsBalance(expr) {
		return ErrUnbalancedBrackets
	}

	if err := v.checkOperatorUsage(expr); err != nil {
		return err
	}

	return nil
}

func (v *Validator) checkBracketsBalance(expr string) bool {
	balance := 0
	for _, char := range expr {
		switch char {
		case '(':
			balance++
		case ')':
			balance--
			if balance < 0 {
				return false
			}
		}
	}
	return balance == 0
}

func (v *Validator) checkOperatorUsage(expr string) error {
	tokens := v.operatorPattern.FindAllString(expr, -1)
	prevToken := ""

	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}

		if isOperator(token) {
			if isOperator(prevToken) && token != "-" {
				return ErrInvalidOperatorUse
			}

			if token == "-" && (prevToken == "" || isOperator(prevToken) || prevToken == "(") {
				return nil
			}
		}

		prevToken = token
	}

	if isOperator(prevToken) {
		return ErrInvalidOperatorUse
	}

	return nil
}

func isOperator(token string) bool {
	return strings.ContainsAny(token, "+-*/^")
}
