# Closer

[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/alnovi/closer)](https://go.dev/dl/)
[![GitHub License](https://img.shields.io/github/license/alnovi/closer)](https://github.com/alnovi/closer/blob/master/LICENSE)
[![GitHub Release](https://img.shields.io/github/v/release/alnovi/closer)](https://github.com/alnovi/closer/releases)
![coverage](https://raw.githubusercontent.com/alnovi/closer/badges/.badges/master/coverage.svg)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/alnovi/closer/master.yml)

**Closer** — предоставляет унифицированный механизм для корректного и упорядоченного
закрытия всех зависимостей в Go‑проекте (сервисов, соединений, воркеров и т.д.).

## Установка

```sh
go get github.com/alnovi/closer
```

## Ключевые возможности:

- **Централизованное управление закрытием.** Позволяет собрать все объекты, требующие освобождения ресурсов, в единую точку выхода и закрыть их в контролируемом порядке.
- **Упорядоченное завершение в обратном порядке добавления.** Компоненты закрываются в порядке LIFO (Last In, First Out): тот, что был добавлен последним, закрывается первым.
  Это удобно для соблюдения зависимостей — например, сначала останавливаются высокоуровневые сервисы, а затем освобождаются нижележащие ресурсы (соединения, пулы и т.п.).
- **Агрегация ошибок.** При закрытии множества ресурсов собирает все возникшие ошибки и возвращает их суммарно, чтобы не потерять проблемы на отдельных этапах.
- **Безопасное повторное закрытие.** Гарантирует, что повторный вызов закрытия не приведёт к панике или некорректному поведению.

Пакет особенно полезен в долгоживущих сервисах (HTTP/GRPC серверы, фоновые задачи, сетевые соединения и т.д.), где важно гарантировать освобождение
всех ресурсов при остановке приложения.

## Пример использования

```go
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alnovi/closer"
)

func main() {
	cl := closer.New(time.Minute)
	
	signals := []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	ctx, cancel := signal.NotifyContext(context.Background(), signals...)
	defer func() {
		cl.Close()
		cancel()
	}()

	cl.Add(closer.NewWrapHandler("redis", closeRedis))
	cl.Add(closer.NewHandler("postgres", closePostgres))

	<-ctx.Done()
}

func closeRedis() {
	time.Sleep(5 * time.Second)
}

func closePostgres(_ context.Context) error {
	time.Sleep(10 * time.Second)
	return nil
}
```