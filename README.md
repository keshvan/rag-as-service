# RAG-as-a-Service Platform (MVP)

## 📌 Overview

Данный проект представляет собой MVP платформы **RAG-as-a-Service (Retrieval-Augmented Generation)**, позволяющей организациям:

* Загружать собственные документы (PDF, TXT, DOCX)
* Индексировать их в векторной базе данных
* Выполнять поиск и задавать вопросы через LLM
* Получать ответы с указанием источников (attribution)

Ключевое требование системы — **multi-tenancy**: строгая изоляция данных между организациями на всех уровнях.


## 🔐 Multi-Tenancy Model

Изоляция данных реализована на нескольких уровнях:

1. **JWT Payload**

   * Каждый токен содержит:

     * `user_id`
     * `organization_id`

2. **API Layer**

   * Gateway валидирует токен локально
   * Пробрасывает `organization_id` во все downstream сервисы

3. **Vector DB (Qdrant)**

   * Все записи содержат `organization_id`
   * Каждый запрос к Retrieval Service включает фильтр по `organization_id`

4. **Database**

   * Данные логически разделены по организациям

---

## 🧩 Services

### 1. Auth Service (Go)

**Ответственность:**

* Регистрация и аутентификация пользователей
* Управление организациями
* Генерация JWT токенов (асимметричная подпись)

**Особенности:**

* Использует приватный ключ (RSA/EdDSA)
* В payload токена добавляет:

  ```json
  {
    "user_id": "...",
    "organization_id": "..."
  }
  ```

---

### 2. API Gateway (Go)

**Единая точка входа в систему**

**Ответственность:**

* Локальная валидация JWT (по публичному ключу)
* Роутинг запросов:

  * `/auth/*` → Auth Service
  * `/documents/*` → internal handlers
  * `/rag/*` → Retrieval + LLM
* Генерация **S3 Pre-signed URLs** для загрузки файлов

**Почему важно:**

* Не делает сетевых вызовов к Auth Service при каждой проверке токена
* Уменьшает latency и повышает отказоустойчивость

---

### 3. Storage (S3)

**Назначение:**

* Хранение оригинальных документов

**Особенности:**

* Прямая загрузка с клиента через pre-signed URL
* API Gateway не участвует в передаче файлов

---

### 4. Ingestion Service (Go / Python, Async)

**Асинхронный pipeline обработки документов**

**Этапы:**

1. Получение события о загрузке файла
2. Парсинг документа (PDF, DOCX, TXT)
3. Chunking (разбиение на части)
4. Генерация embeddings
5. Сохранение в Qdrant
6. Обновление статуса документа

**Метаданные каждого чанка:**

```json
{
  "organization_id": "...",
  "document_id": "...",
  "chunk_index": 0,
  "text": "..."
}
```

---

### 5. Retrieval Service (gRPC)

**Семантический поиск**

**Ответственность:**

* Принимает запрос пользователя
* Генерирует embedding запроса
* Выполняет поиск в Qdrant

**Критически важно:**

* Каждый запрос ОБЯЗАТЕЛЬНО фильтруется:

  ```
  organization_id = <from JWT>
  ```

**Результат:**

* Топ-N релевантных чанков

---

### 6. LLM Router (gRPC)

**Оркестратор LLM-запросов**

**Ответственность:**

* Принимает:

  * user query
  * контекст (chunks)
* Формирует prompt
* Отправляет в LLM провайдера (OpenAI / Claude / др.)
* Возвращает:

  * ответ
  * источники (chunks)

**Дополнительно:**

* Возможность переключения провайдеров
* Централизованная логика prompt engineering

---

### 7. Frontend (React / Next.js)

**Пользовательский интерфейс**

**Функциональность:**

* Аутентификация
* Управление документами
* Загрузка файлов (через S3)
* Просмотр статусов индексации
* Чат-интерфейс для RAG

---

## 🗄️ Data Layer

### PostgreSQL

Используются две схемы:

#### `rag_auth`

* users
* organizations
* memberships

#### `rag_app`

* documents
* ingestion_jobs
* document_chunks (metadata only)

---

### Qdrant (Vector DB)

**Хранит:**

* embeddings
* текст чанков
* metadata

**Обязательное поле:**

```
organization_id
```

---

## 🔄 Основные флоу

### 📥 Загрузка документа

1. Frontend → API Gateway
2. Gateway → генерирует S3 pre-signed URL
3. Frontend → загружает файл напрямую в S3
4. Создается ingestion job
5. Worker обрабатывает файл
6. Данные сохраняются в Qdrant

---

### 💬 RAG-запрос

1. Frontend → API Gateway (с JWT)
2. Gateway → Retrieval Service
3. Retrieval → Qdrant (с фильтром organization_id)
4. Retrieval → LLM Router
5. LLM Router → LLM provider
6. Ответ возвращается пользователю с источниками

---