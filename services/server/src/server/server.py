import socket
import logger
import safe_socket
import lottery
import protocol

class Server:
    def __init__(self, server_host: str, server_port: int, storage_path: str) -> None:
        self.server_host = server_host
        self.server_port = server_port
        self.lottery = lottery.Lottery(storage_path)

    def _handle_client(self, client_socket):
        action = "handle-client"
        agency_bets = []
        current_agency_id = None
        
        try:
            logger.info(action, logger.LogResult.in_progress)
            while True:
                header = safe_socket.recv_all(client_socket, protocol.HEADER_SIZE)
                if not header:
                    break

                msg_type = header[0]
                payload_len = int.from_bytes(header[1:5], byteorder='big')

                if msg_type == protocol.MSG_BET:
                    payload = safe_socket.recv_all(client_socket, payload_len)
                    bet = protocol.deserialize_bet(payload)
                    agency_bets.append(bet)
                    current_agency_id = bet.agency_id

                elif msg_type == protocol.MSG_END:
                    if agency_bets:
                        self.lottery.store_bets(agency_bets)

                    winners = []
                    for bet in self.lottery.load_bets():
                        if bet.agency_id == current_agency_id and self.lottery.has_won(bet):
                            winners.append(bet)

                    response_msg = protocol.serialize_winners(winners)
                    safe_socket.send_all(client_socket, response_msg)

                    logger.info(
                        action,
                        logger.LogResult.success,
                        "agency-id",
                        current_agency_id,
                        "bets-amount",
                        len(agency_bets),
                        "winners-amount",
                        len(winners),
                    )
                    return
        except Exception as e:
            logger.error(
                action,
                logger.LogResult.fail,
                "agency-id",
                current_agency_id,
            )
            raise e
        
        finally:
            client_socket.close()

    def run(self):
        action = "accept-connection"
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as server_socket:
            server_socket.bind((self.server_host, self.server_port))
            server_socket.listen()
            while True:
                try:
                    logger.info(action, logger.LogResult.in_progress)
                    client_socket, _ = server_socket.accept()
                except Exception as e:
                    logger.error(action, logger.LogResult.fail)
                    raise e
                logger.info(action, logger.LogResult.success)

                self._handle_client(client_socket)
