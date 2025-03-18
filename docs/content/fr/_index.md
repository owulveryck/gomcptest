---
title: gomcptest
---

{{< blocks/cover title="GoMCPTest: Une aire de jeu en Go pour tester et POCer les systèmes MCP" image_anchor="top" height="full" >}}
<a class="btn btn-lg btn-primary me-3 mb-4" href="docs/">
  En savoir plus <i class="fas fa-arrow-alt-circle-right ms-2"></i>
</a>
<a class="btn btn-lg btn-secondary me-3 mb-4" href="https://github.com/owulveryck/gomcptest">
  Télécharger <i class="fab fa-github ms-2 "></i>
</a>
<p class="lead mt-5">Une preuve de concept pour implémenter le Protocole de Contexte de Modèle avec des outils personnalisés</p>
{{< blocks/link-down color="info" >}}
{{< /blocks/cover >}}


{{% blocks/lead color="primary" %}}
gomcptest est une preuve de concept (POC) démontrant comment implémenter un Protocole de Contexte de Modèle (MCP) avec un hôte personnalisé pour expérimenter avec des systèmes d'agents. Il fournit des outils pour construire et tester des agents IA qui peuvent interagir avec leur environnement via des appels de fonction.
{{% /blocks/lead %}}


{{% blocks/section color="dark" type="row" %}}
{{% blocks/feature icon="fa-lightbulb" title="Intégration du Protocole MCP" %}}
gomcptest implémente le Protocole de Contexte de Modèle (MCP) pour construire des systèmes d'agents personnalisés qui peuvent interagir avec des outils et leur environnement.

Consultez la [documentation d'architecture](/fr/docs/explanation/architecture) pour plus de détails !
{{% /blocks/feature %}}


{{% blocks/feature icon="fab fa-github" title="Contributions bienvenues !" url="https://github.com/owulveryck/gomcptest" %}}
Nous utilisons un processus de contribution par [Pull Request](https://github.com/owulveryck/gomcptest/pulls) sur **GitHub**. Les nouveaux utilisateurs sont toujours les bienvenus !
{{% /blocks/feature %}}


{{% blocks/feature icon="fa-cogs" title="Outils Extensibles" %}}
Créez des outils personnalisés avec notre architecture extensible et nos couches de compatibilité API pour construire vos propres agents.

Consultez la [référence des outils](/fr/docs/reference/tools) pour plus d'informations.
{{% /blocks/feature %}}


{{% /blocks/section %}}


{{% blocks/section %}}
## Structure de la Documentation
{.h1 .text-center}

Notre documentation suit le [Framework de Documentation Divio](https://documentation.divio.com/), organisant le contenu en tutoriels, guides pratiques, références et explications.
{.text-center}
{{% /blocks/section %}}


{{% blocks/section type="row" %}}

{{% blocks/feature icon="fa-graduation-cap" title="Tutoriels" url="/fr/docs/tutorials/" %}}
Contenu orienté apprentissage qui vous guide à travers une série d'étapes pour compléter un projet. Parfait pour les débutants qui commencent avec gomcptest.

Commencez avec notre [tutoriel de démarrage](/fr/docs/tutorials/getting-started/).
{{% /blocks/feature %}}

{{% blocks/feature icon="fa-tools" title="Guides Pratiques" url="/fr/docs/how-to/" %}}
Contenu orienté problème qui vous guide à travers les étapes pour répondre à des cas d'utilisation et des tâches spécifiques avec gomcptest.

Apprenez à [créer un outil personnalisé](/fr/docs/how-to/create-custom-tool/).
{{% /blocks/feature %}}

{{% blocks/feature icon="fa-book" title="Référence" url="/fr/docs/reference/" %}}
Descriptions techniques des composants, APIs et outils de gomcptest avec des détails complets.

Consultez notre [Référence des Outils](/fr/docs/reference/tools/).
{{% /blocks/feature %}}

{{% blocks/feature icon="fa-lightbulb" title="Explications" url="/fr/docs/explanation/" %}}
Contenu orienté compréhension qui explique les concepts et fournit du contexte sur le fonctionnement de gomcptest.

Lisez à propos du [Protocole MCP](/fr/docs/explanation/mcp-protocol/).
{{% /blocks/feature %}}

{{% /blocks/section %}}


{{% blocks/section %}}
## Composants Clés
{.h1 .text-center}

gomcptest se compose de composants hôtes comme le serveur compatible OpenAI et cliGCP, ainsi que d'une variété d'outils compatibles MCP qui permettent la fonctionnalité d'agent.
{.text-center}
{{% /blocks/section %}}

{{% blocks/section type="row" %}}

{{% blocks/feature icon="fa-server" title="Serveur compatible OpenAI" %}}
Un serveur qui implémente l'interface API OpenAI et se connecte à Vertex AI de Google pour l'inférence de modèle.

Situé dans `host/openaiserver`.
{{% /blocks/feature %}}

{{% blocks/feature icon="fa-terminal" title="cliGCP" %}}
Une interface en ligne de commande similaire à Claude Code ou ChatGPT qui interagit avec les modèles Gemini et les outils MCP.

Située dans `host/cliGCP`.
{{% /blocks/feature %}}

{{% blocks/feature icon="fa-tools" title="Outils MCP" %}}
Divers outils qui permettent la fonctionnalité d'agent :
- Bash, Edit, GlobTool, GrepTool
- LS, Replace, View
- dispatch_agent

Situés dans le répertoire `tools`.
{{% /blocks/feature %}}

{{% /blocks/section %}}
