package main

// Тип для ключей контекста
type contextKey int

// Константы для различных ключей контекста
const (
    idTokenKey contextKey = iota
    // Другие ключи при необходимости:
    // userInfoKey
    // requestIDKey
)