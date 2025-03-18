# Usa una imagen oficial de Go para compilar
FROM golang:1.24.1-alpine AS builder

# Crea un directorio de trabajo
WORKDIR /app

# Copia los archivos go.mod y go.sum para descargar dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copia el resto de tu código
COPY . .

# Compila la aplicación
RUN go build -o main cmd/main.go

# Segunda etapa: imagen minimalista
FROM alpine:3.17

# Crea un directorio para tu app
WORKDIR /app

# Copia el binario compilado desde la etapa anterior
COPY --from=builder /app/main .

# Expón el puerto en el que tu app escucha
EXPOSE 8080

# Comando por defecto para ejecutar la app
CMD ["./main"]
