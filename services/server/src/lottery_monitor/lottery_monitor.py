from collections.abc import Iterator
import threading

import lottery

class LotteryMonitor:
    def __init__(self, storage_path: str) -> None:
        self.lottery = lottery.Lottery(storage_path)
        self.lock = threading.Lock()

    def has_won(self, bet: lottery.Bet) -> bool:
        return self.lottery.has_won(bet)

    # Provee acceso thread safe al método store_bets de Lottery
    def store_bets(self, bets: list[lottery.Bet]) -> None:
        with self.lock:
            self.lottery.store_bets(bets)

    # Provee acceso thread safe al método load_bets de Lottery
    def load_bets(self) -> Iterator[lottery.Bet]:
        with self.lock:
            yield from self.lottery.load_bets()
