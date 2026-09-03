package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strings"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/domain"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

const (
	MsgBet     byte = 0x01 // Apuesta
	MsgEnd     byte = 0x02 // Fin de apuestas
	MsgWinners byte = 0x03 // Respuesta del servidor con lista de ganadores
)

const HeaderSize = 5 // 1 byte para el tipo, 4 para el largo del mensaje
const AgencySize = 4
const DocSize = 4
const BirthSize = 10
const NumSize = 4

func SerializeBet(agencyID uint32, bet *domain.Bet) []byte {
	firstNameBytes := []byte(bet.FirstName)
	lastNameBytes := []byte(bet.LastName)
	birthdateBytes := []byte(bet.Birthdate)

	payloadSize := AgencySize + 1 + len(firstNameBytes) + 1 + len(lastNameBytes) + DocSize + BirthSize + NumSize
	payload := make([]byte, payloadSize)

	offset := 0

	// Agency ID
	binary.BigEndian.PutUint32(payload[offset:offset+4], agencyID)
	offset += AgencySize

	// First name
	payload[offset] = byte(len(firstNameBytes))
	offset++
	copy(payload[offset:], firstNameBytes)
	offset += len(firstNameBytes)

	// Last name
	payload[offset] = byte(len(lastNameBytes))
	offset++
	copy(payload[offset:], lastNameBytes)
	offset += len(lastNameBytes)

	// DNI
	binary.BigEndian.PutUint32(payload[offset:offset+DocSize], uint32(bet.Document))
	offset += DocSize

	// BirthDate
	copy(payload[offset:], birthdateBytes)
	offset += BirthSize

	// Number
	binary.BigEndian.PutUint32(payload[offset:offset+NumSize], uint32(bet.Number))

	return payload
}

func SendBetBatch(conn net.Conn, agencyID uint32, bets []*domain.Bet) error {
	var payloadBuffer bytes.Buffer

	for _, bet := range bets {
		payloadBuffer.Write(SerializeBet(agencyID, bet))
	}

	payload := payloadBuffer.Bytes()
	payloadLen := uint32(len(payload))

	batchHeader := make([]byte, HeaderSize)
	batchHeader[0] = MsgBet
	binary.BigEndian.PutUint32(batchHeader[1:5], payloadLen)

	if err := safe_socket.SendAll(conn, batchHeader); err != nil {
		return err
	}

	if err := safe_socket.SendAll(conn, payload); err != nil {
		return err
	}

	return nil
}

func SendEnd(conn net.Conn) error {
	header := make([]byte, HeaderSize)
	header[0] = MsgEnd
	binary.BigEndian.PutUint32(header[1:5], 0)

	return safe_socket.SendAll(conn, header)
}

func ReceiveWinners(conn net.Conn) ([]string, error) {
	header, err := safe_socket.RecvAll(conn, HeaderSize)
	if err != nil {
		return nil, fmt.Errorf("error leyendo header de respuesta: %w", err)
	}

	msgType := header[0]
	if msgType != MsgWinners {
		return nil, fmt.Errorf("tipo de mensaje inesperado: %d", msgType)
	}

	payloadLen := binary.BigEndian.Uint32(header[1:5])
	if payloadLen == 0 {
		return []string{}, nil
	}

	payload, err := safe_socket.RecvAll(conn, int(payloadLen))
	if err != nil {
		return nil, fmt.Errorf("error leyendo ganadores: %w", err)
	}

	winnersStr := string(payload)
	if strings.TrimSpace(winnersStr) == "" {
		return []string{}, nil
	}

	return strings.Split(winnersStr, "\n"), nil
}
