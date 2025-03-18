---
title: "Utiliser l'Interface en Ligne de Commande cliGCP"
linkTitle: "Tutoriel cliGCP"
weight: 3
description: >-
  Configurer et utiliser l'interface en ligne de commande cliGCP pour interagir avec les LLMs et les outils MCP
---

Ce tutoriel vous guide à travers la configuration et l'utilisation de l'interface en ligne de commande cliGCP pour interagir avec les LLMs et les outils MCP. À la fin, vous serez capable d'exécuter le CLI et d'effectuer des tâches de base avec celui-ci.

## Prérequis

Avant de commencer, assurez-vous que vous avez :

- Go 1.21 ou plus récent installé
- Un compte Google Cloud Platform avec l'API Vertex AI activée
- L'authentification Google Cloud configurée (voir le [tutoriel de démarrage](/fr/docs/tutorials/getting-started/))
- Une installation gomcptest fonctionnelle avec les outils construits

## Configuration de l'Environnement

1. **Définir les variables d'environnement** :

   ```bash
   # Votre ID de projet GCP
   export GCP_PROJECT=votre-projet-id
   
   # Région Vertex AI (par défaut: us-central1)
   export GCP_REGION=us-central1
   
   # Modèles Gemini à utiliser 
   export GEMINI_MODELS=gemini-1.5-pro,gemini-2.0-flash
   
   # Répertoire pour stocker les images temporaires
   export IMAGE_DIR=/tmp/images
   ```

2. **Créer le répertoire d'images** :

   ```bash
   mkdir -p /tmp/images
   ```

## Compilation et Exécution de cliGCP

1. **Compiler cliGCP** :

   ```bash
   cd /chemin/vers/gomcptest
   go build -o bin/cliGCP ./host/cliGCP
   ```

2. **Exécuter cliGCP avec les outils de base** :

   ```bash
   cd /chemin/vers/gomcptest
   bin/cliGCP -mcpservers "bin/GlobTool;bin/GrepTool;bin/LS;bin/View;bin/Bash"
   ```

Vous devriez voir une interface de chat interactive s'ouvrir, où vous pouvez interagir avec le modèle LLM.

## Utilisation de cliGCP

### Commandes de Base

Dans l'interface cliGCP, vous pouvez :

- Écrire des messages textuels pour interagir avec le LLM
- Utiliser `/help` pour afficher les commandes disponibles
- Utiliser `/exit` ou Ctrl+C pour quitter la session

### Exemples d'Interactions

Essayez ces exemples pour tester les capacités de base du système :

1. **Requête simple** :
   ```
   Peux-tu m'expliquer le Protocole de Contexte de Modèle en termes simples ?
   ```

2. **Utilisation de l'outil Bash** :
   ```
   Affiche les 5 derniers fichiers modifiés dans le répertoire courant.
   ```

3. **Recherche de fichiers** :
   ```
   Trouve tous les fichiers Go dans ce projet.
   ```

4. **Lecture et Explication de Code** :
   ```
   Explique ce que fait le fichier main.go dans le répertoire cliGCP.
   ```

## Utilisation d'Outils Avancés

### Dispatch Agent

Pour utiliser l'outil dispatch_agent, qui permet de déléguer des tâches complexes :

```bash
bin/cliGCP -mcpservers "bin/GlobTool;bin/GrepTool;bin/LS;bin/View;bin/dispatch_agent -glob-path bin/GlobTool -grep-path bin/GrepTool -ls-path bin/LS -view-path bin/View;bin/Bash"
```

Maintenant, vous pouvez demander des choses comme :

```
Trouve tous les fichiers Go qui importent le paquet "context" et résume leur but.
```

### Modification de Fichiers

Pour permettre la modification de fichiers, incluez les outils Edit et Replace :

```bash
bin/cliGCP -mcpservers "bin/GlobTool;bin/GrepTool;bin/LS;bin/View;bin/Bash;bin/Edit;bin/Replace"
```

Vous pouvez maintenant demander :

```
Crée un nouveau fichier README.md avec une description de base du projet.
```

## Personnalisation de l'Expérience

Pour une expérience plus personnalisée, vous pouvez définir une instruction système spécifique :

```bash
export SYSTEM_INSTRUCTION="Tu es un assistant en programmation Go expert qui aide à analyser et améliorer le code."
bin/cliGCP -mcpservers "bin/GlobTool;bin/GrepTool;bin/LS;bin/View;bin/Bash"
```

## Résolution des Problèmes Courants

### Erreurs d'Authentification

Si vous rencontrez des erreurs d'authentification :

1. Vérifiez que vous avez exécuté `gcloud auth application-default login`
2. Assurez-vous que votre compte a accès au projet GCP et à l'API Vertex AI
3. Vérifiez que la variable d'environnement `GCP_PROJECT` est correctement définie

### Erreurs d'Outil

Si les outils ne sont pas trouvés ou ne fonctionnent pas :

1. Vérifiez les chemins des outils dans la commande `-mcpservers`
2. Assurez-vous que tous les outils ont été compilés avec `make tools`
3. Vérifiez les permissions d'exécution sur les fichiers d'outils

## Prochaines Étapes

Maintenant que vous avez configuré et utilisé l'interface cliGCP, vous pouvez :

- Explorer les différentes [configurations avancées](/fr/docs/how-to/configure-cligcp/) de cliGCP
- Créer des [outils personnalisés](/fr/docs/how-to/create-custom-tool/) pour étendre les fonctionnalités
- Créer des alias shell ou des scripts pour simplifier l'accès à vos configurations préférées