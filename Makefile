# Makefile для курса символьного исполнения

.PHONY: all build test test-unit test-integration lint fmt clean examples help

# Переменные
GO := go
GOLINT := golangci-lint
BINARY_DIR := bin
EXAMPLES_DIR := examples

all: build test lint ## Сборка, тестирование и линтинг

build: ## Сборка всех компонентов
	@echo "🔨 Сборка проекта..."
	$(GO) build -v ./...

test: test-unit test-integration ## Запуск всех тестов

test-unit: ## Запуск unit тестов
	@echo "🧪 Запуск unit тестов..."
	$(GO) test -v ./pkg/...

test-integration: ## Запуск примеров
	@echo "🔗 Проверка работоспособности примеров..."
	$(GO) run ./examples/basic_z3_example.go
	$(GO) run ./homework1/main.go
	$(GO) run ./homework2/main.go

test-coverage: ## Запуск тестов с покрытием
	@echo "📊 Анализ покрытия кода..."
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Отчёт сохранён в coverage.html"

lint: ## Линтинг кода
	@echo "🔍 Проверка качества кода..."
	$(GO) vet ./...
	$(GO) fmt ./...
	@if command -v $(GOLINT) >/dev/null 2>&1; then \
		$(GOLINT) run; \
	else \
		echo "⚠️  golangci-lint не установлен, пропускаем дополнительную проверку"; \
	fi

fmt: ## Форматирование кода
	@echo "✨ Форматирование кода..."
	$(GO) fmt ./...
	goimports -w . || true

clean: ## Очистка временных файлов
	@echo "🧹 Очистка..."
	$(GO) clean
	rm -rf $(BINARY_DIR)
	rm -f coverage.out coverage.html

examples: ## Запуск демонстрационных примеров
	@echo "🚀 Запуск примеров..."
	@if [ -f $(EXAMPLES_DIR)/basic_z3_example.go ]; then \
		echo "Запуск базового примера Z3:"; \
		cd $(EXAMPLES_DIR) && $(GO) run basic_z3_example.go; \
	fi

# Команды для домашних заданий
hw1: ## Переход к домашнему заданию 1
	@echo "📚 Домашнее задание 1: Control Flow Graph"
	@echo "📁 cd homework1/"
	@echo "📖 Изучите README.md для подробных инструкций"
	@echo "🔧 Реализуйте методы в cfg/types.go, cfg/builder.go, cfg/visualizer.go"
	@echo "▶️  go run main.go - для тестирования"

hw2: ## Переход к домашнему заданию 2  
	@echo "📚 Домашнее задание 2: Символьные выражения"
	@echo "📁 cd homework2/"
	@echo "📖 Изучите README.md для подробных инструкций"  
	@echo "� Реализуйте методы в symbolic/, translator/, ssa_converter/"
	@echo "▶️  go run main.go - для тестирования"

# Установка зависимостей
deps: ## Установка зависимостей
	@echo "📦 Установка зависимостей..."
	$(GO) mod tidy
	$(GO) mod download

deps-dev: deps ## Установка зависимостей для разработки
	@echo "🛠 Установка инструментов разработки..."
	$(GO) install golang.org/x/tools/cmd/goimports@latest || true
	@echo "Для установки golangci-lint:"
	@echo "curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$(go env GOPATH)/bin v1.54.2"

# Проверка Z3
check-z3: ## Проверка установки Z3
	@echo "🔍 Проверка Z3..."
	@if command -v z3 >/dev/null 2>&1; then \
		echo "✅ Z3 установлен: $$(z3 --version)"; \
	else \
		echo "❌ Z3 не найден!"; \
		echo "Установите Z3:"; \
		echo "  macOS: brew install z3"; \
		echo "  Ubuntu/Debian: sudo apt-get install z3"; \
		echo "  Или соберите из исходников: https://github.com/Z3Prover/z3"; \
		exit 1; \
	fi

# Генерация документации
docs: ## Генерация документации
	@echo "📖 Генерация документации..."
	$(GO) doc -all ./pkg/z3wrapper

init: deps check-z3 ## Первоначальная настройка проекта
	@echo "🎉 Проект готов к работе!"
	@echo ""
	@echo "Доступные команды:"
	@echo "  make examples  - запуск примеров"
	@echo "  make hw1      - информация о ДЗ1 (CFG)"  
	@echo "  make hw2      - информация о ДЗ2 (Символьные выражения)"
	@echo ""
	@echo "Структура проекта:"
	@echo "  homework1/    - готовые темплейты для CFG анализа" 
	@echo "  homework2/    - готовые темплейты для символьных выражений"
	@echo ""
	@echo "Начните с: make examples, затем cd homework1/"

help: ## Показать справку
	@echo "Доступные команды:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)