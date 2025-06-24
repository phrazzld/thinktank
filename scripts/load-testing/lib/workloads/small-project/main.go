package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

// Calculator provides basic arithmetic operations
type Calculator struct {
	result float64
}

// NewCalculator creates a new calculator instance
func NewCalculator() *Calculator {
	return &Calculator{result: 0}
}

// Add performs addition
func (c *Calculator) Add(value float64) *Calculator {
	c.result += value
	return c
}

// Subtract performs subtraction
func (c *Calculator) Subtract(value float64) *Calculator {
	c.result -= value
	return c
}

// Multiply performs multiplication
func (c *Calculator) Multiply(value float64) *Calculator {
	c.result *= value
	return c
}

// Divide performs division with zero check
func (c *Calculator) Divide(value float64) (*Calculator, error) {
	if value == 0 {
		return nil, fmt.Errorf("division by zero")
	}
	c.result /= value
	return c, nil
}

// Result returns the current result
func (c *Calculator) Result() float64 {
	return c.result
}

// Reset clears the calculator
func (c *Calculator) Reset() {
	c.result = 0
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: calculator <num1> <operation> <num2>")
		fmt.Println("Operations: +, -, *, /")
		os.Exit(1)
	}

	num1, err := strconv.ParseFloat(os.Args[1], 64)
	if err != nil {
		log.Fatal("Invalid first number:", err)
	}

	operation := os.Args[2]

	num2, err := strconv.ParseFloat(os.Args[3], 64)
	if err != nil {
		log.Fatal("Invalid second number:", err)
	}

	calc := NewCalculator()
	calc.Add(num1)

	switch operation {
	case "+":
		calc.Add(num2)
	case "-":
		calc.Subtract(num2)
	case "*":
		calc.Multiply(num2)
	case "/":
		_, err := calc.Divide(num2)
		if err != nil {
			log.Fatal("Error:", err)
		}
	default:
		log.Fatal("Unknown operation:", operation)
	}

	fmt.Printf("Result: %.2f\n", calc.Result())
}
