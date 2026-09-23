import { useCallback, useEffect, useRef, useState } from 'react'

import type { PaginaLista } from './types'

export interface ListaPagina<T> {
  termo: string
  itens: T[]
  total: number
  carregando: boolean
  carregandoMais: boolean
  temMais: boolean
  onTermoChange: (termo: string) => void
  onCarregarMais: () => void
  atualizar: () => void
}

type Buscar<T> = (termo: string, limite: number, offset: number) => Promise<PaginaLista<T>>

export function useListaPaginada<T>(buscar: Buscar<T>, limite = 7): ListaPagina<T> {
  const [termo, setTermo] = useState('')
  const [itens, setItens] = useState<T[]>([])
  const [total, setTotal] = useState(0)
  const [carregando, setCarregando] = useState(true)
  const [carregandoMais, setCarregandoMais] = useState(false)
  const requestId = useRef(0)
  const debounce = useRef<ReturnType<typeof setTimeout> | null>(null)

  const aplicaPagina = useCallback(
    (termoAtual: string, offset: number) => {
      const id = ++requestId.current
      if (offset === 0) {
        setCarregando(true)
      } else {
        setCarregandoMais(true)
      }
      buscar(termoAtual, limite, offset)
        .then((pagina) => {
          if (requestId.current !== id) return
          setTotal(pagina.total)
          setItens((prev) => (offset === 0 ? pagina.items : [...prev, ...pagina.items]))
        })
        .catch(() => {
          if (requestId.current !== id) return
          if (offset === 0) {
            setItens([])
            setTotal(0)
          }
        })
        .finally(() => {
          if (requestId.current !== id) return
          if (offset === 0) {
            setCarregando(false)
          } else {
            setCarregandoMais(false)
          }
        })
    },
    [buscar, limite],
  )

  useEffect(() => {
    if (debounce.current) clearTimeout(debounce.current)
    debounce.current = setTimeout(() => aplicaPagina(termo, 0), termo === '' ? 0 : 300)
    return () => {
      if (debounce.current) clearTimeout(debounce.current)
    }
  }, [termo, aplicaPagina])

  const onCarregarMais = useCallback(() => {
    if (carregando || carregandoMais) return
    aplicaPagina(termo, itens.length)
  }, [aplicaPagina, termo, itens.length, carregando, carregandoMais])

  const atualizar = useCallback(() => {
    aplicaPagina(termo, 0)
  }, [aplicaPagina, termo])

  return {
    termo,
    itens,
    total,
    carregando,
    carregandoMais,
    temMais: !carregando && itens.length < total,
    onTermoChange: setTermo,
    onCarregarMais,
    atualizar,
  }
}

export default useListaPaginada