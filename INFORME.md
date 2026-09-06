Redactar un breve informe en donde se detallen los aspectos más importantes de la solución provista, como ser el protocolo de comunicación implementado y los mecanismos para sincronizar la ejecución concurrente.

# Informe

## Trabajo práctico nivelador

El objetivo del siguiente informe es poder explicar los detalles de la resolución del trabajo, entre ellos el protocolo de comunicación utilizado  y el mecanismo de concurrencia.

### Protocolo

Se utiliza un procolo de tipo binario. La trama de cada mensaje está compuesta por un header de 5 bytes seguido de un payload. 

#### Header
El mismo consiste de 5 bytes, que se dividen de la siguiente manera:

* Primer byte: indicador del tipo de mensaje que se está enviando
* Siguientes 4 bytes: largo del payload

Los tres tipos de mensajes que pueden enviarse son:

```go
MsgBet     = 0x01 // Se envía un batch de apuestas

MsgEnd     = 0x02 // Fin de apuestas

MsgWinners = 0x03 // Respuesta del servidor con lista de ganadores
```

El `MsgBet` lo usa el cliente para indicar que va a mandar un lote de apuestas.

El `MsdEnd` lo usa el cliente para indicar al servidor que ya se enviaron todas las apuestas de una agencia.

El `MsgWinners` lo usa el servidor para devolverle al cliente la lista de ganadores.

#### Payload

La serialización de cada apuesta se hace de la siguiente manera:

* agencyId - 4 bytes
* Largo del primer nombre - 1 byte
* Primer nombre - 1 byte por letra
* Largo del apellido - 1 byte
* Apellido - 1 byte por letra
* DNI - 4 bytes
* Fecha de nacimiento - 10 bytes
* Apuesta - 4 bytes