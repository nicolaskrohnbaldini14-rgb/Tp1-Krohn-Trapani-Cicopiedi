package logic

import "errors"

var (
	ErrNoEncontrado = errors.New("no encontrado")
	ErrValidacion   = errors.New("datos inválidos")
	ErrConflicto    = errors.New("conflicto")
)