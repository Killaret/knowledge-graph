/**
 * SearchQuery — Value Object for a graph/note search string.
 *
 * Encapsulates trimming, validation, URL encoding, and equality checks
 * so components no longer duplicate manual string transformations.
 */

export interface SearchQueryProps {
  value: string;
  minLength?: number;
  maxLength?: number;
}

export class SearchQuery {
  readonly minLength: number;
  readonly maxLength: number;

  constructor(
    public readonly raw: string,
    options: { minLength?: number; maxLength?: number } = {}
  ) {
    this.minLength = options.minLength ?? 2;
    this.maxLength = options.maxLength ?? 200;
  }

  get value(): string {
    return this.raw.trim();
  }

  isEmpty(): boolean {
    return this.value === "";
  }

  isValid(): boolean {
    const length = this.value.length;
    return length >= this.minLength && length <= this.maxLength;
  }

  isTooLong(): boolean {
    return this.value.length > this.maxLength;
  }

  toURL(): string {
    return encodeURIComponent(this.value);
  }

  equals(other: SearchQuery): boolean {
    return this.value === other.value;
  }

  toString(): string {
    return this.value;
  }

  static fromURL(encoded: string | null | undefined): SearchQuery {
    try {
      return new SearchQuery(decodeURIComponent(encoded ?? ""));
    } catch {
      return new SearchQuery(encoded ?? "");
    }
  }

  static empty(): SearchQuery {
    return new SearchQuery("");
  }
}
