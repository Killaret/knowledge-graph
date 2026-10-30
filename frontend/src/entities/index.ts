export * from "./achievement/model";
export * from "./graph/model";
export * from "./link/model";
export * from "./note/model";
export * from "./search/model";
export * from "./shared/model";
export * from "./user/model";

// graph-canvas is intentionally not re-exported here because its modules
// contain top-level code that depends on $shared/config; requiring it
// through a barrel would force that side effect when any $entities import
// is used. Import it explicitly via $entities/graph-canvas when needed.
