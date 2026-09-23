import type { Cost, LegAlternatives, RouteAlternative } from '../types'

export interface AltState {
  active: boolean
  step: number
  total: number
  legs: LegAlternatives[] | null
  selections: Record<number, number>
}

interface AltStepperProps {
  alt: AltState
  cost: Cost
  onSelectAlternative: (step: number, alternativeIndex: number) => void
  onConfirm: () => void
  onCancel: () => void
}

const CORES_OPCOES = ['#2563eb', '#7c3aed', '#059669']

function nomeDoTrecho(index: number, total: number, legs: LegAlternatives[] | null): string {
  if (legs === null || legs.length === 0) return ''
  if (total === 1) return 'Origem → Destino'
  if (index === 0) return 'Origem → Parada 1'
  if (index === total - 1) return `Parada ${total - 1} → Destino`
  return `Parada ${index} → Parada ${index + 1}`
}

function resumoAlternativa(alt: RouteAlternative, cost: Cost): string {
  const duracao = `${alt.estimated_duration_minutes.toFixed(1)} min`
  const distancia = `${alt.distance_km.toFixed(1)} km`
  return cost === 'distance' ? `${distancia} · ${duracao}` : `${duracao} · ${distancia}`
}

function AltStepper({ alt, cost, onSelectAlternative, onConfirm, onCancel }: AltStepperProps) {
  if (!alt.active || alt.legs === null) return null
  const leg = alt.legs[alt.step]
  if (leg === undefined) return null
  const selecionadas = Object.keys(alt.selections).length
  const todasSelecionadas = selecionadas >= alt.total

  return (
    <section className="alt-section">
      <div className="section-title">
        <label>Rotas alternativas</label>
        <button type="button" className="button button-small button-secondary" onClick={onCancel}>
          Cancelar
        </button>
      </div>
      <p className="hint">
        Trecho {alt.step + 1} de {alt.total} · {nomeDoTrecho(alt.step, alt.total, alt.legs)}
      </p>
      <ul className="alt-list">
        {leg.alternatives.map((alternativa, index) => {
          const escolhida = alt.selections[alt.step] === index
          return (
            <li key={index} className="alt-row">
              <button
                type="button"
                className={`alt-option ${escolhida ? 'alt-option-selected' : ''}`}
                onClick={() => onSelectAlternative(alt.step, index)}
                title="Usar esta alternativa neste trecho"
              >
                <span
                  className="alt-color-swatch"
                  style={{ backgroundColor: CORES_OPCOES[index % CORES_OPCOES.length] }}
                />
                <span className="alt-label">Rota {index + 1}</span>
                <span className="alt-meta">{resumoAlternativa(alternativa, cost)}</span>
              </button>
            </li>
          )
        })}
      </ul>
      {todasSelecionadas && (
        <button type="button" className="button button-primary button-block" onClick={onConfirm}>
          Usar estas rotas
        </button>
      )}
    </section>
  )
}

export default AltStepper