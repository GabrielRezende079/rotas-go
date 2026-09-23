import type { ReactNode, UIEvent } from 'react'

import type { ListaPagina } from '../lista'
import { Icon } from './Icon'
import { Spinner } from './Spinner'

interface ListaColapsavelProps<T> {
  titulo: string
  dados: ListaPagina<T>
  aberto: boolean
  onToggle: () => void
  acao?: ReactNode
  renderItem: (item: T) => ReactNode
  itemKey: (item: T) => string | number
  vazio?: string
}

export function ListaColapsavel<T>({
  titulo,
  dados,
  aberto,
  onToggle,
  acao,
  renderItem,
  itemKey,
  vazio = 'Nenhum registro encontrado',
}: ListaColapsavelProps<T>) {
  const restantes = dados.total - dados.itens.length

  const onScroll = (evento: UIEvent<HTMLUListElement>) => {
    const alvo = evento.currentTarget
    if (alvo.scrollTop + alvo.clientHeight >= alvo.scrollHeight - 40) {
      dados.onCarregarMais()
    }
  }

  return (
    <section className="colapsavel">
      <div className="colapsavel-cabecalho">
        <button type="button" className="colapsavel-toggle" onClick={onToggle}>
          <Icon name="chevron" size={16} className={`colapsavel-caret${aberto ? ' aberto' : ''}`} />
          <span className="colapsavel-titulo">{titulo}</span>
          {!dados.carregando && <span className="colapsavel-contagem">{dados.total}</span>}
        </button>
        {acao}
      </div>

      {aberto && (
        <div className="colapsavel-corpo">
          <div className="search-box-compact">
            <Icon name="search" size={14} />
            <input
              type="search"
              placeholder={`Buscar em ${titulo.toLowerCase()}...`}
              value={dados.termo}
              onChange={(evento) => dados.onTermoChange(evento.target.value)}
              aria-label={`Buscar em ${titulo}`}
            />
          </div>

          {dados.carregando ? (
            <div className="colapsavel-carregando">
              <Spinner size={18} />
            </div>
          ) : dados.itens.length === 0 ? (
            <p className="colapsavel-vazio">{vazio}</p>
          ) : (
            <ul className="colapsavel-lista" onScroll={onScroll}>
              {dados.itens.map((item) => (
                <li key={itemKey(item)}>{renderItem(item)}</li>
              ))}
            </ul>
          )}

          {dados.temMais && (
            <button type="button" className="colapsavel-mais" onClick={dados.onCarregarMais}>
              {dados.carregandoMais ? (
                <>
                  <Spinner size={14} /> Carregando...
                </>
              ) : (
                <>Carregar mais ({restantes} restantes)</>
              )}
            </button>
          )}

          {!dados.carregando && dados.total > 0 && (
            <p className="colapsavel-rodape">
              {dados.itens.length} de {dados.total}
            </p>
          )}
        </div>
      )}
    </section>
  )
}

export default ListaColapsavel