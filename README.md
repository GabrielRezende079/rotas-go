# Routes Go

Esse é um projeto academico que visa utilizar a **Teoria dos Grafos** e algoritimos **Dijkstra** para encontrar a melhor rota entre um ponto de partida e destino no mapa.

### 🗺️ Visão geral do projeto

A arquitetura poderia ser:

```
                    ┌──────────────────────┐
                    │      Frontend        │
                    │ React + TypeScript   │
                    │      + Leaflet       │
                    └──────────┬───────────┘
                               │
                         HTTP / REST
                               │
                               ▼
                    ┌──────────────────────┐
                    │      API Go          │
                    │       Gin            │
                    ├──────────────────────┤
                    │ Route Controller     │
                    │ Route Service        │
                    │ Graph                │
                    │ Dijkstra / A*        │
                    └──────────┬───────────┘
                               │
                    ┌──────────┴───────────┐
                    ▼                      ▼
             ┌─────────────┐       ┌──────────────┐
             │ PostgreSQL  │       │ OpenStreetMap│
             │ + PostGIS   │       │     dados    │
             └─────────────┘       └──────────────┘
```

### 🏗️ Estrutura do projeto Go

```
route-optimizer/
│
├── cmd/
│   └── api/
│       └── main.go
│
├── internal/
│   ├── graph/
│   │   ├── graph.go
│   │   ├── node.go
│   │   ├── edge.go
│   │   ├── dijkstra.go
│   │   └── astar.go
│   │
│   ├── route/
│   │   ├── controller.go
│   │   ├── service.go
│   │   └── dto.go
│   │
│   └── infrastructure/
│       └── database/
│
├── migrations/
│
├── go.mod
└── README.md
```

### Bibliotecas

**Go**

* gin-gonic/gin — API HTTP
* biblioteca de heap da própria stdlib (container/heap) — prioridade do Dijkstra

**Frontend**

* React
* TypeScript
* Leaflet
* React-Leaflet

**Infra**

* PostgreSQL
* PostGIS
* Docker Compose

**Dados**

* OpenStreetMap