# Evolution PostgreSQL Backup Service

Sistema completo de backup PostgreSQL com **Frontend React** + **Backend Go**, suporte a armazenamento S3-compatible (AWS S3, Backblaze B2, MinIO, Cloudflare R2, Hetzner) e interface web moderna.

## 🐳 **Docker - Inicio Rápido** 

```bash
# Clonar e configurar
git clone <repositorio>
cd evolution-postgres-backup

# Editar config.json com suas instâncias PostgreSQL
# Editar docker-compose.yml com suas credenciais S3

# Executar sistema completo
make docker-up

# Acessar interface web
open http://localhost:3000
```

**🌐 URLs de Acesso:**
- **Frontend Web**: http://localhost:3000
- **API Backend**: http://localhost:3000/api/v1  
- **Health Check**: http://localhost:3000/health

**📚 Documentação Docker Completa**: [DOCKER.md](DOCKER.md)

### 🛠️ **Comandos Principais**

```bash
# Sistema completo
make docker-up      # Iniciar frontend + backend  
make docker-down    # Parar tudo
make docker-logs    # Ver logs

# Desenvolvimento
make dev-backend    # Backend local (Go)
make dev-frontend   # Frontend local (Vite)

# Manutenção  
make docker-rebuild # Rebuild completo
make clean         # Limpar Docker
```

**💡 Dica**: Execute `make help` para ver todos os comandos disponíveis!

### **✅ Configurado para Hetzner Object Storage:**

**Arquivo `.env` já configurado:**
```env
# S3 Configuration (Hetzner Object Storage)
S3_ENDPOINT=https://hel1.your-objectstorage.com
S3_REGION=hel1
S3_BUCKET=backup-chatpolos
S3_ACCESS_KEY_ID=M4WID7GXREH2EC5J30V8
S3_SECRET_ACCESS_KEY=pLTF4vVMSnon1AL6NI9iTU86G0fTaVd7QyG6xfax
S3_USE_SSL=true
```

⚠️ **Importante**: Todas as configurações S3 vêm do arquivo `.env` para máxima segurança!

## 🚀 Funcionalidades

- **Backup Manual**: Via API REST
- **Restore Manual**: Via API REST  
- **Backup Automático**: Com rotinas de cron configuráveis
- **Gerenciamento PostgreSQL**: Cadastro/listagem de servidores PostgreSQL
- **Armazenamento S3**: Suporte completo a serviços S3-compatible
- **Política de Retenção**: Automática baseada em tipo de backup
- **API Segura**: Autenticação via API Key

## 📋 Política de Retenção

| Tipo de Backup | Frequência | Retenção |
|---|---|---|
| **Hourly** | A cada hora | 24 horas |
| **Daily** | Diário às 02:00 | 30 dias |
| **Weekly** | Domingo às 03:00 | 8 semanas |
| **Monthly** | 1º do mês às 04:00 | 12 meses |
| **Manual** | Sob demanda | Permanente |

## 🛠️ Instalação

### Pré-requisitos

- Go 1.21+
- PostgreSQL com `pg_dump` e `psql` no PATH
- Acesso a serviço S3-compatible (AWS S3, Backblaze B2, MinIO, etc.)

### 1. Clone e instale dependências

```bash
git clone <repository-url>
cd evolution-postgres-backup
go mod download
```

### 2. Configure as variáveis de ambiente

Copie o arquivo `.env.example` para `.env`:

```bash
cp .env.example .env
```

#### Para Hetzner Object Storage:

```env
# API Configuration
PORT=8080
API_KEY=your-secure-api-key-here

# S3 Configuration (Hetzner Object Storage)
S3_ENDPOINT=https://hel1.your-objectstorage.com
S3_REGION=hel1
S3_BUCKET=backup-chatpolos
S3_ACCESS_KEY_ID=M4WID7GXREH2EC5J30V8
S3_SECRET_ACCESS_KEY=pLTF4vVMSnon1AL6NI9iTU86G0fTaVd7QyG6xfax
S3_USE_SSL=true

# Application Configuration
LOG_LEVEL=info
BACKUP_TEMP_DIR=/tmp/postgres-backups
```

#### Para AWS S3:

```env
# API Configuration  
PORT=8080
API_KEY=your-secure-api-key-here

# S3 Configuration (AWS S3)
S3_ENDPOINT=
S3_REGION=us-east-1
S3_BUCKET=your-backup-bucket
S3_ACCESS_KEY_ID=your-aws-access-key
S3_SECRET_ACCESS_KEY=your-aws-secret-key
S3_USE_SSL=true

# Application Configuration
LOG_LEVEL=info
BACKUP_TEMP_DIR=/tmp/postgres-backups
```

#### Para MinIO:

```env
# API Configuration
PORT=8080
API_KEY=your-secure-api-key-here

# S3 Configuration (MinIO)
S3_ENDPOINT=http://localhost:9000
S3_REGION=us-east-1
S3_BUCKET=postgres-backups
S3_ACCESS_KEY_ID=minioadmin
S3_SECRET_ACCESS_KEY=minioadmin
S3_USE_SSL=false

# Application Configuration
LOG_LEVEL=info
BACKUP_TEMP_DIR=/tmp/postgres-backups
```

### 3. Configure o arquivo `config.json`

⚠️ **Nota**: As configurações S3 agora são definidas via variáveis de ambiente no `.env`, não no `config.json`. O `config.json` é usado apenas para PostgreSQL e políticas de retenção.

```json
{
  "postgresql_instances": [
    {
      "id": "postgres-1",
      "name": "Production Database",
      "host": "localhost",
      "port": 5432,
      "database": "production_db",
      "username": "postgres",
      "password": "your_password",
      "enabled": true
    }
  ],
  "retention_policy": {
    "hourly": 24,
    "daily": 30,
    "weekly": 8,
    "monthly": 12
  },
  "s3_config": {
    "endpoint": "",
    "region": "",
    "bucket": "",
    "access_key_id": "",
    "secret_access_key": "",
    "use_ssl": true
  }
}
```

As configurações S3 são carregadas automaticamente das variáveis de ambiente definidas no `.env`.

### 4. Execute o serviço

```bash
# Build
go build -o postgres-backup cmd/server/main.go

# Run
./postgres-backup
```

## 📡 API Endpoints

Todas as rotas (exceto `/health`) requerem header `Authorization: your-api-key`.

### Health Check

```bash
GET /health
```

### PostgreSQL Instances

```bash
# Listar instâncias
GET /api/v1/postgres

# Adicionar instância
POST /api/v1/postgres
{
  "name": "My Database",
  "host": "localhost",
  "port": 5432,
  "database": "mydb",
  "username": "postgres", 
  "password": "password",
  "enabled": true
}

# Atualizar instância
PUT /api/v1/postgres/{id}

# Deletar instância  
DELETE /api/v1/postgres/{id}
```

### Backups

```bash
# Listar backups
GET /api/v1/backups

# Criar backup manual
POST /api/v1/backups
{
  "postgresql_id": "postgres-1",
  "backup_type": "manual",
  "database_name": "optional_specific_db"
}

# Ver detalhes do backup
GET /api/v1/backups/{id}
```

### Restore

```bash
# Restaurar backup
POST /api/v1/restore
{
  "backup_id": "backup-uuid",
  "postgresql_id": "postgres-1", 
  "database_name": "target_db"
}
```

## 🔧 Configuração Avançada

### Personalizando Schedule de Backup

O arquivo `internal/scheduler/scheduler.go` contém as configurações de cron:

```go
// Hourly: 0 0 * * * * (a cada hora no minuto 0)
// Daily: 0 0 2 * * * (diário às 02:00)  
// Weekly: 0 0 3 * * 0 (domingo às 03:00)
// Monthly: 0 0 4 1 * * (1º do mês às 04:00)
```

### Estrutura de Arquivos no S3

```
backups/
├── {postgres_id}/
│   ├── hourly/
│   │   ├── 2024/
│   │   │   └── 01/
│   │   │       └── backup_file.sql
│   ├── daily/
│   ├── weekly/
│   └── monthly/
```

### Variáveis de Ambiente Disponíveis

| Variável | Descrição | Padrão |
|---|---|---|
| `PORT` | Porta da API | `8080` |
| `API_KEY` | Chave de autenticação | **obrigatório** |
| `S3_ENDPOINT` | Endpoint S3 (ex: https://s3.region.backblazeb2.com) | vazio |
| `S3_REGION` | Região S3 | **obrigatório** |
| `S3_BUCKET` | Nome do bucket S3 | **obrigatório** |
| `S3_ACCESS_KEY_ID` | Access Key S3 | **obrigatório** |
| `S3_SECRET_ACCESS_KEY` | Secret Key S3 | **obrigatório** |
| `S3_USE_SSL` | Usar SSL/TLS (true/false) | `true` |
| `LOG_LEVEL` | Nível de log | `info` |
| `BACKUP_TEMP_DIR` | Diretório temporário | `/tmp/postgres-backups` |

## 🐳 Docker

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o postgres-backup cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add postgresql-client ca-certificates
WORKDIR /root/
COPY --from=builder /app/postgres-backup .
COPY --from=builder /app/config.json .

CMD ["./postgres-backup"]
```

## 🚀 Deploy

### Com Docker Compose

```yaml
version: '3.8'
services:
  postgres-backup:
    build: .
    ports:
      - "8080:8080"
    environment:
      - API_KEY=your-secure-api-key
      - S3_ACCESS_KEY_ID=your-s3-key
      - S3_SECRET_ACCESS_KEY=your-s3-secret
    volumes:
      - ./config.json:/root/config.json
      - /tmp/postgres-backups:/tmp/postgres-backups
    restart: unless-stopped
```

### Como Systemd Service

```ini
[Unit]
Description=PostgreSQL Backup Service
After=network.target

[Service]
Type=simple
User=postgres-backup
WorkingDirectory=/opt/postgres-backup
ExecStart=/opt/postgres-backup/postgres-backup
Restart=always
Environment=API_KEY=your-secure-api-key

[Install]
WantedBy=multi-user.target
```

## 📊 Monitoramento

### Logs

O serviço registra todas as operações importantes:

```bash
# Acompanhar logs
tail -f /var/log/postgres-backup.log

# Com systemd
journalctl -u postgres-backup -f
```

### Métricas

- Status de backup via `/api/v1/backups`
- Health check via `/health`
- Logs estruturados para integração com Prometheus/Grafana

## 🔒 Segurança

1. **API Key**: Sempre use uma chave segura e única
2. **HTTPS**: Configure um proxy reverso (nginx) com SSL
3. **Firewall**: Limite acesso à porta da API
4. **Credenciais**: Use variáveis de ambiente para dados sensíveis
5. **Backup Encryption**: Configure encryption no bucket S3

## 🐛 Troubleshooting

### Erro de conexão PostgreSQL

```bash
# Teste manual
pg_dump -h host -p port -U username -d database --version
```

### Erro de conexão S3

```bash
# Verifique credenciais e endpoint
# Para Backblaze B2: endpoint deve ser regional
# https://s3.{region}.backblazeb2.com
```

### Permissões de arquivo

```bash
# Diretório temporário
sudo mkdir -p /tmp/postgres-backups
sudo chown postgres-backup:postgres-backup /tmp/postgres-backups
```

## 📝 Licença

MIT License

## 🤝 Contribuição

1. Fork o projeto
2. Crie uma branch para sua feature
3. Commit suas mudanças
4. Push para a branch
5. Abra um Pull Request 