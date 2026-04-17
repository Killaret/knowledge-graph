# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: performance\graph-3d-performance.spec.ts >> 3D Graph Performance @performance >> should maintain 30+ FPS with 50 nodes
- Location: tests\performance\graph-3d-performance.spec.ts:16:3

# Error details

```
Error: Failed to create link: 400 - {"error":"Key: 'createLinkRequest.SourceNoteID' Error:Field validation for 'SourceNoteID' failed on the 'required' tag\nKey: 'createLinkRequest.TargetNoteID' Error:Field validation for 'TargetNoteID' failed on the 'required' tag\nKey: 'createLinkRequest.LinkType' Error:Field validation for 'LinkType' failed on the 'required' tag"}
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
  78  |   // Go backend expects PascalCase field names
  79  |   const payload = {
  80  |     SourceNoteID: sourceId,
  81  |     TargetNoteID: targetId,
  82  |     Weight: weight,
  83  |     LinkType: linkType,
  84  |     Metadata: {},
  85  |   };
  86  | 
  87  |   const response = await request.post(`${getBackendUrl()}/links`, {
  88  |     data: payload,
  89  |   });
  90  | 
  91  |   if (!response.ok()) {
  92  |     const errorText = await response.text();
> 93  |     throw new Error(`Failed to create link: ${response.status()} - ${errorText}`);
      |           ^ Error: Failed to create link: 400 - {"error":"Key: 'createLinkRequest.SourceNoteID' Error:Field validation for 'SourceNoteID' failed on the 'required' tag\nKey: 'createLinkRequest.TargetNoteID' Error:Field validation for 'TargetNoteID' failed on the 'required' tag\nKey: 'createLinkRequest.LinkType' Error:Field validation for 'LinkType' failed on the 'required' tag"}
  94  |   }
  95  | 
  96  |   return await response.json();
  97  | }
  98  | 
  99  | /**
  100 |  * Create a star topology - center note with surrounding notes
  101 |  */
  102 | export async function createStarTopology(
  103 |   request: APIRequestContext,
  104 |   centerNote: Partial<NoteData>,
  105 |   surroundingNotes: Partial<NoteData>[],
  106 |   linkWeight = 0.7
  107 | ): Promise<{
  108 |   center: { id: string };
  109 |   surrounding: Array<{ id: string }>;
  110 | }> {
  111 |   const center = await createNote(request, centerNote);
  112 |   const surrounding = [];
  113 | 
  114 |   for (const noteData of surroundingNotes) {
  115 |     const note = await createNote(request, noteData);
  116 |     surrounding.push(note);
  117 |     await createLink(request, center.id, note.id, linkWeight);
  118 |   }
  119 | 
  120 |   return { center, surrounding };
  121 | }
  122 | 
  123 | /**
  124 |  * Create a chain topology - notes linked in sequence
  125 |  */
  126 | export async function createChainTopology(
  127 |   request: APIRequestContext,
  128 |   notes: Partial<NoteData>[],
  129 |   linkWeight = 0.8
  130 | ): Promise<Array<{ id: string }>> {
  131 |   const created = [];
  132 | 
  133 |   for (let i = 0; i < notes.length; i++) {
  134 |     const note = await createNote(request, notes[i]);
  135 |     created.push(note);
  136 | 
  137 |     // Link to previous note
  138 |     if (i > 0) {
  139 |       await createLink(request, created[i - 1].id, note.id, linkWeight);
  140 |     }
  141 |   }
  142 | 
  143 |   return created;
  144 | }
  145 | 
  146 | /**
  147 |  * Delete a note via API
  148 |  */
  149 | export async function deleteNote(request: APIRequestContext, noteId: string): Promise<void> {
  150 |   const response = await request.delete(`${getBackendUrl()}/notes/${noteId}`);
  151 | 
  152 |   if (!response.ok() && response.status() !== 404) {
  153 |     const errorText = await response.text();
  154 |     throw new Error(`Failed to delete note: ${response.status()} - ${errorText}`);
  155 |   }
  156 | }
  157 | 
  158 | /**
  159 |  * Clean up all test data
  160 |  */
  161 | export async function cleanupTestData(
  162 |   request: APIRequestContext,
  163 |   noteIds: string[]
  164 | ): Promise<void> {
  165 |   for (const id of noteIds) {
  166 |     try {
  167 |       await deleteNote(request, id);
  168 |     } catch (e) {
  169 |       // Ignore errors during cleanup
  170 |       console.warn(`Failed to delete note ${id}:`, e);
  171 |     }
  172 |   }
  173 | }
  174 | 
  175 | /**
  176 |  * Check if backend is available
  177 |  */
  178 | export async function isBackendAvailable(request: APIRequestContext): Promise<boolean> {
  179 |   try {
  180 |     const response = await request.get(`${getBackendUrl()}/notes`, {
  181 |       timeout: 5000,
  182 |     });
  183 |     return response.status() < 500;
  184 |   } catch {
  185 |     return false;
  186 |   }
  187 | }
  188 | 
```