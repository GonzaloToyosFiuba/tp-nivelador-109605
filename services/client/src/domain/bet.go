package domain

import (
	"fmt"
	"strconv"
	"strings"
)

type Bet struct {
	FirstName string
	LastName  string
	Document  int
	Birthdate string
	Number    int
}

// Recibe una línea CSV con los datos de una apuesta
// La parsea y la retorna como un objeto Bet
func ParseBetFromCSV(line string) (*Bet, error) {
	parts := strings.Split(strings.TrimSpace(line), ",")
	if len(parts) < 5 {
		return nil, fmt.Errorf("línea CSV inválida, faltan campos: %s", line)
	}

	doc, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("DNI inválido '%s': %w", parts[2], err)
	}

	num, err := strconv.Atoi(parts[4])
	if err != nil {
		return nil, fmt.Errorf("número de apuesta inválido '%s': %w", parts[4], err)
	}

	return &Bet{
		FirstName: parts[0],
		LastName:  parts[1],
		Document:  doc,
		Birthdate: parts[3],
		Number:    num,
	}, nil
}
