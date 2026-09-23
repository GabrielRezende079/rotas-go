# Rotas Go

Projeto acadêmico de Teoria dos Grafos que calcula rotas sobre a malha viária
real do Espírito Santo. O backend constrói o próprio grafo a partir de dados do
OpenStreetMap e executa implementações próprias de Dijkstra e A*.

O OpenStreetMap fornece os nós e as vias. Ele não calcula a rota: essa parte é
feita pelo código Go deste repositório.

## Arquitetura

```text
OpenStreetMap PBF
        ↓
importador Go (paulmach/osm)
        ↓
grafo direcionado compactado
        ↓
API Gin → Dijkstra / A*  →  PostgreSQL (rotas salvas)
        ↓
React + Leaflet
```

Cada nó OSM usado por uma via dirigível vira um vértice. Cada segmento entre
nós consecutivos vira uma aresta. Os pesos de distância são calculados sobre a
geometria real da via; a duração estimada usa `maxspeed` quando disponível e
uma velocidade padrão por classe da via nos demais casos.

## Preparar os dados reais

Requisitos no macOS:

```bash
brew install go osmium-tool pkgconf
```

Depois execute, na raiz do projeto:

```bash
./scripts/prepare_es_data.sh
```

O script:

1. baixa o extrato estadual publicado por OpenStreetMap France;
2. valida o PBF com Osmium;
3. compila as vias em `data/es-road.graph.gz`.

Os dados grandes ficam fora do Git e podem ser regenerados a qualquer momento.

## Executar

Requisitos do banco (opcional): um PostgreSQL. As rotas salvas dependem dele; sem
`DATABASE_URL` a API roda normalmente, apenas sem a persistência.

```bash
# PostgreSQL local via Docker (opcional)
docker run -d --name rotas-go-pg -e POSTGRES_PASSWORD=rotas -e POSTGRES_DB=rotas \
  -e POSTGRES_USER=rotas -p 5432:5432 postgres:16-alpine
```

Backend (as migrations são aplicadas sozinhas no `startup`):

```bash
cd backend
DATABASE_URL=postgres://rotas:rotas@localhost:5432/rotas go run ./cmd
```

Frontend:

```bash
cd frontend/rotas-web
npm install
npm run dev
```

Abra `http://127.0.0.1:5173`.

Se `data/es-road.graph.gz` não existir, a API inicia com o pequeno grafo
didático original e exibe essa condição na interface. Para apontar a API para
outro cache:

```bash
ROTAS_GRAPH_FILE=/caminho/grafo.gz go run ./cmd
```

### Variáveis de ambiente

| Variável | Padrão | Descrição |
| --- | --- | --- |
| `ROTAS_GRAPH_FILE` | `../data/es-road.graph.gz` | cache do grafo viário |
| `DATABASE_URL` | vazio | string de conexão PostgreSQL; se ausente, rotas salvas ficam desabilitadas |
| `PORT` | `8080` | porta do servidor HTTP |

## API

- `GET /health`: disponibilidade do backend;
- `GET /api/v1/info`: origem e tamanho do grafo carregado;
- `POST /api/v1/route`: calcula uma rota com `dijkstra` ou `astar`, incluindo
  pontos de passagem (`waypoints`) opcionais entre origem e destino;
- `POST /api/v1/routes`: calcula a rota de vários veículos de uma vez
  (requisição em lote, executada em paralelo);
- `POST /api/v1/routes/saved`: salva um lote de rotas calculado;
- `GET /api/v1/routes/saved`: lista os resumos das rotas salvas;
- `GET /api/v1/routes/saved/:id`: busca uma rota salva com o snapshot completo;
- `DELETE /api/v1/routes/saved/:id`: exclui uma rota salva.

Rota com pontos de passagem:

```json
{
  "origin": { "lat": -20.3155, "lng": -40.3128 },
  "destination": { "lat": -20.3297, "lng": -40.2925 },
  "waypoints": [{ "lat": -20.2635, "lng": -40.4166 }],
  "algorithm": "astar"
}
```

A resposta contém distância, duração estimada, nós visitados, tempo do
algoritmo, todos os pontos da geometria percorrida e a quebra por trecho
(`legs`), além dos pontos ajustados (`origin`, `waypoints`, `destination`).

Lote de veículos:

```json
{
  "algorithm": "dijkstra",
  "vehicles": [
    {
      "id": "v1",
      "origin": { "lat": -20.3155, "lng": -40.3128 },
      "destination": { "lat": -20.3297, "lng": -40.2925 }
    },
    {
      "id": "v2",
      "origin": { "lat": -19.539, "lng": -40.63 },
      "destination": { "lat": -19.93, "lng": -40.407 },
      "waypoints": [{ "lat": -20.1288, "lng": -40.3078 }]
    }
  ]
}
```

A resposta traz uma entrada `routes` com a rota calculada de cada veículo,
identificada pelo mesmo `id` enviado na requisição.

## Regras consideradas

- vias adequadas para automóveis;
- direção de `oneway`;
- sentido de rotatórias;
- restrições básicas de acesso;
- velocidade máxima informada no OSM;
- velocidade padrão por categoria quando `maxspeed` não está disponível.

As relações OSM de restrição de conversão ainda não são processadas. Essa
limitação está explícita para não apresentar o modelo acadêmico como um sistema
de navegação completo.

## Fontes e licença

- Dados viários: OpenStreetMap contributors, licença ODbL;
- extrato PBF: OpenStreetMap France.

A atribuição do OpenStreetMap também permanece visível no mapa.
