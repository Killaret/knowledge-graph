/**
 * LinkType — Value Object for a graph link type.
 *
 * Encapsulates the visual style (color, dash pattern), default weight, icon and
 * human-readable description of a link so that the renderer, forms, tooltips and
 * legend no longer rely on string literals or scattered switch statements.
 */

import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

export interface LinkTypeProps {
  type: string;
  label: string;
  icon: string;
  color: string;
  description: string;
  example: string;
  lineDash: number[];
  defaultWeight: number;
  isUi?: boolean;
  creatable?: boolean;
}

export class LinkType {
  constructor(public readonly props: Readonly<LinkTypeProps>) {}

  get type(): string {
    return this.props.type;
  }

  get label(): string {
    return formatMessage(this.props.label, getCurrentLocale());
  }

  get icon(): string {
    return this.props.icon;
  }

  get color(): string {
    return this.props.color;
  }

  get description(): string {
    return formatMessage(this.props.description, getCurrentLocale());
  }

  get example(): string {
    return formatMessage(this.props.example, getCurrentLocale());
  }

  get lineDash(): number[] {
    return this.props.lineDash;
  }

  get defaultWeight(): number {
    return this.props.defaultWeight;
  }

  get isUi(): boolean {
    return this.props.isUi ?? true;
  }

  get creatable(): boolean {
    return this.props.creatable ?? true;
  }

  /**
   * Returns an RGBA color for this link type, mixing the base color with the
   * given weight and fade opacity. Weight is expected to be in [0, 1].
   */
  getColor(weight: number, fadeOpacity: number = 1): string {
    const baseOpacity = 0.4 + (weight ?? 0.5) * 0.4;
    const finalOpacity = baseOpacity * fadeOpacity;

    const r = parseInt(this.color.slice(1, 3), 16);
    const g = parseInt(this.color.slice(3, 5), 16);
    const b = parseInt(this.color.slice(5, 7), 16);
    return `rgba(${r}, ${g}, ${b}, ${finalOpacity})`;
  }

  /**
   * Returns the canvas line dash pattern for this link type, optionally taking
   * the link weight into account for the `related` threshold.
   */
  getLineDash(weight?: number): number[] {
    if (this === LinkType.RELATED) {
      return (weight ?? 0.5) < 0.3 ? [6, 4] : [];
    }
    return [...this.lineDash];
  }

  /** Case-insensitive lookup. Falls back to RELATED. */
  static fromString(type: string | undefined): LinkType {
    const normalized = (type ?? "").toLowerCase().trim();
    return LinkType.MAP.get(normalized) ?? LinkType.RELATED;
  }

  static readonly REFERENCE = new LinkType({
    type: "reference",
    label: "linkType.reference",
    icon: "📖",
    color: "#3366ff",
    description: "linkType.reference.description",
    example: "linkType.reference.example",
    lineDash: [],
    defaultWeight: 0.8,
  });

  static readonly DEPENDENCY = new LinkType({
    type: "dependency",
    label: "linkType.dependency",
    icon: "🔗",
    color: "#ff6600",
    description: "linkType.dependency.description",
    example: "linkType.dependency.example",
    lineDash: [10, 3],
    defaultWeight: 0.7,
  });

  static readonly RELATED = new LinkType({
    type: "related",
    label: "linkType.related",
    icon: "🔀",
    color: "#999999",
    description: "linkType.related.description",
    example: "linkType.related.example",
    lineDash: [],
    defaultWeight: 0.5,
  });

  static readonly CUSTOM = new LinkType({
    type: "custom",
    label: "linkType.custom",
    icon: "✨",
    color: "#ff66ff",
    description: "linkType.custom.description",
    example: "linkType.custom.example",
    lineDash: [2, 6],
    defaultWeight: 0.5,
    isUi: false,
  });

  static readonly PARENT = new LinkType({
    type: "parent",
    label: "linkType.parent",
    icon: "⬆️",
    color: "#2dd4bf",
    description: "linkType.parent.description",
    example: "linkType.parent.example",
    lineDash: [],
    defaultWeight: 0.9,
  });

  static readonly CHILD = new LinkType({
    type: "child",
    label: "linkType.child",
    icon: "⬇️",
    color: "#f472b6",
    description: "linkType.child.description",
    example: "linkType.child.example",
    lineDash: [],
    defaultWeight: 0.9,
  });

  static readonly ALL_TYPES = [
    LinkType.REFERENCE,
    LinkType.DEPENDENCY,
    LinkType.RELATED,
    LinkType.CUSTOM,
    LinkType.PARENT,
    LinkType.CHILD,
  ] as const;

  private static readonly MAP = new Map(
    LinkType.ALL_TYPES.map((linkType) => [linkType.type, linkType])
  );

  static readonly UI_TYPES = LinkType.ALL_TYPES.filter((linkType) => linkType.isUi);
  static readonly CREATABLE_TYPES = LinkType.ALL_TYPES.filter((linkType) => linkType.creatable);
}
