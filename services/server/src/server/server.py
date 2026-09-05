import os
import socket
import threading
import logger
import safe_socket
import lottery_monitor
import protocol

class Server:
    def __init__(self, server_host: str, server_port: int, storage_path: str) -> None:
        self.server_host = server_host
        self.server_port = server_port
        self.lottery_monitor = lottery_monitor.LotteryMonitor(storage_path)

        self.quorum_min = int(os.getenv("AGENCY_QUORUM_MIN", "1"))
        self.barrier = threading.Barrier(self.quorum_min)

    def _handle_client(self, client_socket):
        action = "handle-client"
        total_bets_count = 0
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
                    bets_batch = protocol.deserialize_bet_batch(payload)

                    if bets_batch:
                        self.lottery_monitor.store_bets(bets_batch)
                        total_bets_count += len(bets_batch)
                        if current_agency_id is None:
                            current_agency_id = bets_batch[0].agency_id

                elif msg_type == protocol.MSG_END:
                    logger.info("waiting-quorum", logger.LogResult.in_progress, "agency-id", current_agency_id)
                    self.barrier.wait()

                    winners = []
                    for bet in self.lottery_monitor.load_bets():
                        if bet.agency_id == current_agency_id and self.lottery_monitor.has_won(bet):
                            winners.append(bet)

                    response_msg = protocol.serialize_winners(winners)
                    safe_socket.send_all(client_socket, response_msg)

                    logger.info(
                        action,
                        logger.LogResult.success,
                        "agency-id",
                        current_agency_id,
                        "bets-amount",
                        total_bets_count,
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

                    client_thread = threading.Thread(
                        target=self._handle_client, 
                        args=(client_socket,)
                    )

                    client_thread.start()
                except Exception as e:
                    logger.error(action, logger.LogResult.fail)
                    raise e
