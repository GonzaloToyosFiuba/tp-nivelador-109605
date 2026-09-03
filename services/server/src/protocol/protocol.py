import lottery

import logger

MSG_BET = 1
MSG_END = 2
MSG_WINNERS = 3

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
        # agency_id
        agency_id = int.from_bytes(payload[offset:offset+AGENCY_SIZE], byteorder='big')
        offset += AGENCY_SIZE
    
        # first_name
        fn_len = payload[offset]
        offset += 1
        first_name = payload[offset:offset+fn_len].decode('utf-8')
        offset += fn_len
    
        # last_name
        ln_len = payload[offset]
        offset += 1
        last_name = payload[offset:offset+ln_len].decode('utf-8')
        offset += ln_len
    
        # DNI
        document = int.from_bytes(payload[offset:offset+DOC_SIZE], byteorder='big')
        offset += DOC_SIZE
    
        # birthdate
        birthdate = payload[offset:offset+BIRTH_SIZE].decode('utf-8')
        offset += BIRTH_SIZE
    
        # number
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

def serialize_winners(winners: list[lottery.Bet]) -> bytes:
    lines = []
    for bet in winners:
        line = f"{bet.first_name},{bet.last_name},{bet.document},{bet.birthdate},{bet.number}"
        lines.append(line)

    payload_str = "\n".join(lines)
    payload = payload_str.encode('utf-8')
    payload_len = len(payload)

    header = bytearray(HEADER_SIZE)
    header[0] = MSG_WINNERS
    header[1:5] = payload_len.to_bytes(4, byteorder='big')

    return bytes(header + payload)