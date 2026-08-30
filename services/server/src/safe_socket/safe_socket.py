import socket

# TODO: Complete with a short-read/short-write tolerant implementation

def send_all(socket: socket.socket, bytes):
    bytes_sent = 0
    size = len(bytes)

    while bytes_sent < size:
        n = socket.send(bytes[bytes_sent:])
        if n == 0:
            raise RuntimeError("La conexión con el socket se cerró durante el envío")
        
        bytes_sent += n

def recv_all(socket: socket.socket, size):
    buff = bytearray()

    while len(buff) < size:
        read = socket.recv(size - len(buff))
        if not read:
            if len(buff) == 0:
                return b''
            raise RuntimeError(
                f"Conexión cerrada abruptamente antes de completar el mensaje de {size} bytes (se leyeron {len(buff)})"
            )

        buff.extend(read)

    return bytes(buff)