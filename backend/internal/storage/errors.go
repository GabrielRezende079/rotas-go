package storage

import "errors"

// ErrNaoEncontrada indica que a rota salva não existe.
var ErrNaoEncontrada = errors.New("rota salva não encontrada")

// ErrBaseNaoEncontrada indica que a base não existe.
var ErrBaseNaoEncontrada = errors.New("base não encontrada")

// ErrPlacaDuplicada indica que já existe um veículo com a mesma placa.
var ErrPlacaDuplicada = errors.New("placa já cadastrada")

// ErrVeiculoNaoEncontrado indica que o veículo não existe.
var ErrVeiculoNaoEncontrado = errors.New("veículo não encontrado")
