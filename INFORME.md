# Informe

> TOYOS, Gonzalo Ezequiel - 109605

## Trabajo práctico nivelador

El objetivo del siguiente informe es poder explicar los detalles de la resolución del trabajo, entre ellos el protocolo de comunicación utilizado  y el mecanismo de concurrencia.

### Protocolo

Se utiliza un procolo de tipo binario. La trama de cada mensaje está compuesta por un header de 5 bytes seguido de un payload. 

#### Header
El mismo consiste de 5 bytes, que se dividen de la siguiente manera:

* Primer byte: indicador del tipo de mensaje que se está enviando
* Siguientes 4 bytes: largo del payload

Los cuatro tipos de mensajes que pueden enviarse son:

```go
MsgBet     = 0x01 // Se envía un batch de apuestas

MsgEnd     = 0x02 // Fin de apuestas

MsgWinners = 0x03 // Respuesta del servidor con lista de ganadores

MsgAck     byte = 0x04 // Ack del server tras recibir un batch de apuestas
```

El `MsgBet` lo usa el cliente para indicar que va a mandar un lote de apuestas.

El `MsdEnd` lo usa el cliente para indicar al servidor que ya se enviaron todas las apuestas de una agencia.

El `MsgWinners` lo usa el servidor para devolverle al cliente la lista de ganadores.

El `MsgAck` lo usa el servidor para avisarle al cliente que terminó de recibir un lote completo de apuestas sin problemas.

#### Payload

La serialización de cada apuesta se hace de la siguiente manera:

| Campo | Tamaño |
| :--- | :--- |
| ID de la agencia | 4 bytes |
| Largo del primer nombre | 1 byte |
| Primer nombre | 1 byte por letra |
| Largo del apellido | 1 byte |
| Apellido | 1 byte por letra |
| DNI | 4 bytes |
| Fecha de nacimiento | 10 bytes |
| Apuesta | 4 bytes |

El envío de un batch de apuestas consiste en varias apuestas como las descritas anteriormente, más el header con el tipo de mensaje `MsgBet` y la longitud total de todas las apuestas del batch

El envío de ganadores es prácticamente igual al envío de apuestas, solo que sin enviar el agencyId porque a cada cliente se le envían solo sus ganadores. El header incluye el tipo de mensaje `MsgWinners` y la longitud total de todos los ganadores

| Campo | Tamaño |
| :--- | :--- |
| Largo del primer nombre | 1 byte |
| Primer nombre | 1 byte por letra |
| Largo del apellido | 1 byte |
| Apellido | 1 byte por letra |
| DNI | 4 bytes |
| Fecha de nacimiento | 10 bytes |
| Apuesta | 4 bytes |


### Manejo de concurrencia

#### Lottery Monitor

La clase `Lottery` no es *thread-safe* por sí sola. Cuando múltiples hilos intentan escribir en el archivo mediante `store_bets` o leerlo mediante `load_bets` de manera simultánea, ocurren condiciones de carrera.

Para resolver este problema, se aplicó el patrón de diseño Monitor a través de la clase wrapper `LotteryMonitor`.
Básicamente se añade un `Lock` que se toma en los 2 métodos que necesitan protección: `store_bets` y `load_bets`, y se suelta solo al terminarlos.

Para `load_bets` se usa un iterador y la expresión `yield from` para evitar cargar en memoria una lista completa pero a la vez poder retener el lock hasta la última iteración.

#### Barrera

El sorteo no puede llevarse a cabo hasta que se haya alcanzado una cantidad mínima de agencias participantes configurada en la variable de entorno `AGENCY_QUORUM_MIN`. Como las agencias finalizan el envío de sus apuestas en momentos distintos, se necesita sincronizar a todos los hilos para que ninguno calcule o responda los ganadores de forma prematura.

Para resolverlo se implementa una solución usando `threading.Barrier`. Cuando un hilo recibe un `MSG_END` no procesa a los ganadores, sino que hace un `barrier.wait()` quedando en espera. Una vez que el número de hilos esperando alcanza el valor de `AGENCY_QUORUM_MIN`, la barrera se libera simultáneamente para todos los hilos retenidos. Ahí ya pueden empezar a leer los ganadores y enviarle el resultado a su cliente.

### Limitaciones propias del lenguaje

El **Global Interpreter Lock (GIL)** es un mecanismo de exclusión mutua interno del intérprete de CPython. Su función es asegurar que solo un hilo de Python ejecute código nativo o bytecode del intérprete a la vez, incluso en procesadores con múltiples núcleos. Esto impide que los hilos aprovecho el paralelismo de la CPU para tareas de cómputo intensivo.

En este trabajo práctico no resulta un problema ya que no hay tareas con uso incentivo del CPU, sino que se tienen muchas transmisiones de datos por sockets, lecturas / escrituras en archivos en disco y además se usa un barrera como sincronización, que son todas tareas que no se ven afectadas por el GIL.