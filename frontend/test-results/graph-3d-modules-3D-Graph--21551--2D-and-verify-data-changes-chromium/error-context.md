# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: graph-3d-modules.spec.ts >> 3D Graph - Modular Architecture >> should toggle full graph mode in 2D and verify data changes
- Location: tests\graph-3d-modules.spec.ts:344:3

# Error details

```
Error: Failed to create link: 400 - {"error":"Key: 'createLinkRequest.SourceNoteID' Error:Field validation for 'SourceNoteID' failed on the 'required' tag\nKey: 'createLinkRequest.TargetNoteID' Error:Field validation for 'TargetNoteID' failed on the 'required' tag\nKey: 'createLinkRequest.LinkType' Error:Field validation for 'LinkType' failed on the 'required' tag"}
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
      - strong [ref=e60]: "2444"
      - text: links
```

# Test source

```ts
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
  78  |   // Debug logging
  79  |   console.log('[createLink] Creating link:', { sourceId, targetId, weight, linkType });
  80  |   
  81  |   // Validate inputs
  82  |   if (!sourceId || !targetId) {
  83  |     throw new Error(`Invalid parameters: sourceId=${sourceId}, targetId=${targetId}`);
  84  |   }
  85  |   
  86  |   // Go backend expects PascalCase field names
  87  |   const payload = {
  88  |     SourceNoteID: sourceId,
  89  |     TargetNoteID: targetId,
  90  |     Weight: weight,
  91  |     LinkType: linkType,
  92  |     Metadata: {},
  93  |   };
  94  |   
  95  |   console.log('[createLink] Payload:', JSON.stringify(payload));
  96  | 
  97  |   const response = await request.post(`${getBackendUrl()}/links`, {
  98  |     data: payload,
  99  |   });
  100 | 
  101 |   if (!response.ok()) {
  102 |     const errorText = await response.text();
> 103 |     throw new Error(`Failed to create link: ${response.status()} - ${errorText}`);
      |           ^ Error: Failed to create link: 400 - {"error":"Key: 'createLinkRequest.SourceNoteID' Error:Field validation for 'SourceNoteID' failed on the 'required' tag\nKey: 'createLinkRequest.TargetNoteID' Error:Field validation for 'TargetNoteID' failed on the 'required' tag\nKey: 'createLinkRequest.LinkType' Error:Field validation for 'LinkType' failed on the 'required' tag"}
  104 |   }
  105 | 
  106 |   return await response.json();
  107 | }
  108 | 
  109 | /**
  110 |  * Create a star topology - center note with surrounding notes
  111 |  */
  112 | export async function createStarTopology(
  113 |   request: APIRequestContext,
  114 |   centerNote: Partial<NoteData>,
  115 |   surroundingNotes: Partial<NoteData>[],
  116 |   linkWeight = 0.7
  117 | ): Promise<{
  118 |   center: { id: string };
  119 |   surrounding: Array<{ id: string }>;
  120 | }> {
  121 |   const center = await createNote(request, centerNote);
  122 |   const surrounding = [];
  123 | 
  124 |   for (const noteData of surroundingNotes) {
  125 |     const note = await createNote(request, noteData);
  126 |     surrounding.push(note);
  127 |     await createLink(request, center.id, note.id, linkWeight);
  128 |   }
  129 | 
  130 |   return { center, surrounding };
  131 | }
  132 | 
  133 | /**
  134 |  * Create a chain topology - notes linked in sequence
  135 |  */
  136 | export async function createChainTopology(
  137 |   request: APIRequestContext,
  138 |   notes: Partial<NoteData>[],
  139 |   linkWeight = 0.8
  140 | ): Promise<Array<{ id: string }>> {
  141 |   const created = [];
  142 | 
  143 |   for (let i = 0; i < notes.length; i++) {
  144 |     const note = await createNote(request, notes[i]);
  145 |     created.push(note);
  146 | 
  147 |     // Link to previous note
  148 |     if (i > 0) {
  149 |       await createLink(request, created[i - 1].id, note.id, linkWeight);
  150 |     }
  151 |   }
  152 | 
  153 |   return created;
  154 | }
  155 | 
  156 | /**
  157 |  * Delete a note via API
  158 |  */
  159 | export async function deleteNote(request: APIRequestContext, noteId: string): Promise<void> {
  160 |   const response = await request.delete(`${getBackendUrl()}/notes/${noteId}`);
  161 | 
  162 |   if (!response.ok() && response.status() !== 404) {
  163 |     const errorText = await response.text();
  164 |     throw new Error(`Failed to delete note: ${response.status()} - ${errorText}`);
  165 |   }
  166 | }
  167 | 
  168 | /**
  169 |  * Clean up all test data
  170 |  */
  171 | export async function cleanupTestData(
  172 |   request: APIRequestContext,
  173 |   noteIds: string[]
  174 | ): Promise<void> {
  175 |   for (const id of noteIds) {
  176 |     try {
  177 |       await deleteNote(request, id);
  178 |     } catch (e) {
  179 |       // Ignore errors during cleanup
  180 |       console.warn(`Failed to delete note ${id}:`, e);
  181 |     }
  182 |   }
  183 | }
  184 | 
  185 | /**
  186 |  * Check if backend is available
  187 |  */
  188 | export async function isBackendAvailable(request: APIRequestContext): Promise<boolean> {
  189 |   try {
  190 |     const response = await request.get(`${getBackendUrl()}/notes`, {
  191 |       timeout: 5000,
  192 |     });
  193 |     return response.status() < 500;
  194 |   } catch {
  195 |     return false;
  196 |   }
  197 | }
  198 | 
```