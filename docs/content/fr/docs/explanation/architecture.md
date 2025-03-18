---
title: "Architecture de gomcptest"
linkTitle: "Architecture"
weight: 1
description: >
  Plongée profonde dans l'architecture du système et les décisions de conception
---

{{% pageinfo %}}
Ce document explique l'architecture de gomcptest, les décisions de conception qui la sous-tendent, et comment les différents composants interagissent pour créer un hôte personnalisé pour le Protocole de Contexte de Modèle (MCP).
{{% /pageinfo %}}

## Vue d'Ensemble

Le projet gomcptest implémente un hôte personnalisé qui fournit une implémentation du Protocole de Contexte de Modèle (MCP). Il est conçu pour permettre les tests et l'expérimentation avec des systèmes d'agents sans nécessiter d'intégration directe avec des plateformes LLM commerciales.

Le système est construit avec ces principes clés à l'esprit :
- Modularité : Les composants sont conçus pour être interchangeables
- Compatibilité : L'API imite l'API OpenAI pour une intégration facile
- Extensibilité : De nouveaux outils peuvent être facilement ajoutés au système
- Test : L'architecture facilite le test d'applications basées sur des agents

## Composants Principaux

### Hôte (Serveur OpenAI)

L'hôte est le composant central, situé dans `/host/openaiserver`. Il présente une interface API compatible OpenAI et se connecte à Vertex AI de Google pour l'inférence de modèle. Cette couche de compatibilité facilite l'intégration avec des outils et bibliothèques existants conçus pour OpenAI.

L'hôte a plusieurs responsabilités clés :
1. **Compatibilité API** : Implémentation de l'API de complétion de chat OpenAI
2. **Gestion des Sessions** : Maintien de l'historique et du contexte des conversations
3. **Intégration des Modèles** : Connexion aux modèles Gemini de Vertex AI
4. **Appel de Fonctions** : Orchestration des appels de fonctions/outils basés sur les sorties du modèle
5. **Streaming de Réponses** : Support des réponses en streaming vers le client

Contrairement aux implémentations commerciales, cet hôte est conçu pour le développement et les tests locaux, privilégiant la flexibilité et l'observabilité aux fonctionnalités prêtes pour la production comme l'authentification ou la limitation de débit.

### Outils MCP

Les outils sont des exécutables autonomes qui implémentent le Protocole de Contexte de Modèle. Chaque outil est conçu pour exécuter une fonction spécifique, comme l'exécution de commandes shell ou la manipulation de fichiers.

Les outils suivent un modèle cohérent :
- Ils communiquent via l'entrée/sortie standard en utilisant le protocole JSON-RPC MCP
- Ils exposent un ensemble spécifique de paramètres
- Ils gèrent leurs propres conditions d'erreur
- Ils renvoient les résultats dans un format standardisé

Cette approche permet aux outils d'être :
- Développés indépendamment
- Testés de manière isolée
- Utilisés dans différents environnements hôtes
- Chaînés ensemble dans des flux de travail complexes

### CLI

L'interface en ligne de commande fournit une interface utilisateur similaire à des outils comme "Claude Code" ou "OpenAI ChatGPT". Elle se connecte au serveur compatible OpenAI et offre un moyen d'interagir avec le LLM et les outils via une interface conversationnelle.

## Flux de Données

1. L'utilisateur envoie une requête à l'interface CLI
2. Le CLI transmet cette requête au serveur compatible OpenAI
3. Le serveur envoie la requête au modèle Gemini de Vertex AI
4. Le modèle peut identifier des appels de fonction dans sa réponse
5. Le serveur exécute ces appels de fonction en invoquant les outils MCP appropriés
6. Les résultats sont fournis au modèle pour poursuivre sa réponse
7. La réponse finale est diffusée en streaming vers le CLI et présentée à l'utilisateur

## Explications des Décisions de Conception

### Pourquoi la Compatibilité avec l'API OpenAI ?

L'API OpenAI est devenue un standard de facto dans l'espace des LLM. En implémentant cette interface, gomcptest peut fonctionner avec une grande variété d'outils, de bibliothèques et d'interfaces existants avec une adaptation minimale.

### Pourquoi Google Vertex AI ?

Vertex AI donne accès aux modèles Gemini de Google, qui possèdent de solides capacités d'appel de fonctions. L'implémentation pourrait être étendue pour prendre en charge d'autres fournisseurs de modèles si nécessaire.

### Pourquoi des Outils Autonomes ?

En implémentant les outils comme des exécutables autonomes plutôt que des fonctions de bibliothèque, nous obtenons plusieurs avantages :
- Sécurité par isolation
- Agnosticisme linguistique (les outils peuvent être écrits dans n'importe quel langage)
- Capacité à distribuer les outils séparément de l'hôte
- Tests et développement plus faciles

### Pourquoi MCP ?

Le Protocole de Contexte de Modèle fournit une manière standardisée pour les LLM d'interagir avec des outils externes. En adoptant ce protocole, gomcptest assure la compatibilité avec les outils développés pour d'autres hôtes compatibles MCP.

## Limitations et Orientations Futures

L'implémentation actuelle présente plusieurs limitations :
- Une seule session de chat par instance
- Support limité pour l'authentification et l'autorisation
- Pas de persistance de l'historique des chats entre les redémarrages
- Pas de support intégré pour la limitation de débit ou les quotas

Les améliorations futures pourraient inclure :
- Support pour plusieurs sessions de chat
- Intégration avec des fournisseurs de modèles supplémentaires
- Fonctionnalités de sécurité améliorées
- Gestion des erreurs et journalisation améliorées
- Optimisations de performance pour les déploiements à grande échelle

## Conclusion

L'architecture de gomcptest représente une approche flexible et extensible pour construire des hôtes MCP personnalisés. Elle privilégie la simplicité, la modularité et l'expérience développeur, ce qui en fait une excellente plateforme pour l'expérimentation avec des systèmes d'agents.

En comprenant cette architecture, les développeurs peuvent utiliser efficacement le système, l'étendre avec de nouveaux outils, et potentiellement l'adapter à leurs besoins spécifiques.