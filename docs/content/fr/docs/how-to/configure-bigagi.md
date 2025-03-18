---
title: "Comment Utiliser le Serveur OpenAI avec big-AGI"
linkTitle: "Utilisation avec big-AGI"
weight: 3
description: >-
  Configurer le serveur compatible OpenAI de gomcptest comme backend pour big-AGI
---

Ce guide vous montre comment configurer le serveur compatible OpenAI de gomcptest pour fonctionner avec [big-AGI](https://github.com/enricoros/big-agi), un client web open-source populaire pour les assistants IA.

## Prérequis

- Une installation fonctionnelle de gomcptest
- Le serveur compatible OpenAI en cours d'exécution (voir le [tutoriel du serveur OpenAI](/fr/docs/tutorials/openaiserver-tutorial/))
- [Node.js](https://nodejs.org/) (version 18.17.0 ou plus récente)
- Git

## Pourquoi Utiliser big-AGI avec gomcptest ?

big-AGI fournit une interface web soignée et riche en fonctionnalités pour interagir avec des modèles d'IA. En le connectant au serveur compatible OpenAI de gomcptest, vous obtenez :

- Une interface web professionnelle pour vos interactions avec l'IA
- Support pour les outils/appels de fonction
- Gestion de l'historique des conversations
- Gestion des personas
- Capacités de génération d'images
- Support pour plusieurs utilisateurs

## Configuration de big-AGI

1. **Cloner le dépôt big-AGI** :

   ```bash
   git clone https://github.com/enricoros/big-agi.git
   cd big-agi
   ```

2. **Installer les dépendances** :

   ```bash
   npm install
   ```

3. **Créer un fichier `.env.local`** pour la configuration :

   ```bash
   cp .env.example .env.local
   ```

4. **Éditer le fichier `.env.local`** pour configurer la connexion à votre serveur gomcptest :

   ```
   # Configuration big-AGI

   # URL de votre serveur compatible OpenAI gomcptest
   OPENAI_API_HOST=http://localhost:8080
   
   # Il peut s'agir de n'importe quelle chaîne car le serveur gomcptest n'utilise pas de clés API
   OPENAI_API_KEY=gomcptest-local-server
   
   # Définissez cette valeur sur true pour activer le fournisseur personnalisé
   OPENAI_API_ENABLE_CUSTOM_PROVIDER=true
   ```

5. **Démarrer big-AGI** :

   ```bash
   npm run dev
   ```

6. Ouvrez votre navigateur et accédez à `http://localhost:3000` pour accéder à l'interface big-AGI.

## Configuration de big-AGI pour Utiliser Vos Modèles

Le serveur compatible OpenAI de gomcptest expose les modèles Google Cloud via une API compatible OpenAI. Dans big-AGI, vous devrez configurer les modèles :

1. Ouvrez big-AGI dans votre navigateur
2. Cliquez sur l'icône **Paramètres** (engrenage) en haut à droite
3. Allez dans l'onglet **Modèles**
4. Sous "Modèles OpenAI" :
   - Cliquez sur "Ajouter des modèles"
   - Ajoutez vos modèles par ID (par exemple, `gemini-1.5-pro`, `gemini-2.0-flash`)
   - Définissez la longueur de contexte de manière appropriée (8K-32K selon le modèle)
   - Définissez la capacité d'appel de fonction sur `true` pour les modèles qui la prennent en charge

## Activation de l'Appel de Fonction avec des Outils

Pour utiliser les outils MCP via l'interface d'appel de fonction de big-AGI :

1. Dans big-AGI, cliquez sur l'icône **Paramètres**
2. Allez dans l'onglet **Avancé**
3. Activez "Appel de fonction" dans la section "Fonctionnalités expérimentales"
4. Dans une nouvelle conversation, cliquez sur l'onglet "Fonctions" (icône de plugin) dans l'interface de chat
5. Les outils disponibles de votre serveur gomcptest devraient être listés

## Configuration CORS pour big-AGI

Si vous exécutez big-AGI sur un domaine ou un port différent de celui de votre serveur gomcptest, vous devrez activer CORS côté serveur. Modifiez la configuration du serveur OpenAI :

1. Créez ou modifiez un middleware CORS pour le serveur OpenAI :

   ```go
   // Middleware CORS avec autorisation d'origine spécifique
   func corsMiddleware(next http.Handler) http.Handler {
       return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
           // Autoriser les requêtes provenant de l'origine big-AGI
           w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
           w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
           w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
           
           if r.Method == "OPTIONS" {
               w.WriteHeader(http.StatusOK)
               return
           }
           
           next.ServeHTTP(w, r)
       })
   }
   ```

2. Appliquez ce middleware à vos routes de serveur

## Résolution des Problèmes Courants

### Modèle Non Trouvé

Si big-AGI signale que les modèles ne peuvent pas être trouvés :

1. Vérifiez que votre serveur gomcptest est en cours d'exécution et accessible
2. Vérifiez les journaux du serveur pour vous assurer que les modèles sont correctement enregistrés
3. Assurez-vous que les ID de modèle dans big-AGI correspondent exactement à ceux fournis par votre serveur gomcptest

### L'Appel de Fonction Ne Fonctionne Pas

Si les outils ne fonctionnent pas correctement :

1. Assurez-vous que les outils sont correctement enregistrés dans votre serveur gomcptest
2. Vérifiez que l'appel de fonction est activé dans les paramètres de big-AGI
3. Vérifiez que le modèle que vous utilisez prend en charge l'appel de fonction

### Problèmes de Connexion

Si big-AGI ne peut pas se connecter à votre serveur :

1. Vérifiez la valeur `OPENAI_API_HOST` dans votre fichier `.env.local`
2. Recherchez des problèmes CORS dans la console de développement de votre navigateur
3. Assurez-vous que votre serveur est en cours d'exécution et accessible depuis le navigateur

## Déploiement en Production

Pour une utilisation en production, considérez :

1. **Sécurisation de votre API** :
   - Ajoutez une authentification appropriée à votre serveur OpenAI gomcptest
   - Mettez à jour la `OPENAI_API_KEY` dans big-AGI en conséquence

2. **Déploiement de big-AGI** :
   - Suivez le [guide de déploiement de big-AGI](https://github.com/enricoros/big-agi/blob/main/docs/deployment.md)
   - Configurez les variables d'environnement pour pointer vers votre serveur gomcptest de production

3. **Mise en place de HTTPS** :
   - Pour la production, big-AGI et votre serveur gomcptest devraient utiliser HTTPS
   - Envisagez d'utiliser un proxy inverse comme Nginx avec des certificats Let's Encrypt

## Exemple : Interface de Chat Basique

Une fois que tout est configuré, vous pouvez utiliser l'interface de big-AGI pour interagir avec vos modèles d'IA :

1. Commencez une nouvelle conversation
2. Sélectionnez votre modèle dans le menu déroulant des modèles (par exemple, `gemini-1.5-pro`)
3. Activez l'appel de fonction si vous souhaitez utiliser des outils
4. Commencez à discuter avec votre assistant IA, propulsé par gomcptest

L'interface big-AGI offre une expérience beaucoup plus riche qu'une interface en ligne de commande, avec des fonctionnalités comme l'historique des conversations, le rendu markdown, la mise en évidence du code, et plus encore.