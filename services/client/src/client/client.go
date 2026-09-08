package client

import (
	"bufio"
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

type ClientConfig struct {
	ServerHost string
	ServerPort string
	AgencyId   string
	InputFile  string
	OutputFile string
	BatchSize  string
}

type Client struct {
	conn    net.Conn
	config  ClientConfig
	sigterm bool
}

func NewClient(config ClientConfig) (*Client, error) {
	conn, err := connectToServer(config.ServerHost, config.ServerPort)
	if err != nil {
		logger.Warn("connect-to-server", logger.Fail)
		return nil, err
	}

	client := &Client{conn: conn, config: config, sigterm: false}
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

func (client *Client) Safe_close() {
	client.sigterm = true
	client.conn.Close()
}

func (client *Client) Run() error {
	const mainAction = "process-bets"
	defer client.conn.Close()

	agencyIdNum, err := strconv.Atoi(client.config.AgencyId)
	if err != nil {
		logger.Error("invalid-agency-id", logger.Fail, "agency-id", client.config.AgencyId)
		return err
	}
	agencyID := uint32(agencyIdNum)

	inputFile, err := os.Open(client.config.InputFile)
	if err != nil {
		logger.Error("open-input-file", logger.Fail, "err", err)
		return err
	}
	defer inputFile.Close()

	batchSize, err := strconv.Atoi(client.config.BatchSize)
	if err != nil || batchSize <= 0 {
		logger.Error("invalid-batch-size", logger.Fail, "batch-size", client.config.BatchSize)
		return err
	}

	if err := client.sendBets(inputFile, agencyID, batchSize); err != nil {
		if client.sigterm {
			return nil
		}
		logger.Error("send-bets-fail", logger.Fail, "agency-id", client.config.AgencyId)
		return err
	}

	if err := protocol.SendEnd(client.conn); err != nil {
		if client.sigterm {
			return nil
		}
		logger.Error("send-end-fail", logger.Fail, "agency-id", client.config.AgencyId)
		return err
	}

	winners, err := protocol.ReceiveWinners(client.conn)
	if err != nil {
		if client.sigterm {
			return nil
		}
		logger.Error("recv-winners-fail", logger.Fail, "agency-id", client.config.AgencyId)
		return err
	}

	if err := client.saveWinners(winners); err != nil {
		return err
	}

	logger.Info(mainAction, logger.Success, "agency-id", client.config.AgencyId)
	return nil
}

// Lee las apuestas desde el archivo inputFile y las procesa de a batches,
// con tamaño configurable por parámetro
// Tras procesar un batch de apuestas, las envía según el protocolo
// Y espera el ACK por parte del servidor antes de continuar con el siguiente batch
func (client *Client) sendBets(inputFile *os.File, agencyID uint32, batchSize int) error {
	batch := make([]*domain.Bet, 0, batchSize)

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

		batch = append(batch, bet)

		if len(batch) >= batchSize {
			if err := protocol.SendBetBatch(client.conn, agencyID, batch); err != nil {
				logger.Error("send-bet-batch-fail", logger.Fail, "agency-id", client.config.AgencyId, "err", err)
				return err
			}
			if err := protocol.WaitBatchAck(client.conn); err != nil {
				logger.Error("wait-batch-ack-fail", logger.Fail, "agency-id", client.config.AgencyId, "err", err)
				return err
			}
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		if err := protocol.SendBetBatch(client.conn, agencyID, batch); err != nil {
			logger.Error("send-bet-batch-remanent-fail", logger.Fail, "agency-id", client.config.AgencyId, "err", err)
			return err
		}
		if err := protocol.WaitBatchAck(client.conn); err != nil {
			logger.Error("wait-batch-ack-fail", logger.Fail, "agency-id", client.config.AgencyId, "err", err)
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Error("read-input-file", logger.Fail, "agency-id", client.config.AgencyId)
		return err
	}

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
