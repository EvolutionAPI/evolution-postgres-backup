# 🧪 Guia de Testes da API

Este guia mostra como testar todos os endpoints da API do PostgreSQL Backup Service.

## 🔑 Autenticação

**Todos os endpoints (exceto `/health`) requerem autenticação via API Key:**

```
api-key: a4f3a241-7763-4f3b-9101-0e26c5029f17
```

## 🌐 Base URL

```
Local: http://localhost:8080
```

## 📊 Endpoints Disponíveis

### 1. 🏥 Health Check (Público)

**Verifica se o serviço está funcionando**

```http
GET /health
```

**Resposta de sucesso:**
```json
{
  "success": true,
  "message": "PostgreSQL Backup Service is running",
  "data": {
    "version": "1.0.0",
    "status": "healthy"
  }
}
```

---

### 2. 🗄️ PostgreSQL Instances

#### 📋 Listar Instâncias
```http
GET /api/v1/postgres
api-key: a4f3a241-7763-4f3b-9101-0e26c5029f17
```

**Resposta:**
```json
{
  "success": true,
  "message": "PostgreSQL instances retrieved successfully",
  "data": [
    {
      "id": "postgres-1",
      "name": "Chatpolos Postgres 1",
      "host": "manager.chatpolos.com.br",
      "port": 5432,
      "databases": ["evogo_auth", "evolution_lb"],
      "database": "evogo_auth",
      "username": "postgres",
      "enabled": true
    }
  ]
}
```

#### ➕ Adicionar Instância
```http
POST /api/v1/postgres
api-key: a4f3a241-7763-4f3b-9101-0e26c5029f17
Content-Type: application/json
```

**Body (Método 1 - Array de bancos):**
```json
{
  "name": "Novo PostgreSQL",
  "host": "localhost",
  "port": 5432,
  "databases": [
    "database1",
    "database2",
    "database3"
  ],
  "username": "postgres",
  "password": "senha123",
  "enabled": true
}
```

**Body (Método 2 - String com vírgulas):**
```json
{
  "name": "Novo PostgreSQL",
  "host": "localhost", 
  "port": 5432,
  "database": "database1,database2,database3",
  "username": "postgres",
  "password": "senha123",
  "enabled": true
}
```

#### ✏️ Atualizar Instância
```http
PUT /api/v1/postgres/{id}
api-key: a4f3a241-7763-4f3b-9101-0e26c5029f17
Content-Type: application/json
```

**Body:**
```json
{
  "name": "PostgreSQL Atualizado",
  "host": "novo-host.com",
  "port": 5432,
  "database": "novo_db1,novo_db2",
  "username": "postgres",
  "password": "nova_senha",
  "enabled": true
}
```

#### 🗑️ Deletar Instância
```http
DELETE /api/v1/postgres/{id}
api-key: a4f3a241-7763-4f3b-9101-0e26c5029f17
```

---

### 3. 💾 Backups

#### 📋 Listar Backups
```http
GET /api/v1/backups
api-key: a4f3a241-7763-4f3b-9101-0e26c5029f17
```

**Resposta:**
```json
{
  "success": true,
  "message": "Backups retrieved successfully",
  "data": [
    {
      "id": "backup-uuid-123",
      "postgresql_id": "postgres-1",
      "database_name": "evogo_auth",
      "backup_type": "manual",
      "status": "completed",
      "start_time": "2024-12-01T10:00:00Z",
      "end_time": "2024-12-01T10:05:00Z",
      "file_path": "",
      "file_size": 1024000,
      "s3_key": "backups/postgres-1/manual/2024/12/backup.sql",
      "created_at": "2024-12-01T10:00:00Z"
    }
  ]
}
```

#### 💾 Criar Backup Manual
```http
POST /api/v1/backups
api-key: a4f3a241-7763-4f3b-9101-0e26c5029f17
Content-Type: application/json
```

**Body (backup de banco específico):**
```json
{
  "postgresql_id": "postgres-1",
  "backup_type": "manual",
  "database_name": "evogo_auth"
}
```

**Body (backup do banco padrão):**
```json
{
  "postgresql_id": "postgres-1",
  "backup_type": "manual"
}
```

**Resposta:**
```json
{
  "success": true,
  "message": "Backup started successfully",
  "data": {
    "id": "backup-uuid-456",
    "postgresql_id": "postgres-1",
    "database_name": "evogo_auth",
    "backup_type": "manual",
    "status": "pending",
    "start_time": "2024-12-01T10:15:00Z",
    "created_at": "2024-12-01T10:15:00Z"
  }
}
```

#### 🔍 Ver Detalhes do Backup
```http
GET /api/v1/backups/{backup_id}
api-key: a4f3a241-7763-4f3b-9101-0e26c5029f17
```

---

### 4. 🔄 Restore

#### 📥 Restaurar Backup
```http
POST /api/v1/restore
api-key: a4f3a241-7763-4f3b-9101-0e26c5029f17
Content-Type: application/json
```

**Body:**
```json
{
  "backup_id": "backup-uuid-123",
  "postgresql_id": "postgres-1",
  "database_name": "evogo_auth_restored"
}
```

**Resposta:**
```json
{
  "success": true,
  "message": "Backup restored successfully"
}
```

---

### 5. 📋 Logs (NEW!)

#### 📄 Ver Logs
```http
GET /api/v1/logs?lines=100&date=2024-12-01&level=INFO&component=JOB
api-key: a4f3a241-7763-4f3b-9101-0e26c5029f17
```

**Query Parameters:**
- `lines` (opcional): Número de linhas (padrão: 100)
- `date` (opcional): Data no formato YYYY-MM-DD (padrão: hoje)
- `level` (opcional): Filtro por nível (INFO, WARN, ERROR, DEBUG)
- `component` (opcional): Filtro por componente (JOB, MAIN, etc.)

**Resposta:**
```json
{
  "success": true,
  "message": "Retrieved 25 log lines",
  "data": {
    "date": "2024-12-01",
    "lines": 25,
    "logs": [
      "[2024-12-01 10:15:00] [INFO] [JOB] Started BACKUP job [4cbc97a7]: Database: Chatpolos Postgres 1/evogo_auth",
      "[2024-12-01 10:15:01] [INFO] [JOB] [4cbc97a7] Status: IN_PROGRESS",
      "[2024-12-01 10:15:01] [INFO] [JOB] [4cbc97a7] Local file: /tmp/postgres-backups/backup.sql",
      "[2024-12-01 10:15:01] [INFO] [JOB] [4cbc97a7] Executing pg_dump: postgres@manager.chatpolos.com.br:5432/evogo_auth",
      "[2024-12-01 10:15:05] [INFO] [JOB] [4cbc97a7] pg_dump completed successfully",
      "[2024-12-01 10:15:05] [INFO] [JOB] [4cbc97a7] Backup file size: 1048576 bytes (1.00 MB)",
      "[2024-12-01 10:15:05] [INFO] [JOB] [4cbc97a7] S3 key: backups/postgres-1/manual/2024/12/backup.sql",
      "[2024-12-01 10:15:05] [INFO] [JOB] [4cbc97a7] Starting S3 upload...",
      "[2024-12-01 10:15:08] [INFO] [JOB] [4cbc97a7] S3 upload completed successfully",
      "[2024-12-01 10:15:08] [INFO] [JOB] [4cbc97a7] Local file cleaned up",
      "[2024-12-01 10:15:08] [INFO] [JOB] [4cbc97a7] ✅ SUCCESS: Backup completed in 8s (1.00 MB)"
    ],
    "file_path": "logs/backup_2024-12-01.log"
  }
}
```

#### 📁 Listar Arquivos de Log
```http
GET /api/v1/logs/files
api-key: a4f3a241-7763-4f3b-9101-0e26c5029f17
```

**Resposta:**
```json
{
  "success": true,
  "message": "Found 3 log files",
  "data": [
    {
      "name": "backup_2024-12-01.log",
      "date": "2024-12-01",
      "size": 15420,
      "modified": "2024-12-01T15:30:00Z",
      "path": "logs/backup_2024-12-01.log"
    }
  ]
}
```

#### 🔴 Stream de Logs em Tempo Real (SSE)
```http
GET /api/v1/logs/stream
api-key: a4f3a241-7763-4f3b-9101-0e26c5029f17
```

**Resposta (Server-Sent Events):**
```
data: [2024-12-01 15:30:45] [INFO] [JOB] Started BACKUP job [abc12345]: Database: Test/db1

data: [2024-12-01 15:30:46] [INFO] [JOB] [abc12345] Status: IN_PROGRESS

data: [2024-12-01 15:30:47] [INFO] [JOB] [abc12345] pg_dump completed successfully
```

---

## 🔧 Testando via cURL

### Health Check
```bash
curl http://localhost:8080/health
```

### Listar PostgreSQL Instances
```bash
curl -H "api-key: a4f3a241-7763-4f3b-9101-0e26c5029f17" \
     http://localhost:8080/api/v1/postgres
```

### Criar Backup Manual
```bash
curl -X POST \
  -H "api-key: a4f3a241-7763-4f3b-9101-0e26c5029f17" \
  -H "Content-Type: application/json" \
  -d '{"postgresql_id": "postgres-1", "backup_type": "manual", "database_name": "evogo_auth"}' \
  http://localhost:8080/api/v1/backups
```

### Ver Logs do Job
```bash
curl -H "api-key: a4f3a241-7763-4f3b-9101-0e26c5029f17" \
     "http://localhost:8080/api/v1/logs?lines=50&component=JOB"
```

### Stream de Logs em Tempo Real
```bash
curl -H "api-key: a4f3a241-7763-4f3b-9101-0e26c5029f17" \
     http://localhost:8080/api/v1/logs/stream
```

### Adicionar Nova Instância PostgreSQL
```bash
curl -X POST \
  -H "api-key: a4f3a241-7763-4f3b-9101-0e26c5029f17" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test PostgreSQL",
    "host": "localhost",
    "port": 5432,
    "database": "test_db1,test_db2",
    "username": "postgres",
    "password": "test123",
    "enabled": true
  }' \
  http://localhost:8080/api/v1/postgres
```

---

## 🚀 Testando via Postman

### 1. **Importar Collection**
   - Baixe o arquivo `postman_collection.json` (criado abaixo)
   - No Postman: File → Import → Selecione o arquivo

### 2. **Configurar Environment**
   - Crie um Environment no Postman
   - Adicione as variáveis:
     ```
     base_url: http://localhost:8080
     api_key: a4f3a241-7763-4f3b-9101-0e26c5029f17
     ```

### 3. **Executar Testes**
   - Comece com Health Check
   - Liste PostgreSQL instances
   - Crie um backup manual
   - Monitore o progresso via logs

---

## ⚠️ Códigos de Resposta

| Código | Status | Descrição |
|--------|--------|-----------|
| 200 | OK | Operação realizada com sucesso |
| 201 | Created | Recurso criado com sucesso |
| 202 | Accepted | Backup iniciado (processamento assíncrono) |
| 400 | Bad Request | Dados inválidos ou erro de validação |
| 401 | Unauthorized | API Key inválida ou ausente |
| 404 | Not Found | Recurso não encontrado |
| 409 | Conflict | Recurso já existe (ex: ID duplicado) |
| 500 | Internal Error | Erro interno do servidor |

---

## 🎯 Fluxo de Teste Recomendado

1. **📊 Health Check** - Verificar se API está funcionando
2. **📋 Listar PostgreSQL** - Ver instâncias configuradas  
3. **💾 Backup Manual** - Testar backup de uma instância
4. **📋 Monitorar Logs** - Acompanhar execução em tempo real
5. **👀 Verificar Status** - Ver detalhes do backup criado
6. **🔄 Restore** - Testar restore (opcional)

---

## 🐛 Troubleshooting

### **Erro 401 - Unauthorized**
```json
{
  "success": false,
  "error": "Invalid API key"
}
```
**Solução**: Verificar se o header `api-key` está correto.

### **Erro 404 - PostgreSQL not found**
```json
{
  "success": false,
  "error": "PostgreSQL instance postgres-1 not found"
}
```
**Solução**: Verificar se o `postgresql_id` existe usando `GET /api/v1/postgres`.

### **Erro 500 - S3 Connection**
```json
{
  "success": false,
  "error": "failed to access S3 bucket"
}
```
**Solução**: Verificar configurações S3 no arquivo `.env`.

### **Job não executa**
```
Status: in_progress mas nenhum log aparece
```
**Solução**: 
1. Verificar logs via `GET /api/v1/logs?component=JOB`
2. Verificar se `pg_dump` está instalado: `which pg_dump`
3. Verificar conectividade PostgreSQL
4. Verificar credenciais S3

---

## 📝 Logs

Para acompanhar os logs durante os testes:

```bash
# Via API
curl -H "api-key: a4f3a241-7763-4f3b-9101-0e26c5029f17" \
     "http://localhost:8080/api/v1/logs?component=JOB&lines=50"

# Arquivo local
tail -f logs/backup_$(date +%Y-%m-%d).log

# Se rodando com make dev-simple
# Os logs aparecem no terminal

# Se rodando com Docker
docker-compose logs -f postgres-backup
``` 