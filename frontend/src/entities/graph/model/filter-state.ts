import { SearchQuery } from "$entities/search/model";
import type { Note } from "$entities/note/model";
import type { GraphData } from "./graph-node";

export type FilterView = "graph" | "list";
export type SortBy = "created" | "updated" | "type";

export interface FilterStateProps {
  selectedType?: string;
  sortBy?: SortBy;
  searchQuery?: string;
  currentView?: FilterView;
}

/**
 * FilterState — Value Object that groups the list/graph filter state.
 *
 * Encapsulates type filtering, full-text search (via SearchQuery), and sorting
 * so the main page component no longer duplicates the same manual string
 * checks and switch statements.
 */
export class FilterState {
  readonly selectedType: string;
  readonly sortBy: SortBy;
  readonly searchQuery: SearchQuery;
  readonly currentView: FilterView;

  constructor(props: FilterStateProps = {}) {
    this.selectedType = props.selectedType ?? "all";
    this.sortBy = props.sortBy ?? "created";
    this.searchQuery = new SearchQuery(props.searchQuery ?? "");
    this.currentView = props.currentView ?? "graph";
  }

  get isTypeActive(): boolean {
    return this.selectedType !== "all";
  }

  get isSearchActive(): boolean {
    return !this.searchQuery.isEmpty();
  }

  get isInbox(): boolean {
    return this.selectedType === "inbox";
  }

  getSelectedTypeLabel(
    typeFilters: Array<{ id: string; label: string }>,
  ): string | undefined {
    return typeFilters.find((f) => f.id === this.selectedType)?.label;
  }

  matchesSearch(note: { title: string; content?: string | null }): boolean {
    if (!this.isSearchActive) return true;
    const q = this.searchQuery.value.toLowerCase();
    return (
      note.title.toLowerCase().includes(q) ||
      (note.content ?? "").toLowerCase().includes(q)
    );
  }

  matchesType(note: Note, getNoteType: (note: Note) => string): boolean {
    if (!this.isTypeActive) return true;
    if (this.isInbox) {
      const tags = note.metadata?.tags as string[] | undefined;
      return tags?.some((tag: string) => tag === "#inbox") ?? false;
    }
    return getNoteType(note) === this.selectedType;
  }

  filterNotes(notes: Note[], getNoteType: (note: Note) => string): Note[] {
    return notes.filter(
      (n) => this.matchesType(n, getNoteType) && this.matchesSearch(n),
    );
  }

  sortNotes(notes: Note[]): Note[] {
    const sorted = [...notes];
    switch (this.sortBy) {
      case "created":
        sorted.sort(
          (a, b) =>
            new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
        );
        break;
      case "updated":
        sorted.sort(
          (a, b) =>
            new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime(),
        );
        break;
      case "type":
        sorted.sort((a, b) =>
          (a.type || "unknown").localeCompare(b.type || "unknown"),
        );
        break;
    }
    return sorted;
  }

  applyFiltersAndSort(
    notes: Note[],
    getNoteType: (note: Note) => string,
  ): Note[] {
    return this.sortNotes(this.filterNotes(notes, getNoteType));
  }

  filterGraphData(
    graphData: GraphData,
    allNotes: Note[],
    getNoteType: (note: Note) => string,
  ): GraphData {
    if (!graphData.nodes.length) return graphData;

    const allowedIds = new Set(
      allNotes
        .filter(
          (n) => this.matchesType(n, getNoteType) && this.matchesSearch(n),
        )
        .map((n) => n.id),
    );

    const filteredNodes = graphData.nodes.filter((n) => allowedIds.has(n.id));
    const filteredNodeIds = new Set(filteredNodes.map((n) => n.id));
    const filteredLinks = graphData.links.filter(
      (l) => filteredNodeIds.has(l.source) && filteredNodeIds.has(l.target),
    );

    return { nodes: filteredNodes, links: filteredLinks };
  }

  with(props: Partial<FilterStateProps>): FilterState {
    return new FilterState({
      selectedType: props.selectedType ?? this.selectedType,
      sortBy: props.sortBy ?? this.sortBy,
      searchQuery: props.searchQuery ?? this.searchQuery.value,
      currentView: props.currentView ?? this.currentView,
    });
  }
}
