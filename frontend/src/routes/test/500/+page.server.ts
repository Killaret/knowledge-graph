import { error } from "@sveltejs/kit";
import type { PageServerLoad } from "./$types";

export const load: PageServerLoad = ({ url }) => {
  // Only trigger a 500 response when explicitly requested for regression testing.
  if (url.searchParams.get("trigger") === "500") {
    error(500, "Intentional test error");
  }

  return {};
};
