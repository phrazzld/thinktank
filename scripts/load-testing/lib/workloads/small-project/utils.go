package main

import (
	"fmt"
	"math"
)

// MathUtils provides additional mathematical utilities
type MathUtils struct{}

// Power calculates base^exponent
func (m *MathUtils) Power(base, exponent float64) float64 {
	return math.Pow(base, exponent)
}

// SquareRoot calculates square root
func (m *MathUtils) SquareRoot(value float64) (float64, error) {
	if value < 0 {
		return 0, fmt.Errorf("cannot calculate square root of negative number")
	}
	return math.Sqrt(value), nil
}

// Factorial calculates factorial of a non-negative integer
func (m *MathUtils) Factorial(n int) (int64, error) {
	if n < 0 {
		return 0, fmt.Errorf("factorial is not defined for negative numbers")
	}
	if n == 0 || n == 1 {
		return 1, nil
	}

	result := int64(1)
	for i := 2; i <= n; i++ {
		result *= int64(i)
	}
	return result, nil
}

// IsPrime checks if a number is prime
func (m *MathUtils) IsPrime(n int) bool {
	if n < 2 {
		return false
	}
	if n == 2 {
		return true
	}
	if n%2 == 0 {
		return false
	}

	limit := int(math.Sqrt(float64(n)))
	for i := 3; i <= limit; i += 2 {
		if n%i == 0 {
			return false
		}
	}
	return true
}

// GCD calculates greatest common divisor using Euclidean algorithm
func (m *MathUtils) GCD(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// LCM calculates least common multiple
func (m *MathUtils) LCM(a, b int) int {
	return (a * b) / m.GCD(a, b)
}
