# Utiliser une image officielle Go
FROM golang:1.23

# Définir le répertoire de travail
WORKDIR /app

# Copier les fichiers de l'application
COPY . .

# Installer les dépendances
RUN go mod tidy

# Compiler l'application
RUN go build -o main .

# Exposer le port
EXPOSE 8080

# Lancer l'application
CMD ["./main"]
