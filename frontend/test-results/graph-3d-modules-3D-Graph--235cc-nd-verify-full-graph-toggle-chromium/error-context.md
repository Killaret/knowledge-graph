# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: graph-3d-modules.spec.ts >> 3D Graph - Modular Architecture >> should open graph page and verify full graph toggle
- Location: tests\graph-3d-modules.spec.ts:236:3

# Error details

```
Error: Failed to create link: 400 - {"error":"Key: 'createLinkRequest.SourceNoteID' Error:Field validation for 'SourceNoteID' failed on the 'required' tag\nKey: 'createLinkRequest.TargetNoteID' Error:Field validation for 'TargetNoteID' failed on the 'required' tag"}
```

# Page snapshot

```yaml
- generic [ref=e3]:
  - generic [ref=e4]:
    - generic [ref=e5]:
      - button "2D" [ref=e6] [cursor=pointer]:
        - img [ref=e7]
        - generic [ref=e17]: 2D
      - button "3D" [ref=e18] [cursor=pointer]:
        - img [ref=e19]
        - generic [ref=e23]: 3D
      - button "List" [ref=e24] [cursor=pointer]:
        - img [ref=e25]
        - generic [ref=e26]: List
    - generic [ref=e27]:
      - button "🌌 All" [ref=e28] [cursor=pointer]:
        - generic [ref=e29]: 🌌
        - generic [ref=e30]: All
      - button "⭐ Stars" [ref=e31] [cursor=pointer]:
        - generic [ref=e32]: ⭐
        - generic [ref=e33]: Stars
      - button "🪐 Planets" [ref=e34] [cursor=pointer]:
        - generic [ref=e35]: 🪐
        - generic [ref=e36]: Planets
      - button "☄️ Comets" [ref=e37] [cursor=pointer]:
        - generic [ref=e38]: ☄️
        - generic [ref=e39]: Comets
      - button "🌀 Galaxies" [ref=e40] [cursor=pointer]:
        - generic [ref=e41]: 🌀
        - generic [ref=e42]: Galaxies
    - generic [ref=e43]:
      - textbox "Search notes..." [ref=e44]
      - button "Search" [ref=e45] [cursor=pointer]:
        - img [ref=e46]
    - button "Menu" [ref=e50] [cursor=pointer]:
      - img [ref=e51]
    - button "Create new note" [ref=e52] [cursor=pointer]:
      - img [ref=e53]
  - generic [ref=e56]:
    - generic [ref=e57]:
      - strong [ref=e58]: "100"
      - text: nodes
    - generic [ref=e59]:
      - strong [ref=e60]: "2448"
      - text: links
```

# Test source

```ts
  1   | import type { APIRequestContext } from '@playwright/test';
  2   | 
  3   | export interface NoteData {
  4   |   id?: string;
  5   |   title: string;
  6   |   content?: string;
  7   |   type?: string;
  8   |   metadata?: Record<string, unknown>;
  9   | }
  10  | 
  11  | export interface LinkData {
  12  |   id?: string;
  13  |   sourceNoteId: string;
  14  |   targetNoteId: string;
  15  |   weight?: number;
  16  |   link_type?: string;
  17  |   metadata?: Record<string, unknown>;
  18  | }
  19  | 
  20  | /**
  21  |  * Get backend base URL from environment or default
  22  |  */
  23  | export function getBackendUrl(): string {
  24  |   return process.env.BACKEND_URL || 'http://localhost:8080';
  25  | }
  26  | 
  27  | /**
  28  |  * Create a note via API
  29  |  */
  30  | export async function createNote(
  31  |   request: APIRequestContext,
  32  |   data: Partial<NoteData>
  33  | ): Promise<{ id: string; [key: string]: unknown }> {
  34  |   const payload = {
  35  |     title: data.title || 'Test Note',
  36  |     content: data.content || 'Test content',
  37  |     type: data.type,
  38  |     metadata: data.metadata || {},
  39  |   };
  40  | 
  41  |   const response = await request.post(`${getBackendUrl()}/notes`, {
  42  |     data: payload,
  43  |   });
  44  | 
  45  |   if (!response.ok()) {
  46  |     const errorText = await response.text();
  47  |     throw new Error(`Failed to create note: ${response.status()} - ${errorText}`);
  48  |   }
  49  | 
  50  |   return await response.json();
  51  | }
  52  | 
  53  | /**
  54  |  * Create multiple notes in batch
  55  |  */
  56  | export async function createNotes(
  57  |   request: APIRequestContext,
  58  |   notes: Partial<NoteData>[]
  59  | ): Promise<Array<{ id: string; [key: string]: unknown }>> {
  60  |   const created = [];
  61  |   for (const noteData of notes) {
  62  |     const note = await createNote(request, noteData);
  63  |     created.push(note);
  64  |   }
  65  |   return created;
  66  | }
  67  | 
  68  | /**
  69  |  * Create a link between two notes
  70  |  */
  71  | export async function createLink(
  72  |   request: APIRequestContext,
  73  |   sourceId: string,
  74  |   targetId: string,
  75  |   weight = 0.5,
  76  |   linkType = 'related'
  77  | ): Promise<{ id: string; [key: string]: unknown }> {
  78  |   const payload: LinkData = {
  79  |     sourceNoteId: sourceId,
  80  |     targetNoteId: targetId,
  81  |     weight,
  82  |     link_type: linkType,
  83  |     metadata: {},
  84  |   };
  85  | 
  86  |   const response = await request.post(`${getBackendUrl()}/links`, {
  87  |     data: payload,
  88  |   });
  89  | 
  90  |   if (!response.ok()) {
  91  |     const errorText = await response.text();
> 92  |     throw new Error(`Failed to create link: ${response.status()} - ${errorText}`);
      |           ^ Error: Failed to create link: 400 - {"error":"Key: 'createLinkRequest.SourceNoteID' Error:Field validation for 'SourceNoteID' failed on the 'required' tag\nKey: 'createLinkRequest.TargetNoteID' Error:Field validation for 'TargetNoteID' failed on the 'required' tag"}
  93  |   }
  94  | 
  95  |   return await response.json();
  96  | }
  97  | 
  98  | /**
  99  |  * Create a star topology - center note with surrounding notes
  100 |  */
  101 | export async function createStarTopology(
  102 |   request: APIRequestContext,
  103 |   centerNote: Partial<NoteData>,
  104 |   surroundingNotes: Partial<NoteData>[],
  105 |   linkWeight = 0.7
  106 | ): Promise<{
  107 |   center: { id: string };
  108 |   surrounding: Array<{ id: string }>;
  109 | }> {
  110 |   const center = await createNote(request, centerNote);
  111 |   const surrounding = [];
  112 | 
  113 |   for (const noteData of surroundingNotes) {
  114 |     const note = await createNote(request, noteData);
  115 |     surrounding.push(note);
  116 |     await createLink(request, center.id, note.id, linkWeight);
  117 |   }
  118 | 
  119 |   return { center, surrounding };
  120 | }
  121 | 
  122 | /**
  123 |  * Create a chain topology - notes linked in sequence
  124 |  */
  125 | export async function createChainTopology(
  126 |   request: APIRequestContext,
  127 |   notes: Partial<NoteData>[],
  128 |   linkWeight = 0.8
  129 | ): Promise<Array<{ id: string }>> {
  130 |   const created = [];
  131 | 
  132 |   for (let i = 0; i < notes.length; i++) {
  133 |     const note = await createNote(request, notes[i]);
  134 |     created.push(note);
  135 | 
  136 |     // Link to previous note
  137 |     if (i > 0) {
  138 |       await createLink(request, created[i - 1].id, note.id, linkWeight);
  139 |     }
  140 |   }
  141 | 
  142 |   return created;
  143 | }
  144 | 
  145 | /**
  146 |  * Delete a note via API
  147 |  */
  148 | export async function deleteNote(request: APIRequestContext, noteId: string): Promise<void> {
  149 |   const response = await request.delete(`${getBackendUrl()}/notes/${noteId}`);
  150 | 
  151 |   if (!response.ok() && response.status() !== 404) {
  152 |     const errorText = await response.text();
  153 |     throw new Error(`Failed to delete note: ${response.status()} - ${errorText}`);
  154 |   }
  155 | }
  156 | 
  157 | /**
  158 |  * Clean up all test data
  159 |  */
  160 | export async function cleanupTestData(
  161 |   request: APIRequestContext,
  162 |   noteIds: string[]
  163 | ): Promise<void> {
  164 |   for (const id of noteIds) {
  165 |     try {
  166 |       await deleteNote(request, id);
  167 |     } catch (e) {
  168 |       // Ignore errors during cleanup
  169 |       console.warn(`Failed to delete note ${id}:`, e);
  170 |     }
  171 |   }
  172 | }
  173 | 
  174 | /**
  175 |  * Check if backend is available
  176 |  */
  177 | export async function isBackendAvailable(request: APIRequestContext): Promise<boolean> {
  178 |   try {
  179 |     const response = await request.get(`${getBackendUrl()}/notes`, {
  180 |       timeout: 5000,
  181 |     });
  182 |     return response.status() < 500;
  183 |   } catch {
  184 |     return false;
  185 |   }
  186 | }
  187 | 
```