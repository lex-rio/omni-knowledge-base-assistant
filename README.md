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
