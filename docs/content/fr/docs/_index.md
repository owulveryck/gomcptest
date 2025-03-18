---
title: Documentation
linkTitle: Docs
menu: {main: {weight: 20}}
---

{{% pageinfo %}}
gomcptest est une preuve de concept (POC) démontrant comment implémenter un Protocole de Contexte de Modèle (MCP) avec un hôte personnalisé pour expérimenter avec des systèmes d'agents.
{{% /pageinfo %}}

# Documentation gomcptest

Bienvenue dans la documentation de gomcptest. Ce projet est une preuve de concept (POC) démontrant comment implémenter un Protocole de Contexte de Modèle (MCP) avec un hôte personnalisé pour expérimenter avec des systèmes d'agents.

## Structure de la Documentation

Notre documentation suit le [Framework de Documentation Divio](https://documentation.divio.com/), qui organise le contenu en quatre types distincts : tutoriels, guides pratiques, référence et explication. Cette approche garantit que différents besoins d'apprentissage sont satisfaits avec le format de contenu approprié.

## Tutoriels : Contenu orienté apprentissage

Les tutoriels sont des leçons qui vous guident pas à pas à travers une série d'étapes pour réaliser un projet. Ils se concentrent sur l'apprentissage par la pratique et aident les débutants à se familiariser avec le système.

| Tutoriel | Description |
|----------|-------------|
| [Démarrage avec gomcptest](/fr/docs/tutorials/getting-started/) | Un guide complet pour les débutants sur la configuration de l'environnement, la création d'outils et l'exécution de votre premier agent. Parfait pour les nouveaux utilisateurs. |
| [Construire Votre Premier Serveur Compatible OpenAI](/fr/docs/tutorials/openaiserver-tutorial/) | Instructions étape par étape pour exécuter et configurer le serveur compatible OpenAI qui communique avec les modèles LLM et exécute les outils MCP. |
| [Utiliser l'Interface en Ligne de Commande cliGCP](/fr/docs/tutorials/cligcp-tutorial/) | Guide pratique pour configurer et utiliser l'outil cliGCP pour interagir avec les LLMs et effectuer des tâches à l'aide des outils MCP. |

## Guides Pratiques : Contenu orienté problème

Les guides pratiques sont des recettes qui vous guident à travers les étapes impliquées dans la résolution de problèmes clés et de cas d'utilisation. Ils sont pragmatiques et orientés vers des objectifs.

| Guide Pratique | Description |
|--------------|-------------|
| [Comment Créer un Outil MCP Personnalisé](/fr/docs/how-to/create-custom-tool/) | Étapes pratiques pour créer un nouvel outil personnalisé compatible avec le Protocole de Contexte de Modèle, y compris des modèles de code et des exemples. |
| [Comment Configurer le Serveur Compatible OpenAI](/fr/docs/how-to/configure-openaiserver/) | Solutions pour configurer et personnaliser le serveur OpenAI pour différents cas d'utilisation, y compris les variables d'environnement, la configuration des outils et la configuration de production. |
| [Comment Configurer l'Interface en Ligne de Commande cliGCP](/fr/docs/how-to/configure-cligcp/) | Guides pour personnaliser l'outil cliGCP avec des variables d'environnement, des arguments de ligne de commande et des configurations spécialisées pour différentes tâches. |

## Référence : Contenu orienté information

Les guides de référence sont des descriptions techniques des mécanismes et de leur fonctionnement. Ils décrivent en détail comment les choses fonctionnent et sont précis et complets.

| Référence | Description |
|-----------|-------------|
| [Référence des Outils](/fr/docs/reference/tools/) | Référence complète de tous les outils compatibles MCP disponibles, leurs paramètres, formats de réponse et gestion des erreurs. |
| [Référence du Serveur Compatible OpenAI](/fr/docs/reference/openaiserver/) | Documentation technique de l'architecture du serveur, des points d'accès API, des options de configuration et des détails d'intégration avec Vertex AI. |
| [Référence cliGCP](/fr/docs/reference/cligcp/) | Référence détaillée de la structure de commande cliGCP, des composants, des paramètres, des modèles d'interaction et des états internes. |

## Explication : Contenu orienté compréhension

Les documents d'explication discutent et clarifient les concepts pour élargir la compréhension du lecteur sur les sujets. Ils fournissent du contexte et éclairent les idées.

| Explication | Description |
|-------------|-------------|
| [Architecture de gomcptest](/fr/docs/explanation/architecture/) | Plongée profonde dans l'architecture du système, les décisions de conception et comment les différents composants interagissent pour créer un hôte MCP personnalisé. |
| [Comprendre le Protocole de Contexte de Modèle (MCP)](/fr/docs/explanation/mcp-protocol/) | Exploration de ce qu'est le MCP, comment il fonctionne, les décisions de conception qui le sous-tendent et comment il se compare aux approches alternatives pour l'intégration d'outils LLM. |

## Composants du Projet

gomcptest se compose de plusieurs composants clés qui fonctionnent ensemble :

### Composants Hôtes

- **Serveur compatible OpenAI** (`host/openaiserver`) : Un serveur qui implémente l'interface API OpenAI et se connecte à Vertex AI de Google pour l'inférence de modèle.
- **cliGCP** (`host/cliGCP`) : Une interface en ligne de commande similaire à Claude Code ou ChatGPT qui interagit avec les modèles Gemini et les outils MCP.

### Outils

Le répertoire `tools` contient divers outils compatibles MCP :

- **Bash** : Exécute des commandes bash dans une session shell persistante
- **Edit** : Modifie le contenu d'un fichier en remplaçant un texte spécifié
- **GlobTool** : Trouve des fichiers correspondant à des modèles glob
- **GrepTool** : Recherche dans le contenu des fichiers à l'aide d'expressions régulières
- **LS** : Liste les fichiers et répertoires
- **Replace** : Remplace complètement le contenu d'un fichier
- **View** : Lit le contenu des fichiers
- **dispatch_agent** : Lance un nouvel agent avec accès à des outils spécifiques