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
API Gin → Dijkstra / A*
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

Backend:

```bash
cd backend
go run ./cmd
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

## API

- `GET /health`: disponibilidade do backend;
- `GET /api/v1/info`: origem e tamanho do grafo carregado;
- `POST /api/v1/route`: calcula uma rota com `dijkstra` ou `astar`.

Exemplo:

```json
{
  "origin": { "lat": -20.3155, "lng": -40.3128 },
  "destination": { "lat": -20.3297, "lng": -40.2925 },
  "algorithm": "astar"
}
```

A resposta contém distância, duração estimada, nós visitados, tempo do
algoritmo e todos os pontos da geometria percorrida.

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
