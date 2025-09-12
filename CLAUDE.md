# Claude AI Assistant Context

This document provides context about the gomcptest project for AI assistants like Claude.

## Project Overview

**gomcptest** is a proof of concept (POC) demonstrating how to implement a Model Context Protocol (MCP) with a custom-built host for testing agentic systems. The codebase is primarily written from scratch in Go to provide clear understanding of the underlying mechanisms.

## Key Components

### Host Applications
- **`host/openaiserver`**: Custom OpenAI-compatible API server using Google Gemini with embedded chat UI
- **`host/cliGCP`**: CLI tool similar to Claude Code for testing agentic interactions
- **`host/openaiserver/simpleui`**: Standalone UI server that provides a web-based chat interface

### MCP Tools (in `tools/` directory)
- **Bash**: Execute bash commands
- **Edit**: Edit file contents
- **GlobTool**: Find files matching glob patterns
- **GrepTool**: Search file contents with regular expressions
- **LS**: List directory contents
- **Replace**: Replace entire file contents
- **View**: View file contents
- **dispatch_agent**: Specialized agent dispatcher for various tasks
- **imagen**: Image generation and manipulation using Google Imagen
- **duckdbserver**: DuckDB server for data processing

## Build System

Use the root Makefile to build all tools and servers:
```bash
# Build all tools and servers
make all

# Build only tools
make tools

# Build only servers  
make servers

# Run a specific tool for testing
make run TOOL=Bash

# Install binaries to a directory
make install INSTALL_DIR=/path/to/install
```

## Configuration

Environment variables are used for configuration:
- `GCP_PROJECT`: Google Cloud Project ID
- `GCP_REGION`: Google Cloud Region (default: us-central1)
- `GEMINI_MODELS`: Comma-separated list of Gemini models
- `LOG_LEVEL`: Logging level (DEBUG, INFO, WARN, ERROR)
- `PORT`: Server port (default: 8080)

**Note**: `IMAGEN_MODELS` and `IMAGE_DIR` are no longer needed for the hosts as imagen functionality is now provided by the independent MCP tool.

## Testing

Tests are available for most components. Use standard Go testing:
```bash
go test ./...
```

## Web UI

The project includes a modern web-based chat interface called **AgentFlow** for interacting with the agentic system:

### UI Access Methods

1. **Embedded UI** (via main openaiserver):
   ```bash
   # Start the main server
   ./bin/openaiserver
   # Access UI at: http://localhost:8080/ui
   ```

2. **Standalone UI Server** (via simpleui):
   ```bash
   # Start the main API server
   ./bin/openaiserver -port=4000
   
   # In another terminal, start the UI server
   cd host/openaiserver/simpleui
   go run . -ui-port=8081 -api-url=http://localhost:4000
   # Access UI at: http://localhost:8081
   ```

### UI Features

- **Mobile-optimized**: Responsive design with mobile web app capabilities
- **Real-time chat**: Streaming responses with proper SSE handling
- **Modern interface**: Clean, professional design with gradient backgrounds
- **Template-based**: Uses Go templates for flexible configuration
- **CORS support**: Proper cross-origin headers for API communication

### UI Configuration

The UI server supports the following options:
- `-ui-port`: Port to serve the UI (default: 8080)
- `-api-url`: OpenAI server API URL to proxy requests to
- `OPENAISERVER_URL`: Environment variable for API URL (default: http://localhost:4000)

### Template Architecture

The UI uses a Go template system (`chat-ui.html.tmpl`) that receives a `BaseURL` parameter:
- When served by main openaiserver (`/ui` endpoint): `BaseURL` is empty (same server)
- When served by simpleui server: `BaseURL` points to the separate API server

## Safety Considerations

⚠️ **WARNING**: These tools can execute commands and modify files. Use in a sandboxed environment when possible.

## Documentation

Comprehensive documentation is available at: https://owulveryck.github.io/gomcptest/

The documentation is auto-generated using Hugo and includes:
- Architecture explanations
- How-to guides
- Tutorials
- Reference documentation

## Current State

The project is actively maintained with recent commits focusing on:
- **AgentFlow UI**: Modern web-based chat interface with mobile optimization
- **Template system**: Flexible UI template architecture supporting multiple deployment modes
- Comprehensive Imagen tool suite with HTTP server
- Rationalized build system with single root Makefile
- Better logging mechanisms
- Package updates
- Resource management improvements

## Usage for AI Assistants

When working with this codebase:
1. Tools are MCP-compatible and can be composed together
2. The project follows Go conventions and module structure
3. Each tool has its own README with specific usage instructions
4. Tests provide good examples of expected behavior
5. The host applications demonstrate how to integrate with external APIs (Google Gemini)
6. **UI testing**: Use the simpleui server for isolated UI development and testing
7. **Template modifications**: The chat UI template supports both embedded and standalone modes
- The documentation present in @docs/ follows this framework:
# **Tutorials**

Tutorials are *lessons* that take the reader by the hand through a series of steps to complete a project of some kind. They are what your project needs in order to show a beginner that they can achieve something with it.

They are wholly learning-oriented, and specifically, they are oriented towards *learning how* rather than *learning that*.

You are the teacher, and you are responsible for what the student will do. Under your instruction, the student will execute a series of actions to achieve some end.

The end and the actions are up to you, but deciding what they should be can be hard work. The end has to be *meaningful*, but also *achievable* for a complete beginner.

The important thing is that having done the tutorial, the learner is in a position to make sense of the rest of the documentation, and the software itself.

Most software projects have really bad \- or non-existent \- tutorials. Tutorials are what will turn your learners into users. A bad or missing tutorial will prevent your project from acquiring new users.

Of the sections describing the four kinds of documentation, this is by far the longest \- that's because tutorials are the most misunderstood and most difficult to do well. The best way of teaching is to have a teacher present, interacting with the student. That's rarely possible, and our written tutorials will be at best a far-from-perfect substitute. That's all the more reason to pay special attention to them.

Tutorials need to be useful for the beginner, easy to follow, meaningful and extremely robust, and kept up-to-date. You might well find that writing and maintaining your tutorials can occupy as much time and energy as the other three parts put together.

## **Analogy from cooking**

Consider an analogy of teaching a child to cook.

*What* you teach the child to cook isn’t really important. What’s important is that the child finds it enjoyable, and gains confidence, and wants to do it again.

*Through* the things the child does, it will learn important things about cooking. It will learn what it is like to be in the kitchen, to use the utensils, to handle the food.

This is because using software, like cooking, is a matter of craft. It’s knowledge \- but it is *practical* knowledge, not *theoretical* knowledge.

When we learn a new craft or skill, we always begin learning it by doing.

## **How to write good tutorials**

### **Allow the user to learn by doing**

In the beginning, we only learn anything by doing \- it’s how we learn to talk, or walk.

In your software tutorial, your learner needs to *do* things. The different things that they do while following your tutorial need to cover a wide range of tools and operations, building up from the simplest ones at the start to more complex ones.

### **Get the user started**

It’s perfectly acceptable if your beginner’s first steps are hand-held baby steps. It’s also perfectly acceptable if what you get the beginner to do is not the way an experienced person would, or even if it’s not the ‘correct’ way \- a tutorial for beginners is not the same thing as a manual for best practice.

The point of a tutorial is to get your learner started on their journey, not to get them to a final destination.

### **Make sure that your tutorial works**

One of your jobs as a tutor is to inspire the beginner’s confidence: in the software, in the tutorial, in the tutor and, of course, in their own ability to achieve what’s being asked of them.

There are many things that contribute to this. A friendly tone helps, as does consistent use of language, and a logical progression through the material. But the single most important thing is that what you ask the beginner to do must work. The learner needs to see that the actions you ask them to take have the effect you say they will have.

If the learner's actions produce an error or unexpected results, your tutorial has failed \- even if it’s not your fault. When your students are there with you, you can rescue them; if they’re reading your documentation on their own you can’t \- so you have to prevent that from happening in advance. This is without doubt easier said than done.

### **Ensure the user sees results immediately**

Everything the learner does should accomplish something comprehensible, however small. If your student has to do strange and incomprehensible things for two pages before they even see a result, that’s much too long. The effect of every action should be visible and evident as soon as possible, and the connection to the action should be clear.

The conclusion of each section of a tutorial, or the tutorial as a whole, must be a meaningful accomplishment.

### **Make your tutorial repeatable**

Your tutorial must be reliably repeatable. This not easy to achieve: people will be coming to it with different operating systems, levels of experience and tools. What’s more, any software or resources they use are quite likely themselves to change in the meantime.

The tutorial has to work for all of them, every time.

Tutorials unfortunately need regular and detailed testing to make sure that they still work.

### **Focus on concrete steps, not abstract concepts**

Tutorials need to be concrete, built around specific, particular actions and outcomes.

The temptation to introduce abstraction is huge; it is after all how most computing derives its power. But all learning proceeds from the particular and concrete to the general and abstract, and asking the learner to appreciate levels of abstraction before they have even had a chance to grasp the concrete is poor teaching.

### **Provide the minimum necessary explanation**

Don’t explain anything the learner doesn’t need to know in order to complete the tutorial. Extended discussion is important \- just not in a tutorial. In a tutorial, it is an obstruction and a distraction. Only the bare minimum is appropriate. Instead, link to explanations elsewhere in the documentation.

### **Focus only on the steps the user needs to take**

Your tutorial needs to be focused on the task in hand. Maybe the command you’re introducing has many other options, or maybe there are different ways to access a certain API. It doesn’t matter: right now, your learner does not need to know about those in order to make progress.

# **Explanation**

Explanation, or discussions, *clarify and illuminate a particular topic*. They broaden the documentation’s coverage of a topic.

They are understanding-oriented.

Explanations can equally well be described as *discussions*; they are discursive in nature. They are a chance for the documentation to relax and step back from the software, taking a wider view, illuminating it from a higher level or even from different perspectives. You might imagine a discussion document being read at leisure, rather than over the code.

This section of documentation is rarely explicitly created, and instead, snippets of explanation are scattered amongst other sections. Sometimes, the section exists, but has a name such as *Background* or *Other notes* or *Key topics* \- these names are not always useful.

Discussions are less easy to create than it might seem \- things that are straightforward to explain when you have the starting-point of someone’s question are less easy when you have a blank page and have to write down something about it.

A topic isn’t defined by a specific task you want to achieve, like a how-to guide, or what you want the user to learn, like a tutorial. It’s not defined by a piece of the machinery, like reference material. It’s defined by what you think is a reasonable area to try to cover at one time, so the division of topics for discussion can sometimes be a little arbitrary.

## **Analogy from cooking**

Think about a work that discusses food and cooking in the context of history, science and technology. It's *about* cooking and the kitchen.

It doesn't teach, it's not a collection of recipes, and it doesn't just describe.

Instead, it analyses, considers things from multiple perspectives. It might explain why it is we now do things the way we do, or even describe bad ways of doing things, or obscure alternatives.

It deepens our knowledge and makes it richer, even if it isn't knowledge we can actually apply in any practical sense \- but it doesn't need to be, in order to be valuable.

It's something we might read at our leisure, away from the kitchen itself, when we want to think about cooking at a higher level, and to understand more about the subject.

## **How to write a good explanation**

### **Provide context**

Explanations are the place for background and context \- for example, *Web forms and how they are handled in Django*, or *Search in django CMS*.

They can also explain *why* things are so \- design decisions, historical reasons, technical constraints.

### **Discuss alternatives and opinions**

Explanation can consider alternatives, or multiple different approaches to the same question. For example, in an article on Django deployment, it would be appropriate to consider and evaluate different web server options.

Discussions can even consider and weigh up contrary *opinions* \- for example, whether test modules should be in a package directory, or not.

### **Don't instruct, or provide technical reference**

Explanation should do things that the other parts of the documentation do not. It’s not the place of an explanation to instruct the user in how to do something. Nor should it provide technical description. These functions of documentation are already taken care of in other sections.

# **How-to guides**

How-to guides take the reader through the steps required to solve a real-world problem.

They are recipes, directions to achieve a specific end \- for example: *how to create a web form*; *how to plot a three-dimensional data-set*; *how to enable LDAP authentication*.

They are wholly goal-oriented.

How-to guides are wholly distinct from tutorials and must not be confused with them:

* A tutorial is what you decide a beginner needs to know.  
* A how-to guide is an answer to a question that only a user with some experience could even formulate.

In a how-to guide, you can assume some knowledge and understanding. You can assume that the user already knows how to do basic things and use basic tools.

Unlike tutorials, how-to guides in software documentation tend to be done fairly well. They’re also fun and easy to write.

## **Analogy from cooking**

Think about a recipe, for preparing something to eat.

A recipe has a clear, defined end. It addresses a specific question. It shows someone \- who can be assumed to have some basic knowledge already \- how to achieve something.

Someone who has never cooked before can't be expected to follow a recipe with success, so a recipe is not a substitute for a cooking lesson. At the same time, someone who reads a recipe would be irritated to find that it tries to teach basics that they know already, or contains irrelevant discussion of the ingredients.

## **How to write good how-to guides**

### **Provide a series of steps**

How-to guides must contain a list of steps, that need to be followed in order (just like tutorials do). You don’t have to start at the very beginning, just at a reasonable starting point. How-to guides should be reliable, but they don’t need to have the cast-iron repeatability of a tutorial.

### **Focus on results**

How-to guides must focus on achieving a practical goal. Anything else is a distraction. As in tutorials, detailed explanations are out of place here.

### **Solve a particular problem**

A how-to guide must address a specific question or problem: *How do I …?*

This is one way in which how-to guides are distinct from tutorials: when it comes to a how-to guide, the reader can be assumed to know *what* they should achieve, but don’t yet know *how* \- whereas in the tutorial, *you* are responsible for deciding what things the reader needs to know about.

### **Don't explain concepts**

A how-to guide should not explain things. It’s not the place for discussions of that kind; they will simply get in the way of the action. If explanations are important, link to them.

### **Allow for some flexibility**

A how-to guide should allow for slightly different ways of doing the same thing. It needs just enough flexibility in it that the user can see how it will apply to slightly different examples from the one you describe, or understand how to adapt it to a slightly different system or configuration from the one you’re assuming. Don’t be so specific that the guide is useless for anything except the exact purpose you have in mind.

### **Leave things out**

Practical usability is more valuable than completeness. Tutorials need to be complete, end-to-end guides; how-to guides do not. They can start and end where it seems appropriate to you. They don’t need to mention everything that there is to mention either, just because it is related to the topic. A bloated how-to guide doesn’t help the user get speedily to their solution.

### **Name guides well**

The title of a how-to document should tell the user exactly what it does. *How to create a class-based view* is a good title. *Creating a class-based view* or worse, *Class-based views*, are not.

# **Reference guides**

Reference guides are *technical descriptions of the machinery* and how to operate it.

Reference guides have one job only: to describe. They are code-determined, because ultimately that's what they describe: key classes, functions, APIs, and so they should list things like functions, fields, attributes and methods, and set out how to use them.

Reference material is information-oriented.

By all means technical reference can contain examples to illustrate usage, but it should not attempt to explain basic concepts, or how to achieve common tasks.

Reference material should be simple and to the point.

Note that description does include basic description of how to use the machinery \- how to instantiate a particular class, or invoke a certain method, for example, or precautions that must be taken when passing something to a function. However this is simply part of its function as technical reference, and emphatically not to be confused with a how-to guide \- *describing correct usage of software* (technical reference) is not the same as *showing how to use it to achieve a certain end* (how-to documentation).

For some developers, reference guides are the only kind of documentation they can imagine. They already understand their software, they know how to use it. All they can imagine that other people might need is technical information about it.

Reference material tends to be written well. It can even \- to some extent \- be generated automatically, but this is never sufficient on its own.

## **Analogy from cooking**

Consider an encyclopedia article about an ingredient, say ginger.

When you look up *ginger* in a reference work, what you want is *information* about the ingredient \- information describing its provenance, its behaviour, its chemical constituents, how it can be cooked.

You expect that whatever ingredient you look up, the information will be presented in a similar way. And you expect to be informed of basic facts, such as *ginger is a member of the family that includes turmeric and cardamom*.

This is also where you'd expect to be alerted about potential problems, such as: *ginger is known to provoke heartburn in some individuals* or: *ginger may interfere with the effects of anticoagulants, such as warfarin or aspirin*.

## **How to write good reference guides**

### **Structure the documentation around the code**

Give reference documentation the same structure as the codebase, so that the user can navigate both the code and the documentation for it at the same time. This will also help the maintainers see where reference documentation is missing or needs to be updated.

### **Be consistent**

In reference guides, structure, tone, format must all be consistent \- as consistent as those of an encyclopaedia or dictionary.

### **Do nothing but describe**

The only job of technical reference is to describe, as clearly and completely as possible. Anything else (explanation, discussion, instruction, speculation, opinion) is not only a distraction, but will make it harder to use and maintain. Provide examples to illustrate the description when appropriate.

Avoid the temptation to use reference material to instruct in how to achieve things, beyond the basic scope of using the software, and don’t allow explanations of concepts or discussions of topics to develop. Instead, link to how-to guides, explanation and introductory tutorials as appropriate.

### **Be accurate**

These descriptions must be accurate and kept up-to-date. Any discrepancy between the machinery and your description of it will inevitably lead a user astray.