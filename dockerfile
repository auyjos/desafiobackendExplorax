# Etapa 1: Construcción de la aplicación
FROM golang:1.24.1-alpine AS builder

# Instalamos dependencias necesarias para swag (por ejemplo, git)
RUN apk add --no-cache git

WORKDIR /app

# Copiar archivos de dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar el resto del código fuente
COPY . .

# Instalar la herramienta swag para generar la documentación Swagger
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Generar la documentación Swagger en la carpeta 'docs'
RUN swag init --generalInfo cmd/main.go --output docs

# Compilar la aplicación, asegurándote de apuntar al archivo main correcto
RUN go build -o main cmd/main.go

# Etapa 2: Imagen final minimalista
FROM alpine:3.17

WORKDIR /app

# Copiar el binario compilado y la documentación Swagger
COPY --from=builder /app/main .
COPY --from=builder /app/docs ./docs

# Exponer el puerto en el que corre tu app (según tu main, 8080)
EXPOSE 8080

# Ejecutar el binario por defecto
CMD ["./main"]
