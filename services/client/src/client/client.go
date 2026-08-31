package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/domain"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/protocol"
)

const CONNECTION_ATTEMPTS_MAX = 3
const CONNECTION_ATTEMPS_DELAY_MS = 200

// const ECHO_CLIENT_BUFFER_SIZE = 512
// const ECHO_CLIENT_MESSAGE_AMOUNT = 3
// const ECHO_CLIENT_MESSAGE_DELAY_MS = 1000

type ClientConfig struct {
	ServerHost string
	ServerPort string
	AgencyId   string
	InputFile  string
	OutputFile string
}

type Client struct {
	conn   net.Conn
	config ClientConfig
}

func NewClient(config ClientConfig) (*Client, error) {
	conn, err := connectToServer(config.ServerHost, config.ServerPort)
	if err != nil {
		logger.Warn("connect-to-server", logger.Fail)
		return nil, err
	}

	client := &Client{conn: conn, config: config}
	return client, nil
}

func connectToServer(host, port string) (net.Conn, error) {
	const action = "connect-to-server"
	var err error
	var conn net.Conn

	logger.Info(action, logger.InProgress)
	for i := range CONNECTION_ATTEMPTS_MAX {
		conn, err = net.Dial("tcp", host+":"+port)
		if err != nil {
			logger.Warn(action, logger.Fail, "attempt", i)
			time.Sleep(CONNECTION_ATTEMPS_DELAY_MS * time.Millisecond)
			continue
		}

		logger.Info(action, logger.Success)
		break
	}

	return conn, err
}

func (client *Client) Run() error {
	const mainAction = "process-bets"
	defer client.conn.Close()

	agencyIdNum, err := strconv.Atoi(client.config.AgencyId)
	if err != nil {
		logger.Error("invalid-agency-id", logger.Fail, "agency-id", client.config.AgencyId)
		return fmt.Errorf("agency id inválido: %w", err)
	}
	agencyID := uint32(agencyIdNum)

	inputFile, err := os.Open(client.config.InputFile)
	if err != nil {
		logger.Error("open-input-file", logger.Fail, "err", err)
		return err
	}
	defer inputFile.Close()

	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		bet, err := domain.ParseBetFromCSV(line)
		if err != nil {
			logger.Warn("parse-bet-fail", logger.Fail, "agency-id", client.config.AgencyId, "line", line)
			continue
		}

		if err := protocol.SendBet(client.conn, agencyID, bet); err != nil {
			logger.Error("send-bet-fail", logger.Fail, "agency-id", client.config.AgencyId)
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Error("read-input-file", logger.Fail, "agency-id", client.config.AgencyId)
		return err
	}

	if err := protocol.SendEnd(client.conn); err != nil {
		logger.Error("send-end-fail", logger.Fail, "agency-id", client.config.AgencyId)
		return err
	}

	winners, err := protocol.ReceiveWinners(client.conn)
	if err != nil {
		logger.Error("recv-winners-fail", logger.Fail, "agency-id", client.config.AgencyId)
		return err
	}

	if err := client.saveWinners(winners); err != nil {
		return err
	}

	logger.Info(mainAction, logger.Success, "agency-id", client.config.AgencyId)
	return nil
}

// Crea un arhivo con el nombre establecido en Client.ClientConfig.OutpuFile
// En el mismo guarda los ganadores de la apuesta recibidos desde el servidor
// Cada línea se envía como estaba originalmente
func (client *Client) saveWinners(winners []string) error {
	outputFile, err := os.Create(client.config.OutputFile)
	if err != nil {
		logger.Error("create-output-file", logger.Fail, "path", client.config.OutputFile)
		return err
	}
	defer outputFile.Close()

	writer := bufio.NewWriter(outputFile)
	for _, winner := range winners {
		if strings.TrimSpace(winner) == "" {
			continue
		}
		if _, err := writer.WriteString(winner + "\n"); err != nil {
			logger.Error("write-output-fail", logger.Fail, "agency-id", client.config.AgencyId)
			return err
		}
	}

	if err := writer.Flush(); err != nil {
		logger.Error("flush-output-fail", logger.Fail, "agency-id", client.config.AgencyId)
		return err
	}

	return nil
}
