# Gemini-CLI Document Revision Prompt
## Sovereign Fleet — Architecture Document Suite Revision

---

You are the principal architect of the Sovereign Fleet documentation suite. You have been given four input documents and a set of precise revision instructions. Your task is to revise three existing documents and produce three new documents. Work through each document in the order listed. Do not change the voice, philosophy, or intent of the originals — your role is structural editing, generalization, and extraction, not rewriting.

---

## INPUT DOCUMENTS

You have access to the following documents in your working directory:

1. `SOVEREIGN_WORKSPACE_MANIFESTO.md`
2. `POLY_WORKSPACE_MANIFESTO.md`
3. `MULTI_FLIGHT_GOVERNANCE.md`
4. `FLEET_WORKSPACE_REVIEW_20260303.md`

---

## SHARED CONTEXT: KEY CONVENTIONS

Before editing, internalize these conventions. They apply to all documents:

- **jeBNF** stands for "Joe's Extended Backus-Naur Form." It is eBNF augmented with Wirth Syntax Notation (WSN) clarifications: ordered option uses `/` (not `|`), and explicit terminal designation uses `;`. When any document references `.jebnf` files or the jeBNF format, do not alter or explain the term — treat it as established convention. You will create a dedicated primer for it separately.

- **Manifesto documents are normative, not descriptive.** They describe intended architecture, not current implementation state. The Fleet Review is the document that tracks delta between intent and reality. Do not soften or hedge manifesto language to account for implementation gaps.

- **AntigravitySpace** is a specific instantiation of the general patterns the manifestos define. Named workspaces (`George`, `Olympus2`, `OlympusMCP`, `GND-Registry`, etc.), zone labels (`00SDLC`, `10GNDT`, `20POSI`), and tool names (`Antigravity`, `Gemini-CLI`) belong in implementation documents, not general manifesto documents — unless used briefly as a labelled example.

---

## DOCUMENT 1 — REVISE: `SOVEREIGN_WORKSPACE_MANIFESTO.md`

**Changes required:**

1. **Add a normative preamble** immediately after the title/version line and before Section 1. It should be a short paragraph (3–5 sentences) that:
   - States this document is normative: it describes intended architecture, not current implementation state.
   - Identifies its role in the document hierarchy: it sits below the Poly-Workspace Manifesto and above workspace-specific topology documents.
   - States that implementation status is tracked in Fleet Review documents and Conductor cycle records.

2. **No other changes.** This document is well-structured and appropriately general. Do not alter any other content.

---

## DOCUMENT 2 — REVISE: `POLY_WORKSPACE_MANIFESTO.md`

**Changes required:**

1. **Add the same normative preamble** as Document 1 (adapted to this document's position in the hierarchy: it sits above the Sovereign Workspace Manifesto and below the Federation root governance layer).

2. **Generalize the Tier II table** in Section 2. Replace the specific zone examples (`00SDLC`, `10GNDT`) with generic placeholders (e.g., `nnXXXX`). The table structure and descriptions remain; only the specific names are replaced. Add a footnote or inline note reading: *"See the Federation Topology document for the AntigravitySpace instantiation of this tier structure."*

3. **Audit all remaining sections** for any reference to specific workspace names, zone labels, tool names, or cycle numbers. For each:
   - If the reference is illustrative (i.e., one example among many possible), retain it but wrap it in a clearly labelled `> **Example (AntigravitySpace):**` blockquote.
   - If the reference is structural (i.e., the section depends on it), extract the specific content and replace it with a generalized description. Note in a comment: `<!-- Specific implementation: see ANTIGRAVITY_FEDERATION_TOPOLOGY.md -->`.

4. **Section 4 — Governance: The Conductor Protocol**: Replace the current content with a generalized description of the Conductor as an abstract Shared Intent Layer. It must specify the four required *properties* of any compliant Conductor implementation:
   - Maintains a canonical record of active Tracks and their current state.
   - Is the authoritative source for resolving intent conflicts between Flights.
   - Publishes a machine-readable state that any compliant AI tool can consume.
   - Accepts "Landing" events that update the global state at cycle completion.
   
   Do not name specific tools, directory structures, or file formats in this section. End the section with: *"See `CONDUCTOR_IMPLEMENTATION_ANTIGRAVITY.md` for the reference implementation using Antigravity and Gemini-CLI."*

---

## DOCUMENT 3 — REVISE: `MULTI_FLIGHT_GOVERNANCE.md`

**Changes required:**

1. **Add the same normative preamble** (adapted: this document sits at the Federation governance layer, above the Poly-Workspace Manifesto).

2. **Section 2A — Replace the hardware-specific identity model** with a generalized **Composite Platform Identity (CPI)** model. The revision must:

   - Retain the *requirements* intact: every Flight must have a verifiable, non-repudiable platform identity that is bound to a specific execution context, time-windowed, and cryptographically signed.
   - Replace all TPM-specific language with CPI. Define CPI as: *"A Composite Platform Identity (CPI) is a cryptographically signed attestation token that binds a Flight to its execution context. The token must satisfy the host Sovereign Policy (jeBNF) at the time of execution. The mechanism by which the CPI is produced is determined by the deployment profile."*
   - Add a **CPI Binding Profiles** table immediately after the CPI definition. The table must have three columns: `Deployment Context`, `CPI Mechanism`, `Notes`. Populate it with these four rows:

     | Deployment Context | CPI Mechanism | Notes |
     |---|---|---|
     | Bare metal / owned hardware | TPM attestation + hardware fingerprint | Highest assurance; preferred for sovereign nodes |
     | Cloud VM / ephemeral compute | Workload identity (e.g., SPIFFE, cloud IAM) | Token scoped to VM lifecycle |
     | Container / CI environment | Short-lived signed attestation token | Issued at job start; expires at job end |
     | Developer workstation | Hardware key (YubiKey, Passkey) + OS identity | Human-in-Command binding |

   - Update the **"Refusal to Run" Mandate** to read: a Sovereign Seal artifact is prohibited from executing on any platform that cannot produce a valid CPI token satisfying the current Sovereign Policy (jeBNF), is present on the Global Prohibition Registry, or violates the active Temporal Window. Remove the TPM-specific attestation check as a named requirement (it is now subsumed by the CPI Binding Profile for bare metal).

3. **Section 2C — Sovereign Mesh**: This section is currently labelled "C" but follows section "B" in the document while another "C" appears later. Audit and correct the section lettering to be sequential (A, B, C, D, E) throughout Section 2.

4. **No other content changes.** The Resonance Bus, Pulse Packet, Semantic Delta-Sync, ACoC, and LanceDB sections are appropriately general and should not be altered.

---

## DOCUMENT 4 — CREATE: `ANTIGRAVITY_FEDERATION_TOPOLOGY.md`

**This is a new document. Create it in full.**

**Purpose:** This document is the AntigravitySpace-specific instantiation of the Poly-Workspace Manifesto's Federation model. Content extracted from `POLY_WORKSPACE_MANIFESTO.md` during revision belongs here.

**Structure:**

```
# AntigravitySpace Federation Topology
**Instantiation of:** Poly-Workspace Manifesto v3.1
**Last Reviewed:** [date from FLEET_WORKSPACE_REVIEW_20260303.md]

## 1. Zone Structure
[The specific zone table from the Poly Manifesto: 00SDLC, 10GNDT, 20POSI, 30INFR, 40RNDL, 50RNDF, 99PUBL — sourced from the Fleet Review overview table]

## 2. Active Workspaces
[Summary of named workspaces per zone, sourced from the Fleet Review]

## 3. Conductor Implementation
[One paragraph: states that the AntigravitySpace Conductor is implemented per CONDUCTOR_IMPLEMENTATION_ANTIGRAVITY.md, and references the CYC-NNN cycle numbering convention]

## 4. Fleet Health Snapshot
[The Fleet Health and Highest-Priority Issues tables from the Fleet Review, Section 4]
```

Do not add content beyond what can be sourced directly from the four input documents.

---

## DOCUMENT 5 — CREATE: `CONDUCTOR_IMPLEMENTATION_ANTIGRAVITY.md`

**This is a new document. Create it in full.**

**Purpose:** Documents the concrete Conductor implementation for the AntigravitySpace fleet using Antigravity and Gemini-CLI. This is the reference implementation of the abstract Conductor protocol defined in the Poly-Workspace Manifesto.

**Structure:**

```
# Conductor Implementation — AntigravitySpace
**Implements:** Conductor Protocol (Poly-Workspace Manifesto v3.1, Section 4)
**Primary Tools:** Antigravity, Gemini-CLI

## 1. Directory Structure
[Document the /conductor/ directory as the root. Describe its role as the Supreme Law of the Federation for this instantiation.]

## 2. Cycle Track Schema (jeBNF)
[Document the CYC-NNN numbering convention. Describe what a cycle track record contains: Track-ID, current state, active goal, landing status. Note that the schema is expressed in jeBNF — see jeBNF_PRIMER.md for format reference.]

## 3. Coord-State
[Document the .coord-state.jebnf file: its role as the canonical workspace state hash, the meaning of DIRTY entries, and the GIT_HASH / coord-hash relationship as evidenced in the Fleet Review.]

## 4. Tool Integration
### 4.1 Antigravity
[Document the @BIN and @RUN protocol references. Note the fleet tools (fleet-index, fleet-substrate, fleet-mission-runner) as the primary consumers of Conductor state.]

### 4.2 Gemini-CLI
[Document the Shadow-Sync Handshake: the "Request for Summary" flow where Gemini-CLI requests semantic context from the conductor/ before pulling code. Document the Semantic Delta-Sync performed on network re-entry.]

## 5. Landing Protocol
[Document what constitutes a completed "Landing": committed/pushed results, updated Global Registry, Chronicle produced. Reference the Multi-Flight Governance doc for the distributed variant.]
```

Populate each section using only content that can be sourced or reasonably inferred from the four input documents. Where a section topic is referenced in the inputs but not fully specified, write a one-sentence placeholder: *"[To be specified in CYC-NNN — see Fleet Review recommendations.]"*

---

## DOCUMENT 6 — CREATE: `jeBNF_PRIMER.md`

**This is a new document. Create it in full.**

**Purpose:** A concise reference for any contributor or AI tool encountering `.jebnf` files in the fleet. One page maximum.

**Structure:**

```
# jeBNF Primer
**Format:** Joe's Extended Backus-Naur Form
**Lineage:** eBNF + Wirth Syntax Notation (WSN) clarifications

## 1. Lineage and Purpose
[2–3 sentences: jeBNF is eBNF extended with two specific WSN clarifications to improve precision in machine-readable sovereign policy and coordination state files.]

## 2. Key Divergences from Standard eBNF

| Construct | Standard eBNF | jeBNF |
|---|---|---|
| Alternation (unordered) | `\|` | `\|` (retained) |
| Alternation (ordered) | not specified | `/` — first matching alternative is selected |
| Terminal designation | convention-dependent | `;` — explicit terminal marker |

## 3. Worked Example
[Produce a short, plausible jeBNF fragment representing a Pulse Packet (from Multi-Flight Governance Section 2C): Flight-ID, Track-ID, Intent-Hash, Current-Goal. Show the grammar definition and one concrete instance. Use the ordered `/` and terminal `;` conventions.]

## 4. File Conventions in the Fleet
[List the known .jebnf file roles: .coord-state.jebnf (workspace coordination hash), FLUX_LEDGER.jebnf (immutable flight recorder), Sovereign Policy files. One line each.]
```

---

## EXECUTION ORDER

Process documents in this sequence to avoid forward-reference issues:

1. `jeBNF_PRIMER.md` (new — no dependencies)
2. `SOVEREIGN_WORKSPACE_MANIFESTO.md` (revise — lightest changes)
3. `POLY_WORKSPACE_MANIFESTO.md` (revise — extractions feed Document 4)
4. `ANTIGRAVITY_FEDERATION_TOPOLOGY.md` (new — depends on Poly revision)
5. `MULTI_FLIGHT_GOVERNANCE.md` (revise — CPI model)
6. `CONDUCTOR_IMPLEMENTATION_ANTIGRAVITY.md` (new — depends on all revisions)

After completing all six documents, output a one-paragraph **Revision Summary** listing: files modified, files created, any content that could not be sourced from the inputs and was left as a placeholder, and any ambiguities encountered that require human clarification.

---

*End of prompt.*
