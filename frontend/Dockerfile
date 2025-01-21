# Utiliser une image officielle Node.js pour le build
FROM node:20

# Définir le répertoire de travail
WORKDIR /app

# Copier les fichiers de l'application
COPY . .

# Installer les dépendances et construire l'application avec Yarn
RUN yarn install && yarn build

# Installer un serveur HTTP statique pour servir les fichiers (par exemple, serve)
RUN yarn global add serve

# Exposer le port
EXPOSE 3000

# Lancer l'application
CMD ["serve", "-s", "build"]

