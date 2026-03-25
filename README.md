# Omni Knowledge Base Assistant

> **Note on Project Maturity:** This document outlines the current vision and architectural blueprint for the Omni platform. All features, implementation details, and release timelines are subject to change as we gather feedback from early adopters and refine the technology.

## Table of Contents
1. [Executive Summary & Concept](#1-executive-summary--concept)
2. [The Problem Statement](#2-the-problem-statement)
3. [Value Proposition & Business Model](#3-value-proposition--business-model)
4. [Product Packaging & Deployment](#4-product-packaging--deployment)
5. [User Experience & Multi-Tenant Architecture](#5-user-experience--multi-tenant-architecture)
6. [Target Use Cases](#6-target-use-cases)
7. [Integrations & Extensibility](#7-integrations--extensibility)
8. [Technical Architecture & Security](#8-technical-architecture--security)
9. [Future Extensions (Roadmap)](#9-future-extensions-roadmap)
10. [Service & Support Policies](#10-service--support-policies)
11. [MVP Development Plan](#11-mvp-development-plan)

---

## 1. Executive Summary & Concept
**Omni Knowledge Base Assistant** is an intelligent, multimodal knowledge platform that functions as a proactive digital assistant for corporate teams. It securely aggregates scattered organizational data (documents, text, files, audio, video) and enables intelligent querying via Retrieval-Augmented Generation (RAG). 

Instead of passively waiting for search queries, the Omni system acts as an active participant in daily workflows: answering contextual questions via text or voice, reminding users of deadlines, and automating business processes safely through secure, privacy-preserving communication channels.

## 2. The Problem Statement
Organizations struggle with the rapid accumulation of unstructured knowledge. Critical challenges include:
* **Fragmented Information:** Data is scattered across PDFs, internal tools, emails, media files, and team conversations.
* **Poor Searchability:** Traditional keyword search fails to understand context or intent.
* **Security & Privacy Risks:** Companies cannot upload sensitive data to public cloud AI tools. **Standard messengers also pose a risk**, as data often passes through third-party servers.
* **Adoption Complexity:** SMBs lack AI expertise, and users are reluctant to adopt standalone dashboards.

## 3. Value Proposition & Business Model
* **Uncompromising Data Security:** Omni guarantees that sensitive company data never leaks to public networks, providing full compliance with internal policies.
* **Sovereign Communication:** By supporting peer-to-peer (P2P) and abstract messenger protocols, Omni ensures that even the *interaction* with the AI remains within the client's control.
* **Professional Integration Services:** Our team builds tailored connectors and automated voice-to-action workflows for specific enterprise needs.

## 4. Product Packaging & Deployment
To accommodate different infrastructure capabilities, Omni is delivered in two distinct models:

### Hardware Appliance ("AI-in-a-Box")
A pre-configured physical server (with optional local GPU support) shipped directly to the client. It offers true plug-and-play setup and secure network isolation, acting as a fully localized AI infrastructure.
*   **Plug-and-Play Setup:** Ready for production deployment with minimal networking configuration.
*   **Backup-First Onboarding:** During initial setup, Omni mandates the configuration of automatic backups to a local NAS, SMB share, or P2P-synced device to ensure data sovereignty.
*   **Hardware Resilience:** Integrated battery backup (mini-UPS) is recommended (or included) to provide graceful shutdown during power failures, preventing database corruption.
*   **Direct-to-Chat Data Ingestion:** Users populate their personal knowledge base by simply forwarding files directly to their secure messenger chat with Omni.
*   **Service-Linked Warranty:** Hardware replacement guarantees (e.g., 48-hour swap) are provided exclusively for units covered by a **2-year Extended Service Subscription**.

### Hardware Tiers
Different hardware tiers unlock different capabilities. The software is identical across all tiers — a single Go binary with embedded web UI.

| Tier | Hardware Class | CPU | RAM | LLM Capability |
|------|---------------|-----|-----|-----------------|
| Entry | ARM SoC (e.g. RK3328) | 4× Cortex-A53 | 4 GB | Cloud API only |
| Mid | ARM SoC (e.g. RK3566/RK3399) | 4× A55 / 2× A72+4× A55 | 4–8 GB | Small local models (1.5–3B) or cloud API |
| Pro | x86 mini-PC (e.g. Intel N100) | 4C/4T | 8–16 GB | Mid-size local models (3–7B) |
| Server | x86 with dedicated GPU | Multi-core + GPU | 32+ GB | Large local models (13–70B), full air-gap |

The Entry tier (MVP target) uses cloud LLM APIs for inference while keeping all documents and metadata stored locally on the device. Higher tiers progressively add local inference capabilities up to fully air-gapped operation.

### Deployment Model (Entry/Mid Tiers)
*   **Single Binary:** The entire application — API server, chat UI, admin panel — is compiled into one Go binary (~15–20 MB) with embedded static assets. Deployment is copying one file.
*   **Native systemd Service:** No Docker or container runtime. The binary runs as a systemd unit directly on the host OS, minimizing RAM overhead.
*   **Storage Separation:** The OS and application binary reside on the SD card (read-mostly). All user data — documents, database, vector index, backups — lives on a SATA HDD, protecting the SD card from write wear.

### Software-Only Version
A Docker-based deployment designed for on-premise installation on the client's existing enterprise servers.
*   **Containerized Isolation:** Deployed as a set of sandboxed containers (Docker/Podman), ensuring Omni remains a secure "neighbor" that cannot access other files, databases, or processes on the host server.
*   **Encrypted Local Storage:** All internal indices and chat logs are stored in encrypted volumes, isolated from the host's primary file system.
*   **Resource Capping:** Admins can strictly limit CPU/GPU and RAM usage to prevent Omni from impacting the performance of other mission-critical services running on the same hardware.

## 5. User Experience & Multi-Tenant Architecture
Omni provides a bifurcated interface strategy to balance deep administrative control with frictionless daily usage.

### Administrative Control (Web-Based)
*   **Centralized Admin Panel:** Accessible via a secure local browser connection (e.g., `http://omni-box.local/admin`). Administrators manage users, define Role-Based Access Control (RBAC), connect data sources, and monitor system health and logs.
*   **Granular Permissioning:** Admins assign specific knowledge silos to individual users or teams, ensuring strict data isolation within the organization.

### Multi-Channel User Interaction
*   **Local Web Chat:** A full-featured, browser-based chat interface (similar to ChatGPT) that runs entirely on the local appliance.
*   **Secure Messenger Integration:** Users interact with Omni through their preferred secure channels: **Telegram**, **Matrix**, **Jami**, or **SimpleX Chat**.
*   **Voice-to-Action Messaging:** Full support for voice messages across all messenger platforms. Omni locally transcribes audio into text, identifies intent, and executes actions or provides synthesized voice responses without any data leaving the local network.
*   **Conversational Skill Acquisition:** Users can "teach" the assistant new workflows directly through natural language. By saying *"Always flag invoices over $1000 for manual review,"* the user creates a persistent automated rule that is instantly codified into the system's logic.
*   **Proactive Assistance:** The system pushes context-aware alerts and reminders directly to the user's chosen messenger, reducing the need for manual check-ins.

### Conversational Onboarding
Instead of presenting a dashboard with empty states, Omni greets new users with an AI-driven onboarding conversation. The chat is the product from minute one — even before any documents are uploaded.

**Onboarding Flow:**
1.  **Introduction:** The AI greets the user, asks their name and what their company does.
2.  **Context Gathering:** The AI asks about the types of documents they work with, what questions they typically need answered, and what problems they want to solve.
3.  **Data Ingestion Prompt:** The AI suggests uploading the first documents directly into the chat (drag & drop or file attach).
4.  **Processing Confirmation:** The AI processes uploaded files in the background and reports back: *"I've processed 3 PDFs (47 pages). Try asking me something about them!"*
5.  **RAG Activation:** From this point on, the AI answers questions using the uploaded knowledge base with source citations.

**Example Onboarding Dialogue:**
```
Omni:  Hi! I'm Omni, your knowledge assistant. What's your name
       and what does your company do?

User:  I'm Anna, we run a small legal consultancy.

Omni:  Nice to meet you, Anna! What types of documents do you
       work with most often? Contracts, court filings, internal
       memos?

User:  Mostly contracts and NDAs. Sometimes regulatory documents.

Omni:  Got it — contracts, NDAs, and regulatory docs. What kind
       of questions do you usually need to answer about them?
       For example: checking specific clauses, comparing terms
       across documents, summarizing key points?

User:  All of the above, but mainly finding specific clauses
       quickly and comparing terms between different contracts.

Omni:  Perfect. You can start by sending me a few documents
       right here in the chat — just drag and drop or attach
       files. I'll study them and then you can try asking me
       questions.

User:  [uploads 3 PDF files]

Omni:  Done! I've processed 3 documents (52 pages total):
       · Service Agreement 2024.pdf — 24 pages
       · NDA Template.pdf — 8 pages
       · Regulatory Compliance Guide.pdf — 20 pages
       Try asking me something about them!

User:  What are the termination clauses in the service agreement?

Omni:  The Service Agreement contains two termination clauses...
       [answer with citations from the document]
```

**Organization Profile:** During onboarding the AI extracts structured data from the conversation (business type, document types, language preferences) and stores it as an organization profile. This profile enriches the system prompt for all subsequent interactions, making the assistant contextually aware of the business from day one.

## 6. Target Use Cases
* **Legal & Compliance:** Indexes confidential case files locally. The assistant drafts templates with zero risk of NDA breaches.
* **Manufacturing & Engineering:** A technician sends a voice message; the assistant replies with a video timestamp showing the fix.
* **Private Healthcare:** Manages patient records and audio consultations while strictly adhering to HIPAA/GDPR.
* **Enterprise HR:** Automates repetitive queries, guiding new hires through onboarding and tracking missing documentation.

## 7. Integrations & Extensibility
The assistant acts as the intelligent core, connecting to tools via an **Abstract Messenger Gateway**.

* **Flexible Communication Layer:** Native support for **Telegram** and **Matrix**, but with the option to switch to **Direct-to-User P2P communication** via **Jami** or **SimpleX Chat**, bypassing centralized messenger servers for high-security environments.
* **Hybrid Skill Architecture (Deterministic Action Layer):** Omni bridges the gap between AI flexibility and enterprise-grade reliability. 
    - **Skill Creation Schema:** When a user "teaches" a skill via chat, the LLM acts as an architect, translating natural language into a structured entry in a local relational database. Each skill record follows a strict schema:
        - **Trigger:** Event source (e.g., *File Upload*, *Keyword Detection*, *Schedule*).
        - **Condition (JSON-Logic):** Predicates to evaluate (e.g., *Amount > 1000*, *File Extension == .pdf*).
        - **Action:** Deterministic task to execute (e.g., *Update CRM*, *Send Alert*, *Generate Report*).
    - **Safe Execution Runtime:** High-level logic is stored as **JSON-Logic** for standard rules or **Sandboxed Python** for complex computations. This ensures the system remains stable and auditable even if the underlying AI model is updated or swapped.
    - **Guaranteed Predictability:** Once codified into the database, skills execute with 100% precision, eliminating "hallucinations" during critical business actions.
* **Model Context Protocol (MCP):** Secure, standardized API connections between the AI engine and external data sources (CRMs, ERPs, Task Trackers).

## 8. Technical Architecture & Security
At its core, Omni utilizes a modular RAG architecture with flexible LLM routing.

### Fully Local Mode (Air-Gapped)
All components—vector database, knowledge storage, embedding generation, and LLM inference—run entirely on-premise.

### Secure Hybrid Mode (Cloud LLM with Controlled Exposure)
The local core applies token-level sensitive data masking before transmission to cloud LLMs. The local system securely rehydrates the real data before displaying the response.

### Zero-Knowledge Hybrid Infrastructure
Omni implements a Zero-Knowledge architectural pattern. Mapping keys for de-anonymizing data never leave the client's local network, ensuring intercepted cloud data remains mathematically indecipherable.

### Sovereign Interaction Layer
Omni's communication protocol is designed to eliminate "meta-data leakage." In its highest security configuration, the system utilizes peer-to-peer (P2P) messaging directly with the user's client, ensuring that neither the content nor the metadata of the conversation passes through a third-party server.

### Minimum-Access-Only Architecture
Omni is architected to operate with zero persistent access to the host Operating System. All core components—LLM engine, vector database, and logic modules—are executed in a **Sandboxed Runtime** (Docker or WebAssembly), meaning even a compromise of the AI layer cannot translate into a compromise of the hardware or adjacent software services. All system updates are delivered as cryptographically signed, read-only images to maintain a trusted state.

### Shared Responsibility Model (Reliability & Legal)
To ensure long-term data integrity and clear legal boundaries, Omni operates on a Shared Responsibility framework:
*   **Omni’s Responsibility:** We provide the secure technical infrastructure, local encryption protocols, and a non-leaking data pipeline. Due to its Air-Gapped/Zero-Knowledge design, technical data leaks from the Omni core are mathematically minimized.
*   **Client’s Responsibility:** The client is responsible for the physical security of the device and the consistent execution of backups. Legal liability for data loss due to hardware failure without active backups rests with the client, as stipulated in the service agreement.

### MVP Implementation Architecture
The MVP targets the Entry-tier hardware (ARM SoC, 4 GB RAM, SD card + SATA HDD) with cloud LLM APIs for inference. The architecture follows hexagonal (ports & adapters) principles — domain logic is decoupled from infrastructure, allowing adapters to be swapped (e.g. cloud LLM → local LLM) without changing business logic.

**Tech Stack (MVP):**

| Component | Choice | Rationale |
|-----------|--------|-----------|
| Language | Go | Single binary, minimal RAM (~10–30 MB), trivial cross-compilation to ARM64 |
| LLM inference | Cloud API (OpenAI / Anthropic) | No local model on Entry tier; adapter interface allows adding local models later |
| Embeddings | Cloud API (OpenAI text-embedding-3-small) | Same adapter pattern; local embedding models can be added for higher tiers |
| Vector search | In-memory brute-force (`[]float32`) | For <50K vectors: ~73 MB RAM, ~10–50 ms search on A53. Zero dependencies. Exact results |
| Metadata DB | SQLite (pure Go, no CGO) | Single file on SATA HDD, zero daemon overhead, trivial backups |
| Frontend | Static HTML/CSS/JS embedded in binary | No framework, no build pipeline. Served by the same Go process |
| Deployment | systemd unit | No Docker on Entry/Mid tiers. One binary, one service file |

**Storage Layout:**
```
SD card (read-mostly):
  /opt/omni/omni              ← single Go binary (~15–20 MB)
  /opt/omni/config.env        ← API keys, settings

SATA HDD (read/write):
  /data/omni/omni.db          ← SQLite (users, orgs, documents, chunks, conversations)
  /data/omni/vectors.bin      ← persisted vector index (loaded into RAM at startup)
  /data/omni/documents/       ← original uploaded files
  /data/omni/backups/         ← automated backups

RAM:
  vectors [][]float32         ← hot vector index for semantic search
  SQLite page cache           ← automatic
```

**RAM Budget (Entry Tier — 4 GB):**
```
OS + systemd:              ~300 MB
Go process (app logic):    ~30–50 MB
In-memory vector index:    ~30–70 MB  (5K–50K chunks)
SQLite page cache:         ~10–20 MB
────────────────────────────────────
Total:                     ~400–450 MB
Free:                      ~3.5 GB
```

**Domain Ports:**
```go
type LlmService interface {
    Complete(ctx context.Context, prompt string, contextChunks []string) iter.Seq2[string, error]
}

type EmbeddingService interface {
    Embed(ctx context.Context, texts []string) ([][]float32, error)
}

type VectorStore interface {
    Add(id string, orgID string, embedding []float32)
    Search(query []float32, orgID string, topK int) []VectorResult
    Remove(ids []string)
    Save(path string) error
    Load(path string) error
}

type DocumentParser interface {
    Parse(content []byte, mimeType string) (string, error)
}
```

**Data Model:**
```sql
CREATE TABLE organizations (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE organization_profiles (
    org_id TEXT PRIMARY KEY REFERENCES organizations(id),
    contact_name TEXT,
    business_description TEXT,
    document_types TEXT,
    preferences TEXT,
    onboarding_completed INTEGER DEFAULT 0,
    raw_answers TEXT,
    updated_at TEXT
);

CREATE TABLE users (
    id TEXT PRIMARY KEY,
    org_id TEXT NOT NULL REFERENCES organizations(id),
    name TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',
    auth_token_hash TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE documents (
    id TEXT PRIMARY KEY,
    org_id TEXT NOT NULL REFERENCES organizations(id),
    filename TEXT NOT NULL,
    mime_type TEXT,
    size_bytes INTEGER,
    status TEXT NOT NULL DEFAULT 'pending',
    error_message TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE chunks (
    id TEXT PRIMARY KEY,
    document_id TEXT NOT NULL REFERENCES documents(id),
    org_id TEXT NOT NULL,
    content TEXT NOT NULL,
    position INTEGER NOT NULL,
    token_count INTEGER
);

CREATE TABLE conversations (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    title TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    conversation_id TEXT NOT NULL REFERENCES conversations(id),
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    sources_json TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
```

**API Endpoints:**
```
POST   /api/documents            Upload file (multipart)
GET    /api/documents            List documents for organization
DELETE /api/documents/:id        Delete document and its chunks

POST   /api/chat                 Send message → SSE streaming response
GET    /api/conversations        List user conversations
GET    /api/conversations/:id    Get conversation messages

POST   /api/auth/login           Login → token
GET    /api/health               System status
```

**Project Structure:**
```
omni/
├── cmd/
│   └── omni/
│       └── main.go                 # Entry point, DI wiring
├── internal/
│   ├── domain/
│   │   ├── ports.go                # Interfaces: LLM, Embedding, VectorStore, Parser
│   │   ├── types.go                # Organization, User, Document, Chunk, Message
│   │   ├── ingest.go               # Use case: file → parse → chunk → embed → store
│   │   ├── query.go                # Use case: question → search → LLM → answer
│   │   └── chunker.go              # Text splitting
│   ├── adapters/
│   │   ├── llm/
│   │   │   └── openai.go           # OpenAI ChatCompletion adapter
│   │   ├── embeddings/
│   │   │   └── openai.go           # OpenAI Embeddings adapter
│   │   ├── parsers/
│   │   │   ├── pdf.go
│   │   │   ├── docx.go
│   │   │   └── plaintext.go
│   │   └── storage/
│   │       ├── sqlite.go           # SQLite adapter (metadata, chunks)
│   │       └── memvec.go           # In-memory vector index + persistence
│   ├── channels/
│   │   ├── http.go                 # HTTP API + SSE streaming
│   │   └── telegram.go            # Telegram bot adapter
│   └── config/
│       └── config.go               # Env-based configuration
├── web/                            # Static frontend (embedded into binary via go:embed)
│   ├── index.html
│   ├── app.js
│   └── style.css
├── migrations/
│   └── 001_initial.sql
├── Makefile                        # build, cross-compile, test targets
├── go.mod
└── go.sum
```

## 9. Future Extensions (Roadmap)
* **Native Mobile Application:** Dedicated iOS/Android apps with built-in P2P communication and biometric authentication.
* **Advanced Agentic Workflows:** Enabling the system to chain multiple tasks independently across different enterprise tools.
* **Voice-Activated Hardware:** Integration with secure, local voice-command hardware for hands-free operation.
* **Federated Knowledge Sharing:** Encrypted, cross-organization data syncing for secure collaboration.

## 10. Service & Support Policies
* **Standard Support:** Includes documentation and regular software updates delivered via secure channels.
* **Extended Service Subscription (2-Year Plan):**
    - **Hardware Swap Guarantee:** 48-hour replacement of the physical appliance in case of failure.
    - **Proactive Health Monitoring:** Secure, opt-in remote monitoring to identify system issues before they impact performance.
    - **Dedicated Integration Assistance:** Priority access to our team for building custom connectors and workflows.

## 11. MVP Development Plan

Chat-first approach: the user interacts with a working AI assistant from minute one. Documents and RAG are layered on top of the conversation, not the other way around.

### Phase 1 — Foundation + Chat + Onboarding
*   Go project setup, domain types, port interfaces
*   SQLite adapter (schema, migrations, basic CRUD)
*   Cloud LLM adapter (OpenAI / Anthropic)
*   HTTP server + SSE streaming
*   Static chat UI (HTML/CSS/JS embedded in binary)
*   AI-driven onboarding conversation (collects organization profile)
*   **User gets:** opens chat → AI greets, asks questions, builds context

### Phase 2 — Document Ingestion
*   File upload via chat (drag & drop / attach)
*   Document parsers: plaintext → PDF → DOCX
*   Text chunker (recursive splitting, ~500 tokens/chunk)
*   Cloud embeddings adapter (OpenAI text-embedding-3-small)
*   In-memory vector index with persistence to SATA HDD
*   Background processing with status updates in chat
*   **User gets:** drops files into chat → AI confirms processing

### Phase 3 — RAG Query
*   Query embedding via cloud API
*   In-memory vector similarity search (brute-force cosine, filtered by org)
*   Prompt assembly (system prompt + org profile + context chunks + user question)
*   Streaming LLM response with source citations
*   **User gets:** asks questions → gets answers with references to source documents

### Phase 4 — Telegram Bot
*   Telegram bot adapter (long polling)
*   Same onboarding + RAG flow via Telegram
*   File forwarding → document ingestion
*   Voice messages → cloud STT → RAG query (optional)
*   **User gets:** full Omni experience through Telegram

### Phase 5 — Auth & Multi-Tenancy
*   User registration and login (token-based)
*   Organization-level data isolation
*   Role-based access (admin / user)
*   **User gets:** multiple users per organization, isolated knowledge bases

### Phase 6 — Deployment & Packaging
*   systemd unit files
*   Install script for ARM64 Debian (Armbian / DietPi)
*   Automated backups to SATA HDD
*   Health monitoring endpoint
*   Cross-compilation: `GOOS=linux GOARCH=arm64 go build`
*   **User gets:** plug in the box, run one command, start chatting
