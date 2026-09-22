#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
DATA_DIR="$PROJECT_DIR/data"
ES_PBF="$DATA_DIR/es-latest.osm.pbf"
GRAPH_FILE="$DATA_DIR/es-road.graph.gz"

mkdir -p "$DATA_DIR"

if [[ ! -s "$ES_PBF" ]]; then
  echo "Baixando extrato OSM do Espírito Santo..."
  curl --fail --location --continue-at - \
    'https://download.openstreetmap.fr/extracts/south-america/brazil/southeast/espirito-santo-latest.osm.pbf' \
    --output "$ES_PBF"
fi

echo "Validando o arquivo PBF..."
osmium fileinfo "$ES_PBF" >/dev/null

echo "Construindo o grafo viário..."
(
  cd "$PROJECT_DIR/backend"
  go run ./cmd/importer -input "$ES_PBF" -output "$GRAPH_FILE"
)

echo "Grafo pronto: $GRAPH_FILE"
