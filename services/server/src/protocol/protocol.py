import lottery

import logger

MSG_BET = 1
MSG_END = 2
MSG_WINNERS = 3
MSG_ACK = 4

HEADER_SIZE = 5 # 1 byte para el tipo, 4 para el largo del mensaje
AGENCY_SIZE = 4
DOC_SIZE = 4
BIRTH_SIZE = 10
NUM_SIZE = 4

def deserialize_bet_batch(payload: bytes) -> list[lottery.Bet]:
    bets = []
    offset = 0
    total_len = len(payload)

    while offset < total_len:
        agency_id = int.from_bytes(payload[offset:offset+AGENCY_SIZE], byteorder='big')
        offset += AGENCY_SIZE
    
        fn_len = payload[offset]
        offset += 1
        first_name = payload[offset:offset+fn_len].decode('utf-8')
        offset += fn_len
    
        ln_len = payload[offset]
        offset += 1
        last_name = payload[offset:offset+ln_len].decode('utf-8')
        offset += ln_len
    
        document = int.from_bytes(payload[offset:offset+DOC_SIZE], byteorder='big')
        offset += DOC_SIZE
    
        birthdate = payload[offset:offset+BIRTH_SIZE].decode('utf-8')
        offset += BIRTH_SIZE
    
        number = int.from_bytes(payload[offset:offset+NUM_SIZE], byteorder='big')
        offset += NUM_SIZE
        

        bets.append(lottery.Bet(
            agency_id=agency_id,
            first_name=first_name,
            last_name=last_name,
            document=document,
            birthdate=birthdate,
            number=number,
        ))

    return bets

def serialize_ack() -> bytes:
    header = bytearray(HEADER_SIZE)
    header[0] = MSG_ACK
    header[1:5] = (0).to_bytes(4, byteorder='big')
    return bytes(header)

def serialize_winners(winners: list[lottery.Bet]) -> bytes:
    payload = bytearray()

    for bet in winners:
        fn_bytes = bet.first_name.encode('utf-8')
        payload.append(len(fn_bytes))
        payload.extend(fn_bytes)

        ln_bytes = bet.last_name.encode('utf-8')
        payload.append(len(ln_bytes))
        payload.extend(ln_bytes)

        payload.extend(bet.document.to_bytes(DOC_SIZE, byteorder='big'))

        payload.extend(bet.birthdate.encode('utf-8'))

        payload.extend(bet.number.to_bytes(NUM_SIZE, byteorder='big'))

    payload_len = len(payload)

    header = bytearray(HEADER_SIZE)
    header[0] = MSG_WINNERS
    header[1:5] = payload_len.to_bytes(4, byteorder='big')

    return bytes(header + payload)