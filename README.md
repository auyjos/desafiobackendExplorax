

```markdown

---

## Tabla de Contenidos

- [Características](#características)
- [Stack y Tecnologías](#stack-y-tecnologías)
- [Estructura del Proyecto](#estructura-del-proyecto)
- [Instalación y Configuración](#instalación-y-configuración)
  - [Desarrollo Local](#desarrollo-local)
  - [Variables de Entorno](#variables-de-entorno)
- [Endpoints](#endpoints)
- [Despliegue en AWS EC2 con Docker](#despliegue-en-aws-ec2-con-docker)
  - [Paso 1: Crear una Instancia EC2](#paso-1-crear-una-instancia-ec2)
  - [Paso 2: Conexión via SSH a la Instancia](#paso-2-conexión-via-ssh-a-la-instancia)
  - [Paso 3: Instalación de Docker](#paso-3-instalación-de-docker)
  - [Paso 4: Transferir o Clonar el Código](#paso-4-transferir-o-clonar-el-código)
  - [Paso 5: Configurar el Dockerfile](#paso-5-configurar-el-dockerfile)
  - [Paso 6: Construir la Imagen Docker](#paso-6-construir-la-imagen-docker)
  - [Paso 7: Configurar Variables de Entorno en la Instancia](#paso-7-configurar-variables-de-entorno-en-la-instancia)
  - [Paso 8: Ejecutar el Contenedor](#paso-8-ejecutar-el-contenedor)
  - [Paso 9: Verificar y Probar el Despliegue](#paso-9-verificar-y-probar-el-despliegue)
  - [Paso 10: Configurar Auto-Reinicio del Contenedor](#paso-10-configurar-auto-reinicio-del-contenedor)
- [Pruebas](#pruebas)
- [Contribución](#contribución)
- [Licencia](#licencia)

---

## Características

- **Registro y Autenticación:**  
  - Registro de usuarios con validación y encriptación de contraseñas (bcrypt).  
  - Autenticación mediante JWT, con generación y validación de tokens.

- **Gestión de Misiones:**  
  - Inicio y finalización de misiones, registrando el progreso del usuario.  
  - Endpoints para consultar misiones activas, completadas y el progreso general.

- **Estadísticas y Leaderboard:**  
  - Cálculo de estadísticas del usuario (total de misiones completadas, tiempo promedio, porcentaje de avance).  
  - Leaderboard global y overview de misiones.

- **Seguridad y Escalabilidad:**  
  - Uso de JWT para proteger endpoints sensibles.  
  - Diseño modular y escalable, preparado para integrarse con un frontend.

---

## Stack y Tecnologías

- **Lenguaje:** Go (Golang)
- **Base de Datos:** MongoDB
- **Autenticación:** JWT
- **Contenerización:** Docker
- **Despliegue:** AWS EC2
- **Testing:** Pruebas unitarias e integración (Postman, etc.)

---

## Estructura del Proyecto

```
/cmd
  main.go               # Punto de entrada de la aplicación
/internal
  /handlers             # Endpoints (auth, misiones, estadísticas, etc.)
  /models               # Modelos de datos (User, Mission, MissionProgress)
  /database             # Conexión a MongoDB y operaciones CRUD
  /middleware           # Middleware de JWT y manejo de errores
  /utils                # Funciones auxiliares (por ejemplo, generación de JWT)
/tests                  # Pruebas unitarias e integración
Dockerfile              # Archivo de Docker para contenerizar la aplicación
.env                   # Archivo de variables de entorno (no incluir en repositorio)
README.md               # Este archivo
```

---

## Instalación y Configuración

### Desarrollo Local

1. **Clonar el Repositorio:**
   ```bash
git clone git@github.com:auyjos/desafiobackendExplorax.git
cd desafiobackendExplorax
   ```

2. **Instalar Dependencias:**
   ```bash
   go mod download
   ```

3. **Configurar Variables de Entorno:**
   Crea un archivo `.env` en la raíz del proyecto:
   ```env
   MONGO_URI=mongodb+srv://<usuario>:<password>@cluster.mongodb.net/explorax?retryWrites=true&w=majority
   JWT_SECRET=MiS3cr3t0
   PORT=8080
   ```

4. **Ejecutar la Aplicación:**
   ```bash
   go run cmd/main.go
   ```
   La API se iniciará en el puerto configurado (por defecto, 8080).

### Variables de Entorno

- **MONGO_URI:** Cadena de conexión a MongoDB.
- **JWT_SECRET:** Clave secreta para firmar tokens JWT.
- **PORT:** Puerto en el que se ejecuta la API.

---

## Endpoints

### Autenticación
- **POST /auth/register:** Registra un nuevo usuario.
- **POST /auth/login:** Autentica un usuario y devuelve un token JWT.

### Misiones (Endpoints Protegidos)
- **POST /missions/start:** Inicia una misión (registra progreso con estado "iniciada").
- **POST /missions/complete:** Completa una misión (actualiza el estado a "completada" y registra la fecha de finalización).
- **GET /missions/progress:** Devuelve el progreso completo del usuario.
- **GET /missions/active:** Lista misiones activas (no completadas).
- **GET /missions/completed:** Lista misiones completadas.
- **GET /missions/statistics:** Devuelve estadísticas del usuario (total completadas, promedio de duración, porcentaje de avance).

### Endpoints Públicos
- **GET /missions/leaderboard:** Ranking global de usuarios basado en misiones completadas.
- **GET /missions/overview:** Estadísticas globales (misión más popular, promedio de duración por misión).

---

## Despliegue en AWS EC2 con Docker

### Paso 1: Crear una Instancia EC2

1. Accede a la [Consola de AWS](https://aws.amazon.com/console/) y dirígete a **EC2**.
2. Haz clic en **Launch Instance**.
3. Selecciona la AMI **Amazon Linux 2**.
4. Elige el tipo de instancia (por ejemplo, `t2.micro` para Free Tier).
5. Configura los detalles de la instancia y asegúrate de abrir los siguientes puertos en el Security Group:
   - **SSH (22)**
   - **HTTP/8080** (o el puerto que uses para tu aplicación)
6. Selecciona o crea un par de llaves (archivo `.pem`).
7. Lanza la instancia y espera a que esté en estado “running”.

### Paso 2: Conectarse a la Instancia via SSH

1. Copia la **Public IPv4 address** o el **Public DNS** de la instancia.
2. En tu terminal local:
   ```bash
   chmod 400 explorax.pem
   ssh -i "explorax.pem" ec2-user@<PublicIPv4>
   ```
   Reemplaza `<PublicIPv4>` con la IP pública de la instancia.

### Paso 3: Instalar Docker en la Instancia

Ejecuta los siguientes comandos en la instancia:

```bash
sudo yum update -y
sudo amazon-linux-extras install docker -y
sudo service docker start
sudo usermod -a -G docker ec2-user
```

Cierra y vuelve a conectarte para que los cambios de grupo surtan efecto. Verifica con:
```bash
docker info
docker run hello-world
```

### Paso 4: Obtener el Código en la Instancia

Puedes clonar tu repositorio (si es público):
```bash
git clone git@github.com:auyjos/desafiobackendExplorax.git
cd desafiobackendExplorax
```
O transferir los archivos usando SCP:
```bash
scp -i explorax.pem -r ./tuProyecto ec2-user@<PublicIPv4>:/home/ec2-user/
```

### Paso 5: Configurar el Dockerfile

Asegúrate de tener un Dockerfile en la raíz del proyecto. Ejemplo:

```dockerfile
# Etapa 1: Compilación
FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main cmd/main.go

# Etapa 2: Imagen final
FROM alpine:3.17
WORKDIR /app
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
```

### Paso 6: Construir la Imagen Docker

Desde el directorio del proyecto:
```bash
docker build -t backend_explorax .
```
Verifica con:
```bash
docker images
```

### Paso 7: Configurar Variables de Entorno

Crea un archivo `.env` en el directorio del proyecto en la instancia:
```env
MONGO_URI=mongodb+srv://<usuario>:<password>@cluster.mongodb.net/explorax?retryWrites=true&w=majority
JWT_SECRET=MiS3cr3t0
PORT=8080
```

### Paso 8: Ejecutar el Contenedor

Ejecuta el contenedor usando el archivo `.env` y la política de reinicio automática:

```bash
docker run -d --restart always -p 8080:8080 --env-file .env --name backend_explorax_container backend_explorax
```

### Paso 9: Verificar el Despliegue

1. Asegúrate de que el Security Group de la instancia permita tráfico en el puerto 8080.
2. Desde tu navegador o Postman, visita:
   ```
   http://<PublicIPv4>:8080
   ```
   Prueba también los endpoints, por ejemplo, `http://<PublicIPv4>:8080/auth/login`.

3. Para ver logs:
   ```bash
   docker logs -f backend_explorax_container
   ```

### Paso 10: Configurar Auto-Reinicio del Contenedor

Si ya no lo has hecho en el comando anterior, asegúrate de que el contenedor se reinicie automáticamente:

```bash
docker run -d --restart always -p 8080:8080 --env-file .env --name backend_explorax_container backend_explorax
```

Esto garantiza que el contenedor se reinicie si la instancia se reinicia o si el contenedor falla.

---

## Pruebas

- **Unitarias:** Ejecuta `go test ./...` en tu entorno local para correr las pruebas.
- **Integración:** Usa Postman o Insomnia para probar manualmente los endpoints.
- **Swagger (Opcional):** Si has integrado Swagger, genera la documentación y prueba los endpoints.

---
```
