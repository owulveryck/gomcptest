---
title: "Comment Configurer l'Interface en Ligne de Commande cliGCP"
linkTitle: "Configurer cliGCP"
weight: 3
description: >-
  Personnaliser l'outil cliGCP avec des variables d'environnement et des options de ligne de commande
---

Ce guide vous montre comment configurer et personnaliser l'interface en ligne de commande cliGCP pour divers cas d'utilisation.

## Prérequis

- Une installation fonctionnelle de gomcptest
- Une familiarité de base avec l'outil cliGCP à partir du tutoriel
- Compréhension des variables d'environnement et de la configuration

## Arguments de Ligne de Commande

L'outil cliGCP accepte les arguments de ligne de commande suivants :

```bash
# Spécifier les serveurs MCP à utiliser (requis)
-mcpservers "outil1;outil2;outil3"

# Exemple avec des arguments d'outil
./cliGCP -mcpservers "./GlobTool;./GrepTool;./dispatch_agent -glob-path ./GlobTool -grep-path ./GrepTool -ls-path ./LS -view-path ./View;./Bash"
```

## Configuration des Variables d'Environnement

### Configuration GCP

Configurez l'intégration avec Google Cloud Platform avec ces variables d'environnement :

```bash
# ID du projet GCP (requis)
export GCP_PROJECT=votre-id-de-projet-gcp

# Région GCP (par défaut : us-central1)
export GCP_REGION=us-central1

# Liste de modèles Gemini séparés par des virgules (requis)
export GEMINI_MODELS=gemini-1.5-pro,gemini-2.0-flash

# Répertoire pour stocker les images (requis pour la génération d'images)
export IMAGE_DIR=/chemin/vers/repertoire/images
```

### Configuration Avancée

Vous pouvez personnaliser le comportement de l'outil cliGCP avec ces variables d'environnement supplémentaires :

```bash
# Définir une instruction système personnalisée pour le modèle
export SYSTEM_INSTRUCTION="Vous êtes un assistant utile spécialisé dans la programmation Go."

# Ajuster la température du modèle (0.0-1.0, la valeur par défaut est 0.2)
# Les valeurs plus basses rendent la sortie plus déterministe, les valeurs plus élevées plus créatives
export MODEL_TEMPERATURE=0.3

# Définir une limite maximale de tokens pour les réponses
export MAX_OUTPUT_TOKENS=2048
```

## Création d'Alias Shell

Pour simplifier l'utilisation, créez des alias shell dans votre `.bashrc` ou `.zshrc` :

```bash
# Ajouter à ~/.bashrc ou ~/.zshrc
alias gpt='cd /chemin/vers/gomcptest/bin && ./cliGCP -mcpservers "./GlobTool;./GrepTool;./LS;./View;./Bash;./Replace"'

# Créer des alias spécialisés pour différentes tâches
alias assistant-code='cd /chemin/vers/gomcptest/bin && GCP_PROJECT=votre-projet GEMINI_MODELS=gemini-2.0-flash ./cliGCP -mcpservers "./GlobTool;./GrepTool;./LS;./View;./Bash;./Replace"'

alias scanner-securite='cd /chemin/vers/gomcptest/bin && SYSTEM_INSTRUCTION="Vous êtes un expert en sécurité concentré sur la recherche de vulnérabilités dans le code" ./cliGCP -mcpservers "./GlobTool;./GrepTool;./LS;./View;./Bash"'
```

## Personnalisation de l'Instruction Système

Pour modifier l'instruction système par défaut, éditez le fichier `agent.go` :

```go
// Dans host/cliGCP/cmd/agent.go
genaimodels[model].SystemInstruction = &genai.Content{
    Role: "user",
    Parts: []genai.Part{
        genai.Text("Vous êtes un agent utile avec accès à des outils. " +
            "Votre travail est d'aider l'utilisateur en effectuant des tâches à l'aide de ces outils. " +
            "Vous ne devez pas inventer d'informations. " +
            "Si vous ne savez pas quelque chose, dites-le et expliquez ce que vous auriez besoin de savoir pour aider. " +
            "Si aucune indication, utilisez le répertoire de travail actuel qui est " + cwd),
    },
}
```

## Création de Configurations Spécifiques aux Tâches

Pour différents cas d'utilisation, vous pouvez créer des scripts de configuration spécialisés :

### Assistant de Revue de Code

Créez un fichier appelé `reviseur-code.sh` :

```bash
#!/bin/bash

export GCP_PROJECT=votre-id-de-projet-gcp
export GCP_REGION=us-central1
export GEMINI_MODELS=gemini-2.0-flash
export IMAGE_DIR=/tmp/images
export SYSTEM_INSTRUCTION="Vous êtes un expert en revue de code. Analysez le code pour détecter les bugs, les problèmes de sécurité et les domaines à améliorer. Concentrez-vous sur la fourniture de commentaires constructifs et d'explications détaillées."

cd /chemin/vers/gomcptest/bin
./cliGCP -mcpservers "./GlobTool;./GrepTool;./LS;./View;./Bash"
```

Rendez-le exécutable :

```bash
chmod +x reviseur-code.sh
```

### Générateur de Documentation

Créez un fichier appelé `generateur-doc.sh` :

```bash
#!/bin/bash

export GCP_PROJECT=votre-id-de-projet-gcp
export GCP_REGION=us-central1
export GEMINI_MODELS=gemini-2.0-flash
export IMAGE_DIR=/tmp/images
export SYSTEM_INSTRUCTION="Vous êtes un spécialiste de la documentation. Votre tâche est d'aider à créer une documentation claire et complète pour le code. Analysez la structure du code et créez une documentation appropriée en suivant les meilleures pratiques."

cd /chemin/vers/gomcptest/bin
./cliGCP -mcpservers "./GlobTool;./GrepTool;./LS;./View;./Bash;./Replace"
```

## Configurations Avancées d'Outils

### Configuration de dispatch_agent

Lorsque vous utilisez l'outil dispatch_agent, vous pouvez configurer son comportement avec des arguments supplémentaires :

```bash
./cliGCP -mcpservers "./GlobTool;./GrepTool;./LS;./View;./dispatch_agent -glob-path ./GlobTool -grep-path ./GrepTool -ls-path ./LS -view-path ./View -timeout 30s;./Bash;./Replace"
```

### Création de Combinaisons d'Outils

Vous pouvez créer des combinaisons d'outils spécialisées pour différentes tâches :

```bash
# Ensemble d'outils pour le développement web
./cliGCP -mcpservers "./GlobTool -include '*.{html,css,js}';./GrepTool;./LS;./View;./Bash;./Replace"

# Ensemble d'outils pour le développement Go
./cliGCP -mcpservers "./GlobTool -include '*.go';./GrepTool;./LS;./View;./Bash;./Replace"
```

## Résolution des Problèmes Courants

### Problèmes de Connexion au Modèle

Si vous rencontrez des difficultés pour vous connecter au modèle Gemini :

1. Vérifiez vos identifiants GCP :
```bash
gcloud auth application-default print-access-token
```

2. Vérifiez que l'API Vertex AI est activée :
```bash
gcloud services list --enabled | grep aiplatform
```

3. Vérifiez que votre projet a accès aux modèles que vous demandez

### Échecs d'Exécution d'Outils

Si les outils échouent à s'exécuter :

1. Assurez-vous que les chemins des outils sont corrects
2. Vérifiez que les outils sont exécutables
3. Vérifiez les problèmes de permission dans les répertoires auxquels vous accédez

### Optimisation des Performances

Pour de meilleures performances :

1. Utilisez des modèles d'outils plus spécifiques pour réduire la portée de la recherche
2. Envisagez de créer des agents spécialisés pour différentes tâches
3. Définissez une température plus basse pour des réponses plus déterministes