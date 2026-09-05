import sys

def generate_docker_compose(num_clients, batch_size=50):
    base_compose = f"""services:
  server:
    build:
      context: ./services/server
      dockerfile: Dockerfile
    container_name: server
    ports:
      - "5678:5678"
    environment:
      - PYTHONUNBUFFERED=1
      - SERVER_HOST=server
      - SERVER_PORT=5678
      - AGENCY_QUORUM_MIN=3
"""

    client_template = """
  client_{id}:
    build:
      context: ./services/client
      dockerfile: Dockerfile
    container_name: client_{id}
    depends_on:
      - server
    volumes:
      - ./input:/input:ro
      - ./output:/output
    environment:
      - AGENCY_ID={id}
      - SERVER_HOST=server
      - SERVER_PORT=5678
      - INPUT_FILE=/input/input-{id}.csv
      - OUTPUT_FILE=/output/output-{id}.csv
      - BATCH_SIZE={batch_size}
"""

    clients_yaml = "".join(client_template.format(id=i, batch_size=batch_size) for i in range(num_clients))
    
    with open("docker-compose.yaml", "w") as f:
        f.write(base_compose + clients_yaml)

    print(
        f"'docker-compose.yaml' generado con {num_clients} clientes "
        f"y BATCH_SIZE={batch_size} correctamente."
    )

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Error, usar: python3 generate_compose.py <cantidad_de_clientes> [batch_size]")
        sys.exit(1)
        
    try:
        clients_count = int(sys.argv[1])
        batch_size = int(sys.argv[2]) if len(sys.argv) > 2 else 50
        generate_docker_compose(clients_count, batch_size)
    except ValueError:
        print("Ingresar un número entero válido")