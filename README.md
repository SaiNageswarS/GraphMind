# 🧠 GraphMind

> Semantic Graphs for Cross-System Code Understanding & Generation

Modern tech systems are sprawling — dozens of microservices, databases, cloud resources like File Storages, Key Vaults, and more. Even a **tiny change** often means updating **multiple codebases** and touching **many interconnected systems**.

This slows down dev team to build new ideas/products while catering to changes in the existing system. Changes to existing system takes away around 70% of Dev Effort in a large org.

**GraphMind** is designed to break that bottleneck.

---

## 🔍 What is GraphMind?

GraphMind is a semantic-aware code intelligence tool purpose-built for **Golang/gRPC-based distributed systems**.

It builds a **unified semantic graph** across all your code repositories, enabling **LLMs to understand and modify your system** from a single specification.

---

## 🧠 How It Works

1. **Repo Analysis**  
   Parses multiple codebases to identify microservice entry points and APIs (especially gRPC in Golang).

2. **Semantic ASTs via LLMs**  
   For each API/method, GraphMind builds an AST. Then it uses **Claude** to semantically annotate the AST — identifying not just *what* functions do, but *why* and *what they touch*.  

   > Example: AST may show an HTTP call — but Claude can infer the target service or resource from URLs or variable names, giving context that ASTs alone miss.

3. **Semantic Graph Construction**  
   Merges all annotated ASTs into a unified **Semantic Graph** using `rdflib`. This cross-repo graph represents a complete view of your system: services, APIs, resources, and dependencies.

   ✅ The semantic graph construction has been successfully tested on the following real-world microservice repositories:
   - [`authGo`](https://github.com/Kotlang/authGo)
   - [`notificationGo`](https://github.com/Kotlang/notificationGo)
   - [`socialGo`](https://github.com/Kotlang/socialGo)
   - [`localizationGo`](https://github.com/Kotlang/localizationGo)

4. **LLM-Powered Code Generation**  
   With this graph, an LLM can:
   - Understand a natural language spec
   - Identify affected APIs/services/resources
   - Generate accurate **multi-repo code changes**
   - Output change summaries and Mermaid.js diagrams to depict code change path.

## 🛠️ Getting Started

```bash
git clone https://github.com/SaiNageswarS/GraphMind.git
cd GraphMind
# Set up Golang and Python environments
# Provide API keys for Claude/OpenAI in .env as per .env.template
# Run temporal server
temporal server start-dev
# build golang worker
.\build.ps1 
# Run worker
.\build\GraphMind
# Run python rdflib server
cd GraphMindPyAPIs
python main.py


