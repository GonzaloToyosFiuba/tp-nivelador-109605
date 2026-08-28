import sys

def generate_docker_compose(num_clients):
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
"""

    client_template = """
  client_{id}:
    build:
      context: ./services/client
      dockerfile: Dockerfile
    container_name: client_{id}
    depends_on:
      - server
    environment:
      - AGENCY_ID={id}
      - SERVER_HOST=server
      - SERVER_PORT=5678
"""

    clients_yaml = "".join(client_template.format(id=i) for i in range(num_clients))
    
    with open("docker-compose.yaml", "w") as f:
        f.write(base_compose + clients_yaml)

    print(f"'docker-compose.yaml' generado con {num_clients} clientes correctamente.")

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Error, usar: python3 generate_compose.py <cantidad_de_clientes>")
        sys.exit(1)
        
    try:
        clients_count = int(sys.argv[1])
        generate_docker_compose(clients_count)
    except ValueError:
        print("Ingresar un número entero válido")