---
title: "Démarrage avec gomcptest"
linkTitle: "Démarrage"
weight: 1
description: >-
  Démarrez rapidement avec gomcptest grâce à ce guide pour débutants
---

Ce tutoriel vous guidera à travers la configuration du système gomcptest et la mise en place de l'authentification Google Cloud pour le projet.

## Prérequis

- Go >= 1.21 installé sur votre système
- Compte Google Cloud avec accès à l'API Vertex AI
- [Google Cloud CLI](https://cloud.google.com/sdk/docs/install) installé
- Familiarité de base avec le terminal/ligne de commande

## Configuration de l'Authentification Google Cloud

Avant d'utiliser gomcptest avec les services de Google Cloud Platform comme Vertex AI, vous devez configurer votre authentification.

### 1. Initialiser le CLI Google Cloud

Si vous n'avez pas encore configuré le CLI Google Cloud, exécutez :

```bash
gcloud init
```

Cette commande interactive vous guidera à travers :
- La connexion à votre compte Google
- La sélection d'un projet Google Cloud
- La définition des configurations par défaut

### 2. Se connecter à Google Cloud

Authentifiez votre CLI gcloud avec votre compte Google :

```bash
gcloud auth login
```

Cela ouvrira une fenêtre de navigateur où vous pourrez vous connecter à votre compte Google.

### 3. Configurer les Identifiants par Défaut de l'Application (ADC)

Les Identifiants par Défaut de l'Application sont utilisés par les bibliothèques clientes pour trouver automatiquement les identifiants lors de la connexion aux services Google Cloud :

```bash
gcloud auth application-default login
```

Cette commande va :
1. Ouvrir une fenêtre de navigateur pour l'authentification
2. Stocker vos identifiants localement (généralement dans `~/.config/gcloud/application_default_credentials.json`)
3. Configurer votre environnement pour utiliser ces identifiants lors de l'accès aux API Google Cloud

Ces identifiants seront utilisés par gomcptest lors de l'interaction avec les services Google Cloud.

## Configuration du Projet

1. **Cloner le dépôt** :
   ```bash
   git clone https://github.com/owulveryck/gomcptest.git
   cd gomcptest
   ```

2. **Construire les Outils** : Compiler tous les outils compatibles MCP
   ```bash
   make tools
   ```

3. **Choisir l'Interface** : 
   - Exécuter le serveur compatible OpenAI : Voir le [Tutoriel du Serveur OpenAI](/fr/docs/tutorials/openaiserver-tutorial/)
   - Utiliser le CLI directement : Voir le [Tutoriel cliGCP](/fr/docs/tutorials/cligcp-tutorial/)

## Prochaines Étapes

Après avoir terminé la configuration de base :
- Explorez les différents outils dans le répertoire `tools`
- Essayez de créer des tâches d'agent avec gomcptest
- Consultez les guides pratiques pour des cas d'utilisation spécifiques