---
title: GoMCPTest
---

{{< blocks/cover title="GoMCPTest: Go playground for Model Context Protocol Experimentations" image_anchor="top" height="full" >}}
<a class="btn btn-lg btn-primary me-3 mb-4" href="docs/">
  Learn More <i class="fas fa-arrow-alt-circle-right ms-2"></i>
</a>
<a class="btn btn-lg btn-secondary me-3 mb-4" href="https://github.com/owulveryck/gomcptest">
  Download <i class="fab fa-github ms-2 "></i>
</a>
<p class="lead mt-5">A proof of concept for implementing the Model Context Protocol with custom-built tools</p>
{{< blocks/link-down color="info" >}}
{{< /blocks/cover >}}


{{% blocks/lead color="primary" %}}
gomcptest is a proof of concept (POC) demonstrating how to implement a Model Context Protocol (MCP) with a custom-built host to play with agentic systems. It provides tools for building and testing AI agents that can interact with their environment through function calling.
{{% /blocks/lead %}}


{{% blocks/section color="dark" type="row" %}}
{{% blocks/feature icon="fa-lightbulb" title="MCP Protocol Integration" %}}
gomcptest implements the Model Context Protocol (MCP) for building custom agentic systems that can interact with tools and their environment.

Check the [architecture documentation](/docs/explanation/architecture) for more details!
{{% /blocks/feature %}}


{{% blocks/feature icon="fab fa-github" title="Contributions welcome!" url="https://github.com/owulveryck/gomcptest" %}}
We do a [Pull Request](https://github.com/owulveryck/gomcptest/pulls) contributions workflow on **GitHub**. New users are always welcome!
{{% /blocks/feature %}}


{{% blocks/feature icon="fa-cogs" title="Extensible Tools" %}}
Create custom tools with our extensible architecture and API compatibility layers to build your own agents.

Check the [tools reference](/docs/reference/tools) for more information.
{{% /blocks/feature %}}


{{% /blocks/section %}}


{{% blocks/section %}}
## Documentation Structure
{.h1 .text-center}

Our documentation follows the [Divio Documentation Framework](https://documentation.divio.com/), organizing content into tutorials, how-to guides, reference, and explanation.
{.text-center}
{{% /blocks/section %}}


{{% blocks/section type="row" %}}

{{% blocks/feature icon="fa-graduation-cap" title="Tutorials" url="/docs/tutorials/" %}}
Learning-oriented content that takes you through a series of steps to complete a project. Perfect for beginners getting started with gomcptest.

Start with our [Getting Started tutorial](/docs/tutorials/getting-started/).
{{% /blocks/feature %}}

{{% blocks/feature icon="fa-tools" title="How-To Guides" url="/docs/how-to/" %}}
Problem-oriented content that guides you through the steps to address specific use cases and tasks with gomcptest.

Learn how to [create a custom tool](/docs/how-to/create-custom-tool/).
{{% /blocks/feature %}}

{{% blocks/feature icon="fa-book" title="Reference" url="/docs/reference/" %}}
Technical descriptions of the gomcptest components, APIs, and tools with comprehensive details.

Check our [Tools Reference](/docs/reference/tools/).
{{% /blocks/feature %}}

{{% blocks/feature icon="fa-lightbulb" title="Explanations" url="/docs/explanation/" %}}
Understanding-oriented content that explains concepts and provides context about how gomcptest works.

Read about the [MCP Protocol](/docs/explanation/mcp-protocol/).
{{% /blocks/feature %}}

{{% /blocks/section %}}


{{% blocks/section %}}
## Key Components
{.h1 .text-center}

gomcptest consists of host components like the OpenAI-compatible server and cliGCP, along with a variety of MCP-compatible tools that enable agent functionality.
{.text-center}
{{% /blocks/section %}}

{{% blocks/section type="row" %}}

{{% blocks/feature icon="fa-server" title="OpenAI-compatible server" %}}
A server that implements the OpenAI API interface and connects to Google's Vertex AI for model inference.

Located in `host/openaiserver`.
{{% /blocks/feature %}}

{{% blocks/feature icon="fa-terminal" title="cliGCP" %}}
A command-line interface similar to Claude Code or ChatGPT that interacts with Gemini models and MCP tools.

Located in `host/cliGCP`.
{{% /blocks/feature %}}

{{% blocks/feature icon="fa-tools" title="MCP Tools" %}}
Various tools that enable agent functionality:
- Bash, Edit, GlobTool, GrepTool
- LS, Replace, View
- dispatch_agent

Located in the `tools` directory.
{{% /blocks/feature %}}

{{% /blocks/section %}}
